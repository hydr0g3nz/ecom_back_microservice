package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// RequestLogger adds request logging middleware
func RequestLogger(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()
		
		// Get correlation ID from context or use "unknown"
		correlationID := GetCorrelationID(c.UserContext())
		if correlationID == "" {
			correlationID = "unknown"
		}
		
		// Use logger with correlation ID
		reqLogger := log.WithCorrelationID(correlationID)
		
		// Store logger in context
		c.Locals("logger", reqLogger)
		
		// Log request
		reqLogger.Info("Request received",
			"method", c.Method(),
			"path", c.Path(),
			"query", c.Query(""),
			"ip", c.IP(),
			"user_agent", c.Get(fiber.HeaderUserAgent),
		)
		
		// Process request
		err := c.Next()
		
		// Calculate latency
		latency := time.Since(start)
		
		// Log response
		reqLogger.Info("Request completed",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", latency.Milliseconds(),
			"bytes", len(c.Response().Body()),
		)
		
		return err
	}
}
