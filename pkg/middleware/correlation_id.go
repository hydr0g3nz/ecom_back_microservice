package middleware

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// CorrelationIDHeader is the HTTP header name for the correlation ID
const CorrelationIDHeader = "X-Correlation-ID"

// contextKey is the type for context keys
type contextKey string

// CorrelationIDKey is the context key for the correlation ID
const CorrelationIDKey contextKey = "correlation_id"

// CorrelationID adds a correlation ID to the request context and response headers
func CorrelationID(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get correlation ID from header or generate a new one
		correlationID := c.Get(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add correlation ID to response headers
		c.Set(CorrelationIDHeader, correlationID)

		// Add correlation ID to context
		ctx := context.WithValue(c.Context(), CorrelationIDKey, correlationID)
		c.SetUserContext(ctx)

		// Use logger with correlation ID
		c.Locals("logger", log.WithCorrelationID(correlationID))

		// Call the next handler
		return c.Next()
	}
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	value := ctx.Value(CorrelationIDKey)
	if value == nil {
		return ""
	}
	return value.(string)
}
