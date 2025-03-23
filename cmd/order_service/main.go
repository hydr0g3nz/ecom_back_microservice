package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	grpcctl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc"
	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	httpctl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/http"
	kafkactl "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/kafka"
	publisher "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/publisher"
	repo "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/config"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	pkglogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.product.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize logger
	logger := pkglogger.NewLogrusLogger()

	// Setup Cassandra connection
	cassandraSession, err := setupCassandra(cfg.Cassandra)
	if err != nil {
		logger.Fatal("Failed to connect to Cassandra", "error", err)
	}
	defer cassandraSession.Close()

	// Initialize repositories
	orderRepo := repo.NewCassandraOrderRepository(cassandraSession)
	orderReadRepo := repo.NewCassandraOrderReadRepository(cassandraSession)
	paymentRepo := repo.NewCassandraPaymentRepository(cassandraSession)
	shippingRepo := repo.NewCassandraShippingRepository(cassandraSession)
	orderEventRepo := repo.NewCassandraOrderEventRepository(cassandraSession)

	// Initialize event publisher
	kafkaTopics := map[string]string{
		"orders":   cfg.Kafka.Topics.Orders,
		"payments": cfg.Kafka.Topics.Payments,
		"shipping": cfg.Kafka.Topics.Shipping,
	}
	eventPublisher := publisher.NewKafkaOrderEventPublisher(cfg.Kafka.Brokers, kafkaTopics)

	// Initialize use cases
	orderUsecase := usecase.NewOrderUsecase(orderRepo, orderReadRepo, orderEventRepo, eventPublisher)
	paymentUsecase := usecase.NewPaymentUsecase(orderRepo, paymentRepo, orderEventRepo, eventPublisher)
	shippingUsecase := usecase.NewShippingUsecase(orderRepo, shippingRepo, orderEventRepo, eventPublisher)
	orderHistoryUsecase := usecase.NewOrderHistoryUsecase(orderEventRepo)

	// Start Kafka consumer
	kafkaConsumer := kafkactl.NewKafkaConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.ConsumerGroup,
		orderUsecase,
		paymentUsecase,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming Kafka messages
	go kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Start HTTP server
	app := setupHTTPServer(cfg, logger, orderUsecase, paymentUsecase, shippingUsecase, orderHistoryUsecase)
	go startHTTPServer(app, cfg.Server.Address)

	// Start gRPC server
	grpcServer := setupGRPCServer(logger, orderUsecase, paymentUsecase, shippingUsecase, orderHistoryUsecase)
	go startGRPCServer(grpcServer, fmt.Sprintf(":%s", cfg.GRPC.Port))

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	logger.Info("Shutting down server...")

	// Cancel context to stop Kafka consumers
	cancel()

	// Shut down HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Shut down gRPC server
	grpcServer.GracefulStop()

	// Close event publisher
	if err := eventPublisher.Close(); err != nil {
		logger.Error("Failed to close event publisher", "error", err)
	}

	logger.Info("Server exiting")
}

func setupCassandra(cfg config.CassandraConfig) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = cfg.Timeout
	cluster.ConnectTimeout = cfg.ConnectTimeout

	if cfg.Username != "" && cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	return cluster.CreateSession()
}

func setupHTTPServer(
	cfg *config.Config,
	logger pkglogger.Logger,
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
	shippingUsecase usecase.ShippingUsecase,
	orderHistoryUsecase usecase.OrderHistoryUsecase,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	// Middlewares
	app.Use(fb_logger.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, err interface{}) {
			logger.Error("Recovered from panic", "error", err, "stack", string(debug.Stack()))
			c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		},
	}))

	// Routes
	api := app.Group("/api")

	// Order handlers
	orderHandler := httpctl.NewOrderHandler(orderUsecase, paymentUsecase, shippingUsecase, orderHistoryUsecase, logger)
	orderHandler.RegisterRoutes(api)

	return app
}

func startHTTPServer(app *fiber.App, address string) {
	if err := app.Listen(address); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func setupGRPCServer(
	logger pkglogger.Logger,
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
	shippingUsecase usecase.ShippingUsecase,
	orderHistoryUsecase usecase.OrderHistoryUsecase,
) *grpc.Server {
	server := grpc.NewServer()

	// Register order service
	orderServer := grpcctl.NewOrderServer(orderUsecase, paymentUsecase, shippingUsecase, orderHistoryUsecase, logger)
	pb.RegisterOrderServiceServer(server, orderServer)

	// Enable reflection for tools like grpcurl
	reflection.Register(server)

	return server
}

func startGRPCServer(server *grpc.Server, address string) {
	// Start the gRPC server
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
