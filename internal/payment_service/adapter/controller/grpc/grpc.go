// internal/product_service/adapter/controller/grpc/product_grpc.go
package grpcctl

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// ProductServer implements the gRPC ProductService interface
type ProductServer struct {
	pb.UnimplementedProductServiceServer
	productUsecase   usecase.ProductUsecase
	categoryUsecase  usecase.CategoryUsecase
	inventoryUsecase usecase.InventoryUsecase
	logger           logger.Logger
}

// NewProductServer creates a new ProductServer instance
func NewProductServer(
	pu usecase.ProductUsecase,
	cu usecase.CategoryUsecase,
	iu usecase.InventoryUsecase,
	logger logger.Logger,
) *ProductServer {
	return &ProductServer{
		productUsecase:   pu,
		categoryUsecase:  cu,
		inventoryUsecase: iu,
		logger:           logger,
	}
}

// CreateProduct creates a new product
func (s *ProductServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC CreateProduct request received", "name", req.Name, "sku", req.Sku)

	product := entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryId,
		ImageURL:    req.ImageUrl,
		SKU:         req.Sku,
		Status:      req.Status,
	}

	createdProduct, err := s.productUsecase.CreateProduct(ctx, &product)
	if err != nil {
		s.logger.Error("Failed to create product", "error", err)
		return nil, handleError(err)
	}

	return convertProductToProto(createdProduct), nil
}

// GetProduct gets a product by ID
func (s *ProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC GetProduct request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	product, err := s.productUsecase.GetProductByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get product", "error", err)
		return nil, handleError(err)
	}

	return convertProductToProto(product), nil
}

// GetProductBySKU gets a product by SKU
func (s *ProductServer) GetProductBySKU(ctx context.Context, req *pb.GetProductBySKURequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC GetProductBySKU request received", "sku", req.Sku)

	if req.Sku == "" {
		return nil, status.Error(codes.InvalidArgument, "product SKU is required")
	}

	product, err := s.productUsecase.GetProductBySKU(ctx, req.Sku)
	if err != nil {
		s.logger.Error("Failed to get product by SKU", "error", err)
		return nil, handleError(err)
	}

	return convertProductToProto(product), nil
}

// ListProducts lists products with optional filtering
func (s *ProductServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	s.logger.Info("gRPC ListProducts request received", "page", req.Page, "pageSize", req.PageSize)

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Convert filters from proto map to Go map
	filters := make(map[string]interface{})
	for k, v := range req.Filters {
		filters[k] = v
	}

	products, total, err := s.productUsecase.ListProducts(ctx, page, pageSize, filters)
	if err != nil {
		s.logger.Error("Failed to list products", "error", err)
		return nil, handleError(err)
	}

	// Convert entities to proto responses
	protoProducts := make([]*pb.ProductResponse, len(products))
	for i, product := range products {
		protoProducts[i] = convertProductToProto(product)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &pb.ListProductsResponse{
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: int32(totalPages),
		Products:   protoProducts,
	}, nil
}

// UpdateProduct updates an existing product
func (s *ProductServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC UpdateProduct request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	product := entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryId,
		ImageURL:    req.ImageUrl,
		SKU:         req.Sku,
		Status:      req.Status,
	}

	updatedProduct, err := s.productUsecase.UpdateProduct(ctx, req.Id, product)
	if err != nil {
		s.logger.Error("Failed to update product", "error", err)
		return nil, handleError(err)
	}

	return convertProductToProto(updatedProduct), nil
}

