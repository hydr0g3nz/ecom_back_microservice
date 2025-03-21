package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/grpc"

	// Update these imports to match your project structure
	grpcctl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc"
	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	httpctl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/http"
	kafkactl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/kafka"
	cassandraRepo "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra"

	appconfig "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/config"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/handler"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/publisher"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/query"
	applogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// Repositories holds all repository implementations
type Repositories struct {
	OrderRepository      repository.OrderRepository
	OrderEventRepository repository.OrderEventRepository
	OrderReadRepository  repository.OrderReadRepository
	PaymentRepository    repository.PaymentRepository
	ShippingRepository   repository.ShippingRepository
}

// CommandUsecases holds all command usecase implementations
type CommandUsecases struct {
	CreateOrderUsecase    command.CreateOrderUsecase
	UpdateOrderUsecase    command.UpdateOrderUsecase
	CancelOrderUsecase    command.CancelOrderUsecase
	ProcessPaymentUsecase command.ProcessPaymentUsecase
	UpdateShippingUsecase command.UpdateShippingUsecase
}

// QueryUsecases holds all query usecase implementations
type QueryUsecases struct {
	GetOrderUsecase     query.GetOrderUsecase
	ListOrdersUsecase   query.ListOrdersUsecase
	OrderHistoryUsecase query.OrderHistoryUsecase
}

// EventHandlers holds all event handlers
type EventHandlers struct {
	InventoryHandler handler.InventoryEventHandler
	PaymentHandler   handler.PaymentEventHandler
}

// Controllers holds all controllers
type Controllers struct {
	HTTP  *httpctl.OrderHandler
	GRPC  *grpcctl.OrderServer
	Kafka *kafkactl.KafkaConsumer
}

// Servers holds all server instances
type Servers struct {
	HTTP *fiber.App
	GRPC *grpc.Server
}

// App represents the application
type App struct {
	config          *appconfig.Config
	cassSession     *gocql.Session
	kafkaProducer   *kafkactl.KafkaProducer
	repositories    *Repositories
	commandUsecases *CommandUsecases
	queryUsecases   *QueryUsecases
	eventHandlers   *EventHandlers
	eventPublisher  publisher.OrderEventPublisher
	controllers     *Controllers
	servers         *Servers
	logger          applogger.Logger
}

// NewApp creates a new application instance
func NewApp(config *appconfig.Config, logger applogger.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the application
func (app *App) Initialize(ctx context.Context) error {
	// Initialize Cassandra session
	cassSession, err := initCassandraSession(app.config.Cassandra)
	if err != nil {
		return fmt.Errorf("failed to initialize Cassandra session: %w", err)
	}
	app.cassSession = cassSession
	app.logger.Info("Connected to Cassandra")

	// Initialize Kafka producer
	kafkaProducer := kafkactl.NewKafkaProducer(app.config.Kafka.Brokers)
	app.kafkaProducer = kafkaProducer
	app.logger.Info("Connected to Kafka")

	// Create event publisher
	eventPublisher := kafkaProducer.CreateOrderEventPublisher()
	app.eventPublisher = eventPublisher

	// Initialize repositories
	repositories, err := app.initRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}
	app.repositories = repositories

	// Initialize command usecases
	commandUsecases := app.initCommandUsecases()
	app.commandUsecases = commandUsecases

	// Initialize query usecases
	queryUsecases := app.initQueryUsecases()
	app.queryUsecases = queryUsecases

	// Initialize event handlers
	eventHandlers := app.initEventHandlers()
	app.eventHandlers = eventHandlers

	// Initialize controllers
	controllers := app.initControllers()
	app.controllers = controllers

	// Initialize servers
	servers := app.initServers()
	app.servers = servers

	return nil
}

// Start starts the application
func (app *App) Start(ctx context.Context) error {
	// Start HTTP server
	go func() {
		app.logger.Info("Starting HTTP server", "address", app.config.Server.Address)
		if err := app.servers.HTTP.Listen(app.config.Server.Address); err != nil {
			app.logger.Fatal("Failed to start HTTP server", "error", err)
		}
	}()

	// Start gRPC server
	go func() {
		app.logger.Info("Starting gRPC server", "port", app.config.GRPC.Port)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", app.config.GRPC.Port))
		if err != nil {
			app.logger.Fatal("Failed to listen for gRPC", "error", err)
		}
		if err := app.servers.GRPC.Serve(lis); err != nil {
			app.logger.Fatal("Failed to serve gRPC", "error", err)
		}
	}()

	// Start Kafka consumer
	app.controllers.Kafka.Start(ctx)
	app.logger.Info("Started Kafka consumer")

	return nil
}

