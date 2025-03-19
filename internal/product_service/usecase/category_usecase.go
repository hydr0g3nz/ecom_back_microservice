package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

type CategoryUsecase interface {
	// CreateCategory creates a new category
	CreateCategory(ctx context.Context, category *entity.Category) (*entity.Category, error)

	// GetCategoryByID retrieves a category by ID
	GetCategoryByID(ctx context.Context, id string) (*entity.Category, error)

	// ListCategories retrieves a list of categories
	ListCategories(ctx context.Context, page, pageSize int) ([]*entity.Category, int, error)

	// GetChildCategories retrieves child categories for a parent category
	GetChildCategories(ctx context.Context, parentID string) ([]*entity.Category, error)

	// UpdateCategory updates an existing category
	UpdateCategory(ctx context.Context, id string, category entity.Category) (*entity.Category, error)
	UpdateCategoryPartial(ctx context.Context, id string, patch map[string]interface{}) (*entity.Category, error)
	// DeleteCategory deletes a category by ID
	DeleteCategory(ctx context.Context, id string) error
}

// categoryUsecase implements the CategoryUsecase interface
type categoryUsecase struct {
	categoryRepo repository.CategoryRepository
	productRepo  repository.ProductRepository
	errBuilder   *utils.ErrorBuilder
}

// NewCategoryUsecase creates a new instance of CategoryUsecase
func NewCategoryUsecase(
	cr repository.CategoryRepository,
	pr repository.ProductRepository,
) CategoryUsecase {
	return &categoryUsecase{
		categoryRepo: cr,
		productRepo:  pr,
		errBuilder:   utils.NewErrorBuilder("CategoryUsecase"),
	}
}

// CreateCategory creates a new category
func (cu *categoryUsecase) CreateCategory(ctx context.Context, category *entity.Category) (*entity.Category, error) {
	// Check if category with same name already exists
	existingCategory, err := cu.categoryRepo.GetByName(ctx, category.Name)
	if err == nil && existingCategory != nil {
		return nil, cu.errBuilder.Err(entity.ErrCategoryAlreadyExists)
	}

	// If parent ID is provided, validate parent exists
	if category.ParentID != nil && *category.ParentID != "" {
		parent, err := cu.categoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return nil, cu.errBuilder.Err(entity.ErrCategoryNotFound)
		}
		// Set level based on parent
		category.Level = parent.Level + 1
	} else {
		// Top-level category
		category.Level = 1
		category.ParentID = nil
	}

	// Generate ID if not provided
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	// Create category
	createdCategory, err := cu.categoryRepo.Create(ctx, *category)
	if err != nil {
		return nil, cu.errBuilder.Err(err)
	}

	return createdCategory, nil
}

// GetCategoryByID retrieves a category by ID
func (cu *categoryUsecase) GetCategoryByID(ctx context.Context, id string) (*entity.Category, error) {
	category, err := cu.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, cu.errBuilder.Err(err)
	}
	return category, nil
}

// ListCategories retrieves a list of categories
func (cu *categoryUsecase) ListCategories(ctx context.Context, page, pageSize int) ([]*entity.Category, int, error) {
	offset := (page - 1) * pageSize
	categories, total, err := cu.categoryRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, cu.errBuilder.Err(err)
	}
	return categories, total, nil
}

// GetChildCategories retrieves child categories for a parent category
func (cu *categoryUsecase) GetChildCategories(ctx context.Context, parentID string) ([]*entity.Category, error) {
	// Ensure the parent category exists
	_, err := cu.categoryRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, cu.errBuilder.Err(entity.ErrCategoryNotFound)
	}

	children, err := cu.categoryRepo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, cu.errBuilder.Err(err)
	}
	return children, nil
}