// DeleteProduct deletes a product by ID
func (s *ProductServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	s.logger.Info("gRPC DeleteProduct request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	err := s.productUsecase.DeleteProduct(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to delete product", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetProductsByCategory gets products by category ID
func (s *ProductServer) GetProductsByCategory(ctx context.Context, req *pb.GetProductsByCategoryRequest) (*pb.ListProductsResponse, error) {
	s.logger.Info("gRPC GetProductsByCategory request received", "categoryId", req.CategoryId)

	if req.CategoryId == "" {
		return nil, status.Error(codes.InvalidArgument, "category ID is required")
	}

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	products, total, err := s.productUsecase.GetProductsByCategory(ctx, req.CategoryId, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to get products by category", "error", err)
		return nil, handleError(err)
	}

	// Convert entities to proto responses
	protoProducts := make([]*pb.ProductResponse, len(products))
	for i, product := range products {
		protoProducts[i] = convertProductToProto(product)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &pb.ListProductsResponse{
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: int32(totalPages),
		Products:   protoProducts,
	}, nil
}

// CreateCategory creates a new category
func (s *ProductServer) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	s.logger.Info("gRPC CreateCategory request received", "name", req.Name)

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	category := entity.Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
	}

	createdCategory, err := s.categoryUsecase.CreateCategory(ctx, &category)
	if err != nil {
		s.logger.Error("Failed to create category", "error", err)
		return nil, handleError(err)
	}

	return convertCategoryToProto(createdCategory), nil
}

// GetCategory gets a category by ID
func (s *ProductServer) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.CategoryResponse, error) {
	s.logger.Info("gRPC GetCategory request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category ID is required")
	}

	category, err := s.categoryUsecase.GetCategoryByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get category", "error", err)
		return nil, handleError(err)
	}

	return convertCategoryToProto(category), nil
}

// ListCategories lists categories with pagination
func (s *ProductServer) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	s.logger.Info("gRPC ListCategories request received", "page", req.Page, "pageSize", req.PageSize)

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	categories, total, err := s.categoryUsecase.ListCategories(ctx, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list categories", "error", err)
		return nil, handleError(err)
	}

	// Convert entities to proto responses
	protoCategories := make([]*pb.CategoryResponse, len(categories))
	for i, category := range categories {
		protoCategories[i] = convertCategoryToProto(category)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &pb.ListCategoriesResponse{
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: int32(totalPages),
		Categories: protoCategories,
	}, nil
}

// UpdateCategory updates an existing category
func (s *ProductServer) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryResponse, error) {
	s.logger.Info("gRPC UpdateCategory request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category ID is required")
	}

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	category := entity.Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
	}

	updatedCategory, err := s.categoryUsecase.UpdateCategory(ctx, req.Id, category)
	if err != nil {
		s.logger.Error("Failed to update category", "error", err)
		return nil, handleError(err)
	}

	return convertCategoryToProto(updatedCategory), nil
}

// DeleteCategory deletes a category by ID
func (s *ProductServer) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	s.logger.Info("gRPC DeleteCategory request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category ID is required")
	}

	err := s.categoryUsecase.DeleteCategory(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to delete category", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetChildCategories gets child categories for a parent category
func (s *ProductServer) GetChildCategories(ctx context.Context, req *pb.GetChildCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	s.logger.Info("gRPC GetChildCategories request received", "parentId", req.ParentId)

	if req.ParentId == "" {
		return nil, status.Error(codes.InvalidArgument, "parent category ID is required")
	}

	categories, err := s.categoryUsecase.GetChildCategories(ctx, req.ParentId)
	if err != nil {
		s.logger.Error("Failed to get child categories", "error", err)
		return nil, handleError(err)
	}

	// Convert entities to proto responses
	protoCategories := make([]*pb.CategoryResponse, len(categories))
	for i, category := range categories {
		protoCategories[i] = convertCategoryToProto(category)
	}

	return &pb.ListCategoriesResponse{
		Total:      int32(len(categories)),
		Page:       1,
		PageSize:   int32(len(categories)),
		TotalPages: 1,
		Categories: protoCategories,
	}, nil
}

// GetInventory gets inventory for a product
func (s *ProductServer) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.InventoryResponse, error) {
	s.logger.Info("gRPC GetInventory request received", "productId", req.ProductId)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	inventory, err := s.inventoryUsecase.GetInventory(ctx, req.ProductId)
	if err != nil {
		s.logger.Error("Failed to get inventory", "error", err)
		return nil, handleError(err)
	}

	return convertInventoryToProto(inventory), nil
}

