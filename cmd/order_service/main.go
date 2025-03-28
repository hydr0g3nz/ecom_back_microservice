// cmd/order_service/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gofiber/fiber/v2"
	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	// Update these imports to match your project structure
	grpcctl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc"
	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	httpctl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/http"
	eventSvc "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event/consumer"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event/producer"
	mongorepo "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/mongo"
	appconfig "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/config"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	applogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// Repositories holds all repository implementations
type Repositories struct {
	OrderRepository repository.OrderRepository
}

// Services holds all service implementations
type Services struct {
	EventService service.EventService
}

// Usecases holds all usecase implementations
type Usecases struct {
	OrderUsecase usecase.OrderUsecase
}

// Controllers holds all controllers
type Controllers struct {
	HTTP *httpctl.OrderHandler
	GRPC *grpcctl.OrderServer
}

// Servers holds all server instances
type Servers struct {
	HTTP *fiber.App
	GRPC *grpc.Server
}

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.order.yaml", "path to config file")
	flag.Parse()

	// Initialize context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	log := applogger.NewZapLogger()
	log.Info("Starting order service")

	// Load configuration
	config, err := appconfig.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize MongoDB database
	db, err := initMongoDB(ctx, config.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB", "error", err)
	}
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			log.Error("Failed to disconnect MongoDB client", "error", err)
		}
	}()

	// Initialize repositories
	repositories := initRepositories(db)

	// Initialize event producers and consumers
	kafkaProducer, err := producer.NewKafkaProducer(config.Kafka.Brokers, log)
	if err != nil {
		log.Fatal("Failed to initialize Kafka producer", "error", err)
	}

	// Initialize usecases
	usecases := initUsecases(repositories, nil) // We'll set event service after initializing usecases

	// Initialize Kafka consumer (needs usecase)
	kafkaConsumer, err := consumer.NewKafkaConsumer(
		config.Kafka.Brokers,
		config.Kafka.GroupID,
		usecases.OrderUsecase,
		log,
	)
	if err != nil {
		log.Fatal("Failed to initialize Kafka consumer", "error", err)
	}

	// Initialize event service
	eventService := eventSvc.NewKafkaEventService(kafkaProducer, kafkaConsumer, log)

	// Update usecases with event service
	usecases = initUsecases(repositories, eventService)

	// Start Kafka consumer
	if err := kafkaConsumer.Start(ctx); err != nil {
		log.Fatal("Failed to start Kafka consumer", "error", err)
	}
	defer func() {
		if err := eventService.Close(); err != nil {
			log.Error("Failed to close event service", "error", err)
		}
	}()

	// Initialize controllers
	controllers := initControllers(usecases, log)

	// Start servers
	servers := initServers(config, controllers, log)

	// Handle graceful shutdown
	handleGracefulShutdown(ctx, cancel, servers, log)
}

// initMongoDB initializes the MongoDB connection
func initMongoDB(ctx context.Context, config appconfig.DatabaseConfig, log applogger.Logger) (*mongo.Database, error) {
	// Set client options
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(config.PoolSize).
		SetConnectTimeout(config.ConnTimeout)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Info("Connected to MongoDB", "uri", config.URI, "database", config.Name)

	// Get database
	db := client.Database(config.Name)
	return db, nil
}

// initRepositories initializes all repositories
func initRepositories(db *mongo.Database) *Repositories {
	return &Repositories{
		OrderRepository: mongorepo.NewMongoOrderRepository(db),
	}
}

// initUsecases initializes all usecases
func initUsecases(repos *Repositories, eventService service.EventService) *Usecases {
	return &Usecases{
		OrderUsecase: usecase.NewOrderUsecase(repos.OrderRepository, eventService),
	}
}

// initControllers initializes all controllers
func initControllers(usecases *Usecases, log applogger.Logger) *Controllers {
	return &Controllers{
		HTTP: httpctl.NewOrderHandler(usecases.OrderUsecase, log),
		GRPC: grpcctl.NewOrderServer(usecases.OrderUsecase, log),
	}
}

// initServers initializes and starts all servers
func initServers(config *appconfig.Config, controllers *Controllers, log applogger.Logger) *Servers {
	// Initialize HTTP server
	httpServer := initHTTPServer(config.Server, controllers.HTTP, log)

	// Start HTTP server
	go func() {
		log.Info("Starting Fiber server", "addr", config.Server.Address)
		if err := httpServer.Listen(config.Server.Address); err != nil {
			log.Fatal("Server failed to start", "error", err)
		}
	}()

	// Initialize and start gRPC server
	grpcServer := initGRPCServer(config.GRPC, controllers.GRPC, log)

	return &Servers{
		HTTP: httpServer,
		GRPC: grpcServer,
	}
}

// initHTTPServer initializes the HTTP server
func initHTTPServer(config appconfig.ServerConfig, handler *httpctl.OrderHandler, log applogger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			log.Error("HTTP error", "status", code, "error", err.Error())
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add middlewares
	app.Use(fb_logger.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, err interface{}) {
			log.Error("Recovered from panic", "error", err, "stack", string(debug.Stack()))
			c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		},
	}))

	// Register routes
	api := app.Group("/api")
	handler.RegisterRoutes(api)

	return app
}

// initGRPCServer initializes and starts the gRPC server
func initGRPCServer(config appconfig.GRPCConfig, server *grpcctl.OrderServer, log applogger.Logger) *grpc.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", config.Port))
	if err != nil {
		log.Fatal("Failed to listen for gRPC", "error", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, server)

	log.Info("Starting gRPC server", "port", config.Port)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC", "error", err)
		}
	}()

	return s
}

// handleGracefulShutdown configures graceful shutdown for all servers
func handleGracefulShutdown(ctx context.Context, cancel context.CancelFunc, servers *Servers, log applogger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	// Shutdown HTTP server
	if err := servers.HTTP.Shutdown(); err != nil {
		log.Error("Error during HTTP server shutdown", "error", err)
	}

	// Shutdown gRPC server
	servers.GRPC.GracefulStop()

	cancel()
	log.Info("Shutdown complete")
}
