package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fb_logger "github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Update these imports to match your project structure
	handler "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller"
	repository "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/repository/gorm"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
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

	// accessSecret := getEnv("JWT_ACCESS_SECRET", "access_secret_key")
	// refreshSecret := getEnv("JWT_REFRESH_SECRET", "refresh_secret_key")

	// accessTokenExpiryStr := getEnv("JWT_ACCESS_EXPIRY", "15m")
	// refreshTokenExpiryStr := getEnv("JWT_REFRESH_EXPIRY", "7d")

	// // Parse durations
	// accessTokenExpiry, err := time.ParseDuration(accessTokenExpiryStr)
	// if err != nil {
	// 	log.Fatal("Invalid access token expiry", "error", err)
	// }

	// refreshTokenExpiry, err := time.ParseDuration(refreshTokenExpiryStr)
	// if err != nil {
	// 	log.Fatal("Invalid refresh token expiry", "error", err)
	// }

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
	db.AutoMigrate(&entity.User{}, &entity.Token{})
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
	// v := validator.NewValidator()
	// hashService := utils.NewBcryptService()
	// tokenSvc := token.NewJWTService(accessSecret, refreshSecret, accessTokenExpiry, refreshTokenExpiry, log)
	userRepo := repository.NewGormUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	tokenRepo := repository.NewGormTokenRepository(db)
	jwtSvc := jwt_service.NewJWTService(jwt_service.Config{SecretKey: "secret_key", Issuer: "user_service", AccessTokenDuration: 15 * time.Minute, RefreshTokenDuration: 7 * 24 * time.Hour})
	tokenUsecase := usecase.NewTokenUsecase(tokenRepo, jwtSvc)
	authUserUsecase := usecase.NewAuthUsecase(userUsecase, tokenUsecase)
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
	// app.Use(recover.New())
	app.Use(fb_logger.New())

	// Initialize HTTP handler and middleware
	userHandler := handler.NewUserHandler(authUserUsecase, userUsecase, log)
	// authMiddleware := http.NewFiberAuthMiddleware(tokenSvc, log)

	// Register routes
	api := app.Group("/api")
	userHandler.RegisterRoutes(api)

	// Protected routes
	// protected := app.Group("/api/protected")
	// protected.Use(authMiddleware.Authenticate)

	// Admin routes example
	// admin := app.Group("/admin")
	// admin.Use(authMiddleware.Authenticate)
	// Example of role-based middleware
	// admin.Use(authMiddleware.RequireRole("admin"))

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
