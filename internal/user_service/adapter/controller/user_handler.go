package handler

import (
	"github.com/gofiber/fiber/v2"
	responsener "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/reponsener"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/dto"
	uc "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// UserHandler handles HTTP requests for the user service
type UserHandler struct {
	authUsecase uc.AuthUsecase
	userUsecase uc.UserUsecase
	logger      logger.Logger
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(authUsecase uc.AuthUsecase, userUsecase uc.UserUsecase, logger logger.Logger) *UserHandler {
	return &UserHandler{
		authUsecase: authUsecase,
		userUsecase: userUsecase,
		logger:      logger,
	}
}

// RegisterRoutes registers the routes for the user service
func (h *UserHandler) RegisterRoutes(r fiber.Router) {
	userGroup := r.Group("/users")

	userGroup.Post("/", h.CreateUser)
	userGroup.Get("/:id", h.GetUser)
	userGroup.Put("/:id", h.UpdateUser)
	userGroup.Delete("/:id", h.DeleteUser)

	userGroup.Post("/login", h.Login)

}

// CreateUser handles the creation of a new user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.UserRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(responsener.ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid request format",
		})
	}

	ctx := c.Context()
	userEntity := req.ToEntity()
	user, err := h.userUsecase.CreateUser(ctx, &userEntity, req.Password)
	if err != nil {
		return h.handleServiceError(c, err, "Failed to create user")
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetUser handles retrieving a user by ID
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responsener.ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Missing user ID",
		})
	}

	ctx := c.Context()
	user, err := h.userUsecase.GetUserByID(ctx, id)
	if err != nil {
		return h.handleServiceError(c, err, "Failed to get user")
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// UpdateUser handles updating an existing user
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responsener.ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Missing user ID",
		})
	}

	var req dto.UserRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(responsener.ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid request format",
		})
	}

	ctx := c.Context()
	user, err := h.userUsecase.UpdateUser(ctx, id, req.ToEntity())
	if err != nil {
		return h.handleServiceError(c, err, "Failed to update user")
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responsener.ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Missing user ID",
		})
	}

	ctx := c.Context()
	err := h.userUsecase.DeleteUser(ctx, id)
	if err != nil {
		return h.handleServiceError(c, err, "Failed to delete user")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.UserLoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(responsener.ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid request format",
		})
	}

	ctx := c.Context()
	tokenPair, err := h.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return h.handleServiceError(c, err, "Failed to login")
	}

	return c.Status(fiber.StatusOK).JSON(responsener.SuccessResponse{
		Code:    fiber.StatusOK,
		Message: "Login successful",
		Data: responsener.LoginResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
	})
}

// handleServiceError handles errors from the service layer
