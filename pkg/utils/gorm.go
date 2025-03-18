package utils

import (
	"time"

	"gorm.io/gorm"
)

func DeletedAtPtrToTimePtr(deletedAt *gorm.DeletedAt) *time.Time {
	if deletedAt != nil && deletedAt.Valid {
		return &deletedAt.Time
	}
	return nil
}

// TimePtrToDeletedAt converts *time.Time to *gorm.DeletedAt.
// If the time pointer is not nil, it returns a valid *gorm.DeletedAt.
// If the time pointer is nil, it returns a zero *gorm.DeletedAt.
func TimePtrToDeletedAt(t *time.Time) *gorm.DeletedAt {
	if t != nil {
		return &gorm.DeletedAt{
			Time:  *t,
			Valid: true,
		}
	}
	return &gorm.DeletedAt{
		Time:  time.Time{}, // Zero time
		Valid: false,
	}
}
