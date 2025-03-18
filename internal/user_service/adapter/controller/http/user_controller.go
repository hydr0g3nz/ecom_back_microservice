package httpctl

import (
	"github.com/gofiber/fiber/v2"
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

	// userGroup.Post("/", h.CreateUser)
	userGroup.Get("/:id", h.GetUser)
	userGroup.Put("/:id", h.UpdateUser)
	userGroup.Delete("/:id", h.DeleteUser)

	userGroup.Post("/login", h.Login)
	userGroup.Post("/register", h.Register)
}

// CreateUser handles the creation of a new user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.UserRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	userEntity := req.ToEntity()
	user, err := h.userUsecase.CreateUser(ctx, &userEntity, req.Password)
	if err != nil {
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusCreated, "User created", user)
}

// GetUser handles retrieving a user by ID
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	user, err := h.userUsecase.GetUserByID(ctx, id)
	if err != nil {
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "User retrieved", user)
}

// UpdateUser handles updating an existing user
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.UserRequest
	if err := c.BodyParser(&req); err != nil {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	user, err := h.userUsecase.UpdateUser(ctx, id, req.ToEntity())
	if err != nil {
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "User updated", user)
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.userUsecase.DeleteUser(ctx, id)
	if err != nil {
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusNoContent, "User deleted", nil)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.UserLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	tokenPair, err := h.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Login successful", tokenPair)
}
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req dto.UserRequest
	if err := c.BodyParser(&req); err != nil {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	user, tokenPair, err := h.authUsecase.Register(ctx, req.ToEntity(), req.Password)
	if err != nil {
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Registration successful", fiber.Map{
		"user":      user,
		"tokenPair": tokenPair,
	},
	)
}
