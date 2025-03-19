package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

type ProductUsecase interface {
	// CreateProduct creates a new product
	CreateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// GetProductByID retrieves a product by ID
	GetProductByID(ctx context.Context, id string) (*entity.Product, error)

	// GetProductBySKU retrieves a product by SKU
	GetProductBySKU(ctx context.Context, sku string) (*entity.Product, error)

	// ListProducts retrieves a list of products with optional filtering
	ListProducts(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Product, int, error)

	// UpdateProduct updates an existing product
	UpdateProduct(ctx context.Context, id string, product entity.Product) (*entity.Product, error)

	UpdateProductPartial(ctx context.Context, id string, patch map[string]interface{}) (*entity.Product, error)
	// DeleteProduct deletes a product by ID
	DeleteProduct(ctx context.Context, id string) error

	// GetProductsByCategory retrieves products by category ID
	GetProductsByCategory(ctx context.Context, categoryID string, page, pageSize int) ([]*entity.Product, int, error)
}

// productUsecase implements the ProductUsecase interface
type productUsecase struct {
	productRepo   repository.ProductRepository
	categoryRepo  repository.CategoryRepository
	inventoryRepo repository.InventoryRepository
	errBuilder    *utils.ErrorBuilder
}

// NewProductUsecase creates a new instance of ProductUsecase
func NewProductUsecase(
	pr repository.ProductRepository,
	cr repository.CategoryRepository,
	ir repository.InventoryRepository,
) ProductUsecase {
	return &productUsecase{
		productRepo:   pr,
		categoryRepo:  cr,
		inventoryRepo: ir,
		errBuilder:    utils.NewErrorBuilder("ProductUsecase"),
	}
}

// CreateProduct creates a new product
func (pu *productUsecase) CreateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	// Check if product with same SKU already exists
	existingProduct, err := pu.productRepo.GetBySKU(ctx, product.SKU)
	if err == nil && existingProduct != nil {
		return nil, pu.errBuilder.Err(entity.ErrProductSKUExists)
	}

	// Validate category exists
	_, err = pu.categoryRepo.GetByID(ctx, product.CategoryID)
	if err != nil {
		return nil, pu.errBuilder.Err(entity.ErrCategoryNotFound)
	}

	// Set default status if not provided
	if product.Status == "" {
		product.Status = valueobject.Active.String()
	}

	// Generate ID if not provided
	if product.ID == "" {
		product.ID = uuid.New().String()
	}

	// Create product
	createdProduct, err := pu.productRepo.Create(ctx, *product)
	if err != nil {
		return nil, pu.errBuilder.Err(err)
	}

	// Create inventory record with initial quantity of 0
	inventory := entity.Inventory{
		ID:        uuid.New().String(),
		ProductID: createdProduct.ID,
		Quantity:  0,
		Reserved:  0,
	}

	_, err = pu.inventoryRepo.Create(ctx, inventory)
	if err != nil {
		// Log the error but don't fail the product creation
		// In a real implementation, you might want to handle this better
		// Perhaps delete the product or use a transaction
	}

	return createdProduct, nil
}

// GetProductByID retrieves a product by ID
func (pu *productUsecase) GetProductByID(ctx context.Context, id string) (*entity.Product, error) {
	product, err := pu.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pu.errBuilder.Err(err)
	}
	return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (pu *productUsecase) GetProductBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	product, err := pu.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, pu.errBuilder.Err(err)
	}
	return product, nil
}

// ListProducts retrieves a list of products with optional filtering
func (pu *productUsecase) ListProducts(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Product, int, error) {
	offset := (page - 1) * pageSize
	products, total, err := pu.productRepo.List(ctx, offset, pageSize, filters)
	if err != nil {
		return nil, 0, pu.errBuilder.Err(err)
	}
	return products, total, nil
}

