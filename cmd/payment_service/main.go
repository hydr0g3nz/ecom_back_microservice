// cmd/inventory_service/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	// Update these imports to match your project structure

	httpctl "github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/controller/http"
	eventSvc "github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/event"
	messaging "github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/event"
	gormrepo "github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/repository/gorm"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/repository/gorm/model"
	appconfig "github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/config"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/usecase"
	applogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// Repositories holds all repository implementations
type Repositories struct {
	InventoryRepository repository.InventoryRepository
}

// Services holds all service implementations
type Services struct {
	EventPublisher  service.EventPublisherService
	EventSubscriber service.EventSubscriberService
}

// Usecases holds all usecase implementations
type Usecases struct {
	InventoryUsecase   usecase.InventoryUsecase
	ReservationUsecase usecase.ReservationProcessorUsecase
}

// Controllers holds all controllers
type Controllers struct {
	HTTP *httpctl.InventoryHandler
}

type GormLogAdapter struct {
	log applogger.Logger
}

func (l *GormLogAdapter) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log.Info(message)
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
	db, err := initDatabase(config.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}

	// Initialize repositories
	repositories := initRepositories(db)
	eventConfig := &messaging.KafkaConfig{
		Brokers:         config.Messaging.Brokers,
		InventoryTopic:  config.Messaging.InventoryTopic,
		OrderTopic:      config.Messaging.OrderTopic,
		ConsumerGroupID: config.Messaging.ConsumerGroupID,
	}
	eventServicePublisher, err := eventSvc.NewKafkaEventPublisher(eventConfig)
	if err != nil {
		log.Fatal("Failed to initialize Kafka event publisher", "error", err)
	}
	// Initialize usecases
	// usecases := initUsecases(repositories, nil) // We'll set event service after initializing usecases
	usecases := initUsecases(repositories, eventServicePublisher)

	// Initialize Kafka consumer (needs usecase)
	kafkaConsumer, err := eventSvc.NewKafkaEventSubscriber(eventConfig, usecases.ReservationUsecase)
	if err != nil {
		log.Fatal("Failed to initialize Kafka consumer", "error", err)
	}

	// Initialize event service

	// eventServiceSubscriber := eventSvc.NewKafkaEventSubscriberService(kafkaConsumer, log)

	// Update usecases with event service

	// Start Kafka consumer
	if err := kafkaConsumer.SubscribeToOrderEvents(ctx); err != nil {
		log.Fatal("Failed to start Kafka consumer", "error", err)
	}
	defer func() {
		if err := eventServicePublisher.Close(); err != nil {
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
func initDatabase(config appconfig.DatabaseConfig, log applogger.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.Name)

	gormLogger := gormlogger.New(
		&GormLogAdapter{log: log},
		gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}
	log.Info("Connected to database")

	// Auto migrate models
	if err := db.AutoMigrate(&model.InventoryItem{}, &model.InventoryReservation{}, &model.StockTransaction{}); err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetConnMaxLifetime(config.MaxLife)

	return db, nil
}

// initRepositories initializes all repositories
func initRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		InventoryRepository: gormrepo.NewGormInventoryRepository(db),
	}
}

// initUsecases initializes all usecases
func initUsecases(repos *Repositories, eventService service.EventPublisherService) *Usecases {
	inventoryUsecase := usecase.NewInventoryUsecase(repos.InventoryRepository, eventService)
	return &Usecases{
		InventoryUsecase:   inventoryUsecase,
		ReservationUsecase: usecase.NewReservationProcessorUsecase(repos.InventoryRepository, eventService, inventoryUsecase),
	}
}

// initControllers initializes all controllers
func initControllers(usecases *Usecases, log applogger.Logger) *Controllers {
	return &Controllers{
		HTTP: httpctl.NewInventoryHandler(usecases.InventoryUsecase, log),
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

	return &Servers{
		HTTP: httpServer,
	}
}

// initHTTPServer initializes the HTTP server
func initHTTPServer(config appconfig.ServerConfig, handler *httpctl.InventoryHandler, log applogger.Logger) *fiber.App {
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
	// servers.GRPC.GracefulStop()

	cancel()
	log.Info("Shutdown complete")
}
