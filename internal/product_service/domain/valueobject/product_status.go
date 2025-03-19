package valueobject

import (
	"errors"
	"strings"
)

type ProductStatus string

const (
	Active   ProductStatus = "active"
	Inactive ProductStatus = "inactive"
	Draft    ProductStatus = "draft"
	Deleted  ProductStatus = "deleted"
)

func (s ProductStatus) String() string {
	return string(s)
}

func (s ProductStatus) IsValid() bool {
	statuses := [...]ProductStatus{Active, Inactive, Draft, Deleted}
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func ParseProductStatus(status string) (ProductStatus, error) {
	status = strings.ToLower(status)
	if !ProductStatus(status).IsValid() {
		return "", errors.New("invalid product status")
	}
	return ProductStatus(status), nil
}
