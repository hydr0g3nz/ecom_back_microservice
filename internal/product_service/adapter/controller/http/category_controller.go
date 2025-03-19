package httpctl

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/dto"
)

// CreateCategory handles the creation of a new category
func (h *ProductHandler) CreateCategory(c *fiber.Ctx) error {
	var req dto.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	category := req.ToEntity()
	createdCategory, err := h.categoryUsecase.CreateCategory(ctx, &category)
	if err != nil {
		h.logger.Error("Failed to create category", "error", err)
		return HandleError(c, err)
	}

	response := dto.CategoryResponseFromEntity(createdCategory)
	return SuccessResp(c, fiber.StatusCreated, "Category created successfully", response)
}

// GetCategory handles retrieving a category by ID
func (h *ProductHandler) GetCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	category, err := h.categoryUsecase.GetCategoryByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get category", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.CategoryResponseFromEntity(category)
	return SuccessResp(c, fiber.StatusOK, "Category retrieved successfully", response)
}

// ListCategories handles retrieving a list of categories
func (h *ProductHandler) ListCategories(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	categories, total, err := h.categoryUsecase.ListCategories(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list categories", "error", err)
		return HandleError(c, err)
	}

	// Convert entities to response DTOs
	responseCategories := make([]dto.CategoryResponse, len(categories))
	for i, category := range categories {
		responseCategories[i] = dto.CategoryResponseFromEntity(category)
	}

	paginatedResponse := dto.NewPaginatedResponse(total, page, pageSize, responseCategories)
	return SuccessResp(c, fiber.StatusOK, "Categories retrieved successfully", paginatedResponse)
}

// UpdateCategory handles updating an existing category
func (h *ProductHandler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	category := req.ToEntity()
	updatedCategory, err := h.categoryUsecase.UpdateCategory(ctx, id, category)
	if err != nil {
		h.logger.Error("Failed to update category", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.CategoryResponseFromEntity(updatedCategory)
	return SuccessResp(c, fiber.StatusOK, "Category updated successfully", response)
}

// DeleteCategory handles deleting a category
func (h *ProductHandler) DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.categoryUsecase.DeleteCategory(ctx, id)
	if err != nil {
		h.logger.Error("Failed to delete category", "id", id, "error", err)
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Category deleted successfully", nil)
}

// GetChildCategories handles retrieving child categories for a parent category
func (h *ProductHandler) GetChildCategories(c *fiber.Ctx) error {
	parentId := c.Params("id")
	if parentId == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	children, err := h.categoryUsecase.GetChildCategories(ctx, parentId)
	if err != nil {
		h.logger.Error("Failed to get child categories", "parentId", parentId, "error", err)
		return HandleError(c, err)
	}

	// Convert entities to response DTOs
	responseCategories := make([]dto.CategoryResponse, len(children))
	for i, category := range children {
		responseCategories[i] = dto.CategoryResponseFromEntity(category)
	}

	return SuccessResp(c, fiber.StatusOK, "Child categories retrieved successfully", responseCategories)
}