// UpdateInventory updates inventory for a product
func (s *ProductServer) UpdateInventory(ctx context.Context, req *pb.UpdateInventoryRequest) (*pb.InventoryResponse, error) {
	s.logger.Info("gRPC UpdateInventory request received", "productId", req.ProductId, "quantity", req.Quantity)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	err := s.inventoryUsecase.UpdateInventory(ctx, req.ProductId, int(req.Quantity))
	if err != nil {
		s.logger.Error("Failed to update inventory", "error", err)
		return nil, handleError(err)
	}

	// Get updated inventory
	inventory, err := s.inventoryUsecase.GetInventory(ctx, req.ProductId)
	if err != nil {
		s.logger.Error("Failed to get updated inventory", "error", err)
		return nil, handleError(err)
	}

	return convertInventoryToProto(inventory), nil
}

// ReserveStock reserves stock for a product
func (s *ProductServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*emptypb.Empty, error) {
	s.logger.Info("gRPC ReserveStock request received", "productId", req.ProductId, "quantity", req.Quantity)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}

	err := s.inventoryUsecase.ReserveStock(ctx, req.ProductId, int(req.Quantity))
	if err != nil {
		s.logger.Error("Failed to reserve stock", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// ConfirmReservation confirms a reservation
func (s *ProductServer) ConfirmReservation(ctx context.Context, req *pb.ReserveStockRequest) (*emptypb.Empty, error) {
	s.logger.Info("gRPC ConfirmReservation request received", "productId", req.ProductId, "quantity", req.Quantity)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}

	err := s.inventoryUsecase.ConfirmReservation(ctx, req.ProductId, int(req.Quantity))
	if err != nil {
		s.logger.Error("Failed to confirm reservation", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// CancelReservation cancels a reservation
func (s *ProductServer) CancelReservation(ctx context.Context, req *pb.ReserveStockRequest) (*emptypb.Empty, error) {
	s.logger.Info("gRPC CancelReservation request received", "productId", req.ProductId, "quantity", req.Quantity)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}

	err := s.inventoryUsecase.CancelReservation(ctx, req.ProductId, int(req.Quantity))
	if err != nil {
		s.logger.Error("Failed to cancel reservation", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// CheckStock checks if a product is in stock
func (s *ProductServer) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	s.logger.Info("gRPC CheckStock request received", "productId", req.ProductId, "quantity", req.Quantity)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}

	inStock, err := s.inventoryUsecase.IsInStock(ctx, req.ProductId, int(req.Quantity))
	if err != nil {
		s.logger.Error("Failed to check stock", "error", err)
		return nil, handleError(err)
	}

	return &pb.CheckStockResponse{
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
		InStock:   inStock,
	}, nil
}
func (s *ProductServer) PatchProduct(ctx context.Context, req *pb.PatchProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("gRPC PatchProduct request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	// Convert the fields map from protobuf to Go map for the usecase
	patchData := make(map[string]interface{})

	// Process fields based on which ones are set in the request
	if req.Name != nil {
		patchData["name"] = req.GetName()
	}
	if req.Description != nil {
		patchData["description"] = req.GetDescription()
	}
	if req.Price != nil {
		patchData["price"] = req.GetPrice()
	}
	if req.CategoryId != nil {
		patchData["category_id"] = req.GetCategoryId()
	}
	if req.ImageUrl != nil {
		patchData["image_url"] = req.GetImageUrl()
	}
	if req.Sku != nil {
		patchData["sku"] = req.GetSku()
	}
	if req.Status != nil {
		patchData["status"] = req.GetStatus()
	}

	updatedProduct, err := s.productUsecase.UpdateProductPartial(ctx, req.Id, patchData)
	if err != nil {
		s.logger.Error("Failed to patch product", "error", err)
		return nil, handleError(err)
	}

	return convertProductToProto(updatedProduct), nil
}

// Add a PatchCategory method to handle partial category updates
func (s *ProductServer) PatchCategory(ctx context.Context, req *pb.PatchCategoryRequest) (*pb.CategoryResponse, error) {
	s.logger.Info("gRPC PatchCategory request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "category ID is required")
	}

	// Convert the fields map from protobuf to Go map for the usecase
	patchData := make(map[string]interface{})

	// Process fields based on which ones are set in the request
	if req.Name != nil {
		patchData["name"] = req.GetName()
	}
	if req.Description != nil {
		patchData["description"] = req.GetDescription()
	}
	if req.ParentId != nil {
		if req.GetParentId() == "" {
			// Empty string represents removing the parent (making it top-level)
			patchData["parent_id"] = nil
		} else {
			patchData["parent_id"] = req.GetParentId()
		}
	}

	updatedCategory, err := s.categoryUsecase.UpdateCategoryPartial(ctx, req.Id, patchData)
	if err != nil {
		s.logger.Error("Failed to patch category", "error", err)
		return nil, handleError(err)
	}

	return convertCategoryToProto(updatedCategory), nil
}

// Add a PatchInventory method to handle partial inventory updates
func (s *ProductServer) PatchInventory(ctx context.Context, req *pb.PatchInventoryRequest) (*pb.InventoryResponse, error) {
	s.logger.Info("gRPC PatchInventory request received", "productId", req.ProductId)

	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID is required")
	}

	// Convert the fields map from protobuf to Go map for the usecase
	patchData := make(map[string]interface{})

	// Process fields based on which ones are set in the request
	if req.Quantity != nil {
		patchData["quantity"] = int(req.GetQuantity())
	}
	if req.Reserved != nil {
		patchData["reserved"] = int(req.GetReserved())
	}

	updatedInventory, err := s.inventoryUsecase.UpdateInventoryPartial(ctx, req.ProductId, patchData)
	if err != nil {
		s.logger.Error("Failed to patch inventory", "error", err)
		return nil, handleError(err)
	}

	return convertInventoryToProto(updatedInventory), nil
}

// Helper functions to convert domain entities to protobuf responses
func convertProductToProto(product *entity.Product) *pb.ProductResponse {
	return &pb.ProductResponse{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CategoryId:  product.CategoryID,
		ImageUrl:    product.ImageURL,
		Sku:         product.SKU,
		Status:      product.Status,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}
}

func convertCategoryToProto(category *entity.Category) *pb.CategoryResponse {
	resp := &pb.CategoryResponse{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		Level:       int32(category.Level),
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}

	if category.ParentID != nil {
		resp.ParentId = category.ParentID
	}

	return resp
}

func convertInventoryToProto(inventory *entity.Inventory) *pb.InventoryResponse {
	return &pb.InventoryResponse{
		ProductId: inventory.ProductID,
		Quantity:  int32(inventory.Quantity),
		Reserved:  int32(inventory.Reserved),
		Available: int32(inventory.Quantity - inventory.Reserved),
		UpdatedAt: timestamppb.New(inventory.UpdatedAt),
	}
}

// handleError maps domain errors to appropriate gRPC status errors
func handleError(err error) error {
	var statusCode codes.Code
	var message string

	switch {
	case errors.Is(err, entity.ErrProductNotFound):
		statusCode = codes.NotFound
		message = "Product not found"
	case errors.Is(err, entity.ErrCategoryNotFound):
		statusCode = codes.NotFound
		message = "Category not found"
	case errors.Is(err, entity.ErrInventoryNotFound):
		statusCode = codes.NotFound
		message = "Inventory not found"
	case errors.Is(err, entity.ErrProductSKUExists):
		statusCode = codes.AlreadyExists
		message = "Product SKU already exists"
	case errors.Is(err, entity.ErrCategoryAlreadyExists):
		statusCode = codes.AlreadyExists
		message = "Category already exists"
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = codes.FailedPrecondition
		message = "Insufficient stock"
	case errors.Is(err, entity.ErrInvalidProductData) || errors.Is(err, entity.ErrInvalidCategoryData):
		statusCode = codes.InvalidArgument
		message = "Invalid data provided"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = codes.Internal
		message = "Internal server error"
	default:
		statusCode = codes.Internal
		message = "Something went wrong"
	}

	return status.Error(statusCode, message)
}