// Shutdown gracefully shuts down the application
func (app *App) Shutdown(ctx context.Context) error {
	// Create a WaitGroup to wait for all servers to shut down
	var wg sync.WaitGroup
	wg.Add(3) // HTTP, gRPC, Kafka

	// Shutdown HTTP server
	go func() {
		defer wg.Done()
		if err := app.servers.HTTP.Shutdown(); err != nil {
			app.logger.Error("Error during HTTP server shutdown", "error", err)
		}
		app.logger.Info("HTTP server shutdown complete")
	}()

	// Shutdown gRPC server
	go func() {
		defer wg.Done()
		app.servers.GRPC.GracefulStop()
		app.logger.Info("gRPC server shutdown complete")
	}()

	// Close Kafka consumer
	go func() {
		defer wg.Done()
		if err := app.controllers.Kafka.Close(); err != nil {
			app.logger.Error("Error during Kafka consumer shutdown", "error", err)
		}

		// Close Kafka producer
		if err := app.kafkaProducer.Close(); err != nil {
			app.logger.Error("Error during Kafka producer shutdown", "error", err)
		}
		app.logger.Info("Kafka connections closed")
	}()

	// Wait for all servers to shut down
	wg.Wait()

	// Close Cassandra session
	app.cassSession.Close()
	app.logger.Info("Cassandra session closed")

	return nil
}

// initCassandraSession initializes a connection to Cassandra
func initCassandraSession(config appconfig.CassandraConfig) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = config.Timeout
	cluster.ConnectTimeout = config.ConnectTimeout

	if config.Username != "" && config.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: config.Username,
			Password: config.Password,
		}
	}

	return cluster.CreateSession()
}

// initRepositories initializes all repositories
func (app *App) initRepositories(ctx context.Context) (*Repositories, error) {
	// Initialize Cassandra repositories
	orderRepo := cassandraRepo.NewCassandraOrderRepository(app.cassSession)
	orderEventRepo := cassandraRepo.NewCassandraOrderEventRepository(app.cassSession)
	paymentRepo := cassandraRepo.NewCassandraPaymentRepository(app.cassSession)
	shippingRepo := cassandraRepo.NewCassandraShippingRepository(app.cassSession)

	// Initialize read models
	orderReadRepo := cassandraRepo.NewCassandraOrderReadRepository(app.cassSession)

	return &Repositories{
		OrderRepository:      orderRepo,
		OrderEventRepository: orderEventRepo,
		OrderReadRepository:  orderReadRepo,
		PaymentRepository:    paymentRepo,
		ShippingRepository:   shippingRepo,
	}, nil
}

// initCommandUsecases initializes all command usecases
func (app *App) initCommandUsecases() *CommandUsecases {
	createOrderUsecase := command.NewCreateOrderUsecase(
		app.repositories.OrderRepository,
		app.repositories.OrderEventRepository,
		app.eventPublisher,
	)

	updateOrderUsecase := command.NewUpdateOrderUsecase(
		app.repositories.OrderRepository,
		app.repositories.OrderEventRepository,
		app.eventPublisher,
	)

	cancelOrderUsecase := command.NewCancelOrderUsecase(
		app.repositories.OrderRepository,
		app.repositories.OrderEventRepository,
		app.eventPublisher,
	)

	processPaymentUsecase := command.NewProcessPaymentUsecase(
		app.repositories.OrderRepository,
		app.repositories.PaymentRepository,
		app.repositories.OrderEventRepository,
		app.eventPublisher,
	)

	updateShippingUsecase := command.NewUpdateShippingUsecase(
		app.repositories.OrderRepository,
		app.repositories.ShippingRepository,
		app.repositories.OrderEventRepository,
		app.eventPublisher,
	)

	return &CommandUsecases{
		CreateOrderUsecase:    createOrderUsecase,
		UpdateOrderUsecase:    updateOrderUsecase,
		CancelOrderUsecase:    cancelOrderUsecase,
		ProcessPaymentUsecase: processPaymentUsecase,
		UpdateShippingUsecase: updateShippingUsecase,
	}
}

