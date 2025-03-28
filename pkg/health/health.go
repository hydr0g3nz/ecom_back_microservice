package health

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// Check represents a health check function
type Check func(ctx context.Context) error

// Health contains handlers for health checks
type Health struct {
	logger      logger.Logger
	startTime   time.Time
	dbSession   *gocql.Session
	kafkaBroker string
	checks      map[string]Check
}

// NewHealth creates a new Health instance
func NewHealth(log logger.Logger, dbSession *gocql.Session, kafkaBroker string) *Health {
	h := &Health{
		logger:      log,
		startTime:   time.Now(),
		dbSession:   dbSession,
		kafkaBroker: kafkaBroker,
		checks:      make(map[string]Check),
	}

	// Register standard checks
	h.RegisterCheck("db", h.checkDatabase)
	h.RegisterCheck("kafka", h.checkKafka)

	return h
}

// RegisterCheck registers a new health check
func (h *Health) RegisterCheck(name string, check Check) {
	h.checks[name] = check
}

// GetHandlers returns Fiber handlers for health check endpoints
func (h *Health) GetHandlers() map[string]fiber.Handler {
	return map[string]fiber.Handler{
		"/health":        h.HealthHandler,
		"/health/ready":  h.ReadinessHandler,
		"/health/live":   h.LivenessHandler,
		"/health/info":   h.InfoHandler,
		"/health/status": h.StatusHandler,
	}
}

// checkDatabase checks if the database is reachable
func (h *Health) checkDatabase(ctx context.Context) error {
	if h.dbSession == nil {
		return errors.New("database session not initialized")
	}

	// Execute a simple query to check connectivity
	if err := h.dbSession.Query("SELECT now() FROM system.local").WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("database check failed: %w", err)
	}

	return nil
}

// checkKafka checks if Kafka is reachable
func (h *Health) checkKafka(ctx context.Context) error {
	// In a real implementation, we would check if Kafka is reachable
	// For now, just return success if broker is configured
	if h.kafkaBroker == "" {
		return errors.New("kafka broker not configured")
	}

	return nil
}

// runChecks runs all registered health checks
func (h *Health) runChecks(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	for name, check := range h.checks {
		results[name] = check(ctx)
	}
	
	return results
}

// HealthHandler handles the /health endpoint
func (h *Health) HealthHandler(c *fiber.Ctx) error {
	// Run all checks
	results := h.runChecks(c.Context())
	
	// Check if all checks passed
	allPassed := true
	statusDetails := make(map[string]string)
	
	for name, err := range results {
		if err != nil {
			allPassed = false
			statusDetails[name] = "down"
		} else {
			statusDetails[name] = "up"
		}
	}
	
	status := "up"
	if !allPassed {
		status = "degraded"
		c.Status(fiber.StatusServiceUnavailable)
	}
	
	return c.JSON(fiber.Map{
		"status":  status,
		"details": statusDetails,
	})
}

// ReadinessHandler handles the /health/ready endpoint
func (h *Health) ReadinessHandler(c *fiber.Ctx) error {
	// Run all checks
	results := h.runChecks(c.Context())
	
	// Check if all checks passed
	allPassed := true
	
	for _, err := range results {
		if err != nil {
			allPassed = false
			break
		}
	}
	
	if !allPassed {
		c.Status(fiber.StatusServiceUnavailable)
		return c.JSON(fiber.Map{
			"status": "not ready",
		})
	}
	
	return c.JSON(fiber.Map{
		"status": "ready",
	})
}

// LivenessHandler handles the /health/live endpoint
func (h *Health) LivenessHandler(c *fiber.Ctx) error {
	// Liveness check always returns success if the service is running
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}

// InfoHandler handles the /health/info endpoint
func (h *Health) InfoHandler(c *fiber.Ctx) error {
	// Return basic service information
	info := map[string]interface{}{
		"service":     "order-service",
		"version":     "1.0.0", // This should come from a version package
		"build_time":  "2025-03-22", // This should be injected at build time
		"start_time":  h.startTime.Format(time.RFC3339),
		"uptime":      time.Since(h.startTime).String(),
		"go_version":  runtime.Version(),
		"go_os":       runtime.GOOS,
		"go_arch":     runtime.GOARCH,
		"goroutines":  runtime.NumGoroutine(),
		"cpu_cores":   runtime.NumCPU(),
	}

	return c.JSON(info)
}

// StatusHandler handles the /health/status endpoint
func (h *Health) StatusHandler(c *fiber.Ctx) error {
	// Run all checks
	results := h.runChecks(c.Context())
	
	// Prepare detailed response
	statusDetails := make(map[string]interface{})
	
	for name, err := range results {
		details := map[string]interface{}{
			"status": "up",
			"error":  nil,
		}
		
		if err != nil {
			details["status"] = "down"
			details["error"] = err.Error()
		}
		
		statusDetails[name] = details
	}
	
	// Include memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	memory := map[string]interface{}{
		"alloc":        memStats.Alloc,
		"total_alloc":  memStats.TotalAlloc,
		"sys":          memStats.Sys,
		"num_gc":       memStats.NumGC,
		"heap_objects": memStats.HeapObjects,
	}
	
	return c.JSON(fiber.Map{
		"components": statusDetails,
		"memory":     memory,
		"uptime":     time.Since(h.startTime).String(),
	})
}
