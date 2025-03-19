package httpctl

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// ProductHandler handles HTTP requests for the product service
type ProductHandler struct {
	productUsecase   usecase.ProductUsecase
	categoryUsecase  usecase.CategoryUsecase
	inventoryUsecase usecase.InventoryUsecase
	logger           logger.Logger
}

// NewProductHandler creates a new instance of ProductHandler
func NewProductHandler(
	pu usecase.ProductUsecase,
	cu usecase.CategoryUsecase,
	iu usecase.InventoryUsecase,
	l logger.Logger,
) *ProductHandler {
	return &ProductHandler{
		productUsecase:   pu,
		categoryUsecase:  cu,
		inventoryUsecase: iu,
		logger:           l,
	}
}

// RegisterRoutes registers the routes for the product service
func (h *ProductHandler) RegisterRoutes(r fiber.Router) {
	api := r.Group("/products")

	// Product routes
	api.Post("/", h.CreateProduct)
	api.Get("/", h.ListProducts)
	api.Get("/:id", h.GetProduct)
	api.Put("/:id", h.UpdateProduct)
	api.Delete("/:id", h.DeleteProduct)
	api.Get("/sku/:sku", h.GetProductBySKU)
	api.Get("/category/:categoryId", h.GetProductsByCategory)

	// Category routes
	categoryGroup := r.Group("/categories")
	categoryGroup.Post("/", h.CreateCategory)
	categoryGroup.Get("/", h.ListCategories)
	categoryGroup.Get("/:id", h.GetCategory)
	categoryGroup.Put("/:id", h.UpdateCategory)
	categoryGroup.Delete("/:id", h.DeleteCategory)
	categoryGroup.Get("/:id/children", h.GetChildCategories)

	// Inventory routes
	inventoryGroup := r.Group("/inventory")
	inventoryGroup.Get("/:productId", h.GetInventory)
	inventoryGroup.Put("/:productId", h.UpdateInventory)
	inventoryGroup.Post("/reserve", h.ReserveStock)
	inventoryGroup.Post("/release", h.ReleaseStock)
	inventoryGroup.Post("/confirm", h.ConfirmReservation)
	inventoryGroup.Get("/:productId/stock", h.CheckStock)
}

// CreateProduct handles the creation of a new product
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req dto.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	product := req.ToEntity()
	createdProduct, err := h.productUsecase.CreateProduct(ctx, &product)
	if err != nil {
		h.logger.Error("Failed to create product", "error", err)
		return HandleError(c, err)
	}

	response := dto.ProductResponseFromEntity(createdProduct)
	return SuccessResp(c, fiber.StatusCreated, "Product created successfully", response)
}

// GetProduct handles retrieving a product by ID
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	product, err := h.productUsecase.GetProductByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get product", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.ProductResponseFromEntity(product)
	return SuccessResp(c, fiber.StatusOK, "Product retrieved successfully", response)
}

// GetProductBySKU handles retrieving a product by SKU
func (h *ProductHandler) GetProductBySKU(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	product, err := h.productUsecase.GetProductBySKU(ctx, sku)
	if err != nil {
		h.logger.Error("Failed to get product by SKU", "sku", sku, "error", err)
		return HandleError(c, err)
	}

	response := dto.ProductResponseFromEntity(product)
	return SuccessResp(c, fiber.StatusOK, "Product retrieved successfully", response)
}

// ListProducts handles retrieving a list of products
func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// TODO: Implement more sophisticated filtering
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	ctx := c.Context()
	products, total, err := h.productUsecase.ListProducts(ctx, page, pageSize, filters)
	if err != nil {
		h.logger.Error("Failed to list products", "error", err)
		return HandleError(c, err)
	}

	// Convert entities to response DTOs
	responseProducts := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		responseProducts[i] = dto.ProductResponseFromEntity(product)
	}

	paginatedResponse := dto.NewPaginatedResponse(total, page, pageSize, responseProducts)
	return SuccessResp(c, fiber.StatusOK, "Products retrieved successfully", paginatedResponse)
}

// UpdateProduct handles updating an existing product
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	product := req.ToEntity()
	updatedProduct, err := h.productUsecase.UpdateProduct(ctx, id, product)
	if err != nil {
		h.logger.Error("Failed to update product", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.ProductResponseFromEntity(updatedProduct)
	return SuccessResp(c, fiber.StatusOK, "Product updated successfully", response)
}

// DeleteProduct handles deleting a product
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.productUsecase.DeleteProduct(ctx, id)
	if err != nil {
		h.logger.Error("Failed to delete product", "id", id, "error", err)
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Product deleted successfully", nil)
}

// GetProductsByCategory handles retrieving products by category ID
func (h *ProductHandler) GetProductsByCategory(c *fiber.Ctx) error {
	categoryId := c.Params("categoryId")
	if categoryId == "" {
		return HandleError(c, ErrBadRequest)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	products, total, err := h.productUsecase.GetProductsByCategory(ctx, categoryId, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get products by category", "categoryId", categoryId, "error", err)
		return HandleError(c, err)
	}

	// Convert entities to response DTOs
	responseProducts := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		responseProducts[i] = dto.ProductResponseFromEntity(product)
	}

	paginatedResponse := dto.NewPaginatedResponse(total, page, pageSize, responseProducts)
	return SuccessResp(c, fiber.StatusOK, "Products retrieved successfully", paginatedResponse)
}