// UpdateProduct updates an existing product
func (pu *productUsecase) UpdateProduct(ctx context.Context, id string, product entity.Product) (*entity.Product, error) {
	// Ensure the product exists
	existingProduct, err := pu.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// If SKU is changing, check if new SKU is unique
	if product.SKU != existingProduct.SKU {
		p, err := pu.productRepo.GetBySKU(ctx, product.SKU)
		if err == nil && p != nil && p.ID != id {
			return nil, pu.errBuilder.Err(entity.ErrProductSKUExists)
		}
	}

	// If category is changing, validate new category exists
	if product.CategoryID != existingProduct.CategoryID {
		_, err = pu.categoryRepo.GetByID(ctx, product.CategoryID)
		if err != nil {
			return nil, pu.errBuilder.Err(entity.ErrCategoryNotFound)
		}
	}

	// Set ID to ensure we're updating the correct record
	product.ID = id

	// Update the product
	updatedProduct, err := pu.productRepo.Update(ctx, product)
	if err != nil {
		return nil, pu.errBuilder.Err(err)
	}

	return updatedProduct, nil
}

// DeleteProduct deletes a product by ID
func (pu *productUsecase) DeleteProduct(ctx context.Context, id string) error {
	// Ensure the product exists
	_, err := pu.productRepo.GetByID(ctx, id)
	if err != nil {
		return pu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Delete the product
	err = pu.productRepo.Delete(ctx, id)
	if err != nil {
		return pu.errBuilder.Err(err)
	}

	return nil
}

// GetProductsByCategory retrieves products by category ID
func (pu *productUsecase) GetProductsByCategory(ctx context.Context, categoryID string, page, pageSize int) ([]*entity.Product, int, error) {
	// Ensure the category exists
	_, err := pu.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, 0, pu.errBuilder.Err(entity.ErrCategoryNotFound)
	}

	offset := (page - 1) * pageSize
	products, total, err := pu.productRepo.GetByCategory(ctx, categoryID, offset, pageSize)
	if err != nil {
		return nil, 0, pu.errBuilder.Err(err)
	}

	return products, total, nil
}
func (pu *productUsecase) UpdateProductPartial(ctx context.Context, id string, patch map[string]interface{}) (*entity.Product, error) {
	// Ensure the product exists and get the current state
	existingProduct, err := pu.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Create a modifiable copy of the existing product
	updatedProduct := existingProduct

	// Apply updates from the patch map to the product entity
	for key, value := range patch {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				updatedProduct.Name = name
			}
		case "description":
			if description, ok := value.(string); ok {
				updatedProduct.Description = description
			}
		case "price":
			if price, ok := value.(float64); ok && price > 0 {
				updatedProduct.Price = price
			}
		case "category_id":
			if categoryID, ok := value.(string); ok {
				// Validate that the category exists
				_, err := pu.categoryRepo.GetByID(ctx, categoryID)
				if err != nil {
					return nil, pu.errBuilder.Err(entity.ErrCategoryNotFound)
				}
				updatedProduct.CategoryID = categoryID
			}
		case "image_url":
			if imageURL, ok := value.(string); ok {
				updatedProduct.ImageURL = imageURL
			}
		case "sku":
			if sku, ok := value.(string); ok && sku != existingProduct.SKU {
				// Check if SKU is unique
				p, err := pu.productRepo.GetBySKU(ctx, sku)
				if err == nil && p != nil && p.ID != id {
					return nil, pu.errBuilder.Err(entity.ErrProductSKUExists)
				}
				updatedProduct.SKU = sku
			}
		case "status":
			if status, ok := value.(string); ok {
				// Validate status
				if _, err := valueobject.ParseProductStatus(status); err != nil {
					return nil, pu.errBuilder.Err(err)
				}
				updatedProduct.Status = status
			}
		}
	}

	// Update the product
	updatedProduct, err = pu.productRepo.Update(ctx, *updatedProduct)
	if err != nil {
		return nil, pu.errBuilder.Err(err)
	}

	return updatedProduct, nil
}
