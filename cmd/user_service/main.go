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
	"time"

	"github.com/gofiber/fiber/v2"
	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	// Update these imports to match your project structure
	grpcctl "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/grpc"
	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/grpc/proto"
	httpctl "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/http"

	gormrepo "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/repository/gorm"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/repository/gorm/model"
	appconfig "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/config"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/jwt_service"
	applogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// Repositories holds all repository implementations
type Repositories struct {
	UserRepository  repository.UserRepository
	TokenRepository repository.TokenRepository
}

// Usecases holds all usecase implementations
type Usecases struct {
	UserUsecase  usecase.UserUsecase
	TokenUsecase usecase.TokenUsecase
	AuthUsecase  usecase.AuthUsecase
}

// Controllers holds all controllers
type Controllers struct {
	HTTP *httpctl.UserHandler
	GRPC *grpcctl.UserServer
}

// Servers holds all server instances
type Servers struct {
	HTTP *fiber.App
	GRPC *grpc.Server
}

// GormLogAdapter adapts application logger to GORM's logger interface
type GormLogAdapter struct {
	log applogger.Logger
}

func (l *GormLogAdapter) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log.Info(message)
}

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Initialize context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	log := applogger.NewZapLogger()
	log.Info("Starting user service")

	// Load configuration
	config, err := appconfig.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize database
	db, err := initDatabase(config.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}

	// Initialize repositories
	repositories := initRepositories(db)

	// Initialize usecases
	usecases := initUsecases(repositories, config)

	// Initialize controllers
	controllers := initControllers(usecases, log)

	// Start servers
	servers := initServers(config, controllers, log)

	// Handle graceful shutdown
	handleGracefulShutdown(ctx, cancel, servers, log)
}

// initDatabase initializes the database connection
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
	if err := db.AutoMigrate(&model.Token{}, &model.User{}); err != nil {
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
		UserRepository:  gormrepo.NewGormUserRepository(db),
		TokenRepository: gormrepo.NewGormTokenRepository(db),
	}
}

// initUsecases initializes all usecases
func initUsecases(repos *Repositories, config *appconfig.Config) *Usecases {
	jwtSvc := jwt_service.NewJWTService(config.JWT)

	userUsecase := usecase.NewUserUsecase(repos.UserRepository)
	tokenUsecase := usecase.NewTokenUsecase(repos.TokenRepository, jwtSvc)
	authUsecase := usecase.NewAuthUsecase(userUsecase, tokenUsecase)

	return &Usecases{
		UserUsecase:  userUsecase,
		TokenUsecase: tokenUsecase,
		AuthUsecase:  authUsecase,
	}
}

// initControllers initializes all controllers
func initControllers(usecases *Usecases, log applogger.Logger) *Controllers {
	return &Controllers{
		HTTP: httpctl.NewUserHandler(usecases.AuthUsecase, usecases.UserUsecase, log),
		GRPC: grpcctl.NewUserServer(usecases.AuthUsecase, usecases.UserUsecase, log),
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
func initHTTPServer(config appconfig.ServerConfig, handler *httpctl.UserHandler, log applogger.Logger) *fiber.App {
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
func initGRPCServer(config appconfig.GRPCConfig, server *grpcctl.UserServer, log applogger.Logger) *grpc.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", config.Port))
	if err != nil {
		log.Fatal("Failed to listen for gRPC", "error", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, server)

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