// initQueryUsecases initializes all query usecases
func (app *App) initQueryUsecases() *QueryUsecases {
	getOrderUsecase := query.NewGetOrderUsecase(
		app.repositories.OrderReadRepository,
	)

	listOrdersUsecase := query.NewListOrdersUsecase(
		app.repositories.OrderReadRepository,
	)

	orderHistoryUsecase := query.NewOrderHistoryUsecase(
		app.repositories.OrderEventRepository,
	)

	return &QueryUsecases{
		GetOrderUsecase:     getOrderUsecase,
		ListOrdersUsecase:   listOrdersUsecase,
		OrderHistoryUsecase: orderHistoryUsecase,
	}
}

// initEventHandlers initializes all event handlers
func (app *App) initEventHandlers() *EventHandlers {
	inventoryHandler := handler.NewInventoryEventHandler(
		app.repositories.OrderRepository,
		app.repositories.OrderEventRepository,
		app.commandUsecases.CancelOrderUsecase,
	)

	paymentHandler := handler.NewPaymentEventHandler(
		app.repositories.OrderRepository,
		app.repositories.PaymentRepository,
		app.repositories.OrderEventRepository,
		app.commandUsecases.CancelOrderUsecase,
		app.commandUsecases.ProcessPaymentUsecase,
	)

	return &EventHandlers{
		InventoryHandler: *inventoryHandler,
		PaymentHandler:   *paymentHandler,
	}
}

// initControllers initializes all controllers
func (app *App) initControllers() *Controllers {
	// HTTP controller
	httpHandler := httpctl.NewOrderHandler(
		app.commandUsecases.CreateOrderUsecase,
		app.commandUsecases.UpdateOrderUsecase,
		app.commandUsecases.CancelOrderUsecase,
		app.commandUsecases.ProcessPaymentUsecase,
		app.commandUsecases.UpdateShippingUsecase,
		app.queryUsecases.GetOrderUsecase,
		app.queryUsecases.ListOrdersUsecase,
		app.queryUsecases.OrderHistoryUsecase,
		app.repositories.ShippingRepository,
		app.logger,
	)

	// gRPC controller
	grpcServer := grpcctl.NewOrderServer(
		app.commandUsecases.CreateOrderUsecase,
		app.commandUsecases.UpdateOrderUsecase,
		app.commandUsecases.CancelOrderUsecase,
		app.commandUsecases.ProcessPaymentUsecase,
		app.commandUsecases.UpdateShippingUsecase,
		app.queryUsecases.GetOrderUsecase,
		app.queryUsecases.ListOrdersUsecase,
		app.queryUsecases.OrderHistoryUsecase,
		app.logger,
	)

	// Kafka consumer
	kafkaConsumer := kafkactl.NewKafkaConsumer(
		app.config.Kafka.Brokers,
		app.commandUsecases.CancelOrderUsecase,
		app.commandUsecases.ProcessPaymentUsecase,
		app.commandUsecases.UpdateShippingUsecase,
	)

	return &Controllers{
		HTTP:  httpHandler,
		GRPC:  grpcServer,
		Kafka: kafkaConsumer,
	}
}

// initServers initializes all servers
func (app *App) initServers() *Servers {
	// Initialize HTTP server
	httpServer := fiber.New(fiber.Config{
		ReadTimeout:  app.config.Server.ReadTimeout,
		WriteTimeout: app.config.Server.WriteTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			app.logger.Error("HTTP error", "status", code, "error", err.Error())
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add middlewares
	httpServer.Use(fb_logger.New())
	httpServer.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, err interface{}) {
			app.logger.Error("Recovered from panic", "error", err, "stack", string(debug.Stack()))
			c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		},
	}))

	// Register routes
	api := httpServer.Group("/api")
	app.controllers.HTTP.RegisterRoutes(api)

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, app.controllers.GRPC)

	return &Servers{
		HTTP: httpServer,
		GRPC: grpcServer,
	}
}

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.order.yaml", "path to config file")
	flag.Parse()

	// Initialize logger with logrus implementation instead of zap
	logger := applogger.NewLogrusLogger()
	logger.Info("Starting order service")

	// Load configuration
	config, err := appconfig.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application
	app := NewApp(config, logger)
	if err := app.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize application", "error", err)
	}

	// Start application
	if err := app.Start(ctx); err != nil {
		logger.Fatal("Failed to start application", "error", err)
	}

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown application
	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
	}

	logger.Info("Shutdown complete")
}
