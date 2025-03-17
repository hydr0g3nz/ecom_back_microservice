package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Update these imports to match your project structure
	handler "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller"
	grpcServer "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/grpc"
	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/grpc/proto"
	repository "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/repository/gorm"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/repository/gorm/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/jwt_service"
	applogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

func main() {
	// Initialize logger
	log := applogger.NewZapLogger()
	log.Info("Starting user service")

	// Load configuration from environment variables
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "pass")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3366")
	dbName := getEnv("DB_NAME", "ecom_user_service")
	grpcPort := getEnv("GRPC_PORT", "50051")

	// Connect to database using GORM
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Configure GORM logger
	gormLogger := logger.New(
		&GormLogAdapter{log: log}, // Custom adapter to use your application logger
		logger.Config{
			SlowThreshold:             time.Second, // Threshold for slow SQL queries
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error
			Colorful:                  false,       // Disable color
		},
	)

	// Open connection to the database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	db.AutoMigrate(&model.Token{}, &model.User{})
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	log.Info("Connected to database")

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database connection", "error", err)
	}

	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Initialize dependencies
	userRepo := repository.NewGormUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	tokenRepo := repository.NewGormTokenRepository(db)
	jwtSvc := jwt_service.NewJWTService(jwt_service.Config{
		SecretKey:            getEnv("JWT_SECRET_KEY", "secret_key"),
		Issuer:               "user_service",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	})
	tokenUsecase := usecase.NewTokenUsecase(tokenRepo, jwtSvc)
	authUserUsecase := usecase.NewAuthUsecase(userUsecase, tokenUsecase)

	// Start gRPC server in a goroutine
	go startGRPCServer(grpcPort, authUserUsecase, userUsecase, log)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			// Check for known error types
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

	// Initialize HTTP handler and middleware
	userHandler := handler.NewUserHandler(authUserUsecase, userUsecase, log)

	// Register routes
	api := app.Group("/api")
	userHandler.RegisterRoutes(api)

	// Start server in a goroutine
	serverAddr := getEnv("SERVER_ADDR", "127.0.0.1:8080")
	go func() {
		log.Info("Starting Fiber server", "addr", serverAddr)
		if err := app.Listen(serverAddr); err != nil {
			log.Fatal("Server failed to start", "error", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Error("Error during server shutdown", "error", err)
	}
}

func startGRPCServer(port string, authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase, log applogger.Logger) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Failed to listen for gRPC", "error", err)
	}

	s := grpc.NewServer()
	userServer := grpcServer.NewUserServer(authUsecase, userUsecase, log)
	pb.RegisterUserServiceServer(s, userServer)

	log.Info("Starting gRPC server", "port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC", "error", err)
	}
}

// GormLogAdapter adapts your application logger to GORM's logger interface
type GormLogAdapter struct {
	log applogger.Logger
}

func (l *GormLogAdapter) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log.Info(message)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