// UpdateCategory updates an existing category
func (cu *categoryUsecase) UpdateCategory(ctx context.Context, id string, category entity.Category) (*entity.Category, error) {
	// Ensure the category exists
	existingCategory, err := cu.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, cu.errBuilder.Err(entity.ErrCategoryNotFound)
	}

	// If name is changing, check if new name is unique
	if category.Name != existingCategory.Name {
		c, err := cu.categoryRepo.GetByName(ctx, category.Name)
		if err == nil && c != nil && c.ID != id {
			return nil, cu.errBuilder.Err(entity.ErrCategoryAlreadyExists)
		}
	}

	// If parent ID is changing, validate new parent exists and prevent circular references
	if category.ParentID != nil && *category.ParentID != "" &&
		(existingCategory.ParentID == nil || *existingCategory.ParentID != *category.ParentID) {

		// Prevent setting itself as parent
		if *category.ParentID == id {
			return nil, cu.errBuilder.Err(errors.New("category cannot be its own parent"))
		}

		parent, err := cu.categoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return nil, cu.errBuilder.Err(entity.ErrCategoryNotFound)
		}
		// Update level based on new parent
		category.Level = parent.Level + 1
	} else if category.ParentID == nil || *category.ParentID == "" {
		// Changing to top-level category
		category.Level = 1
		category.ParentID = nil
	}

	// Set ID to ensure we're updating the correct record
	category.ID = id

	// Update the category
	updatedCategory, err := cu.categoryRepo.Update(ctx, category)
	if err != nil {
		return nil, cu.errBuilder.Err(err)
	}

	return updatedCategory, nil
}

// DeleteCategory deletes a category by ID
func (cu *categoryUsecase) DeleteCategory(ctx context.Context, id string) error {
	// Ensure the category exists
	_, err := cu.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return cu.errBuilder.Err(entity.ErrCategoryNotFound)
	}

	// Check if there are child categories
	children, err := cu.categoryRepo.GetChildren(ctx, id)
	if err != nil {
		return cu.errBuilder.Err(err)
	}
	if len(children) > 0 {
		return cu.errBuilder.Err(errors.New("cannot delete category with child categories"))
	}

	// Check if there are products in this category
	products, _, err := cu.productRepo.GetByCategory(ctx, id, 0, 1)
	if err != nil {
		return cu.errBuilder.Err(err)
	}
	if len(products) > 0 {
		return cu.errBuilder.Err(errors.New("cannot delete category with associated products"))
	}

	// Delete the category
	err = cu.categoryRepo.Delete(ctx, id)
	if err != nil {
		return cu.errBuilder.Err(err)
	}

	return nil
}
func (cu *categoryUsecase) UpdateCategoryPartial(ctx context.Context, id string, patch map[string]interface{}) (*entity.Category, error) {
	// Ensure the category exists
	existingCategory, err := cu.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, cu.errBuilder.Err(entity.ErrCategoryNotFound)
	}

	// Create a modifiable copy of the existing category
	updatedCategory := *existingCategory

	// Apply updates from the patch map
	for key, value := range patch {
		switch key {
		case "name":
			if name, ok := value.(string); ok && name != "" {
				// Check if name is unique
				if name != existingCategory.Name {
					c, err := cu.categoryRepo.GetByName(ctx, name)
					if err == nil && c != nil && c.ID != id {
						return nil, cu.errBuilder.Err(entity.ErrCategoryAlreadyExists)
					}
				}
				updatedCategory.Name = name
			}
		case "description":
			if description, ok := value.(string); ok {
				updatedCategory.Description = description
			}
		case "parent_id":
			if parentID, ok := value.(string); ok && parentID != "" {
				// Prevent setting itself as parent
				if parentID == id {
					return nil, cu.errBuilder.Err(errors.New("category cannot be its own parent"))
				}

				// Validate parent exists
				parent, err := cu.categoryRepo.GetByID(ctx, parentID)
				if err != nil {
					return nil, cu.errBuilder.Err(entity.ErrCategoryNotFound)
				}

				// Set parentID and update level
				parentIDCopy := parentID // Create a copy to use as pointer
				updatedCategory.ParentID = &parentIDCopy
				updatedCategory.Level = parent.Level + 1
			} else if value == nil {
				// Remove parent (make it a top-level category)
				updatedCategory.ParentID = nil
				updatedCategory.Level = 1
			}
		}
	}

	// Update the category
	result, err := cu.categoryRepo.Update(ctx, updatedCategory)
	if err != nil {
		return nil, cu.errBuilder.Err(err)
	}

	return result, nil
}
