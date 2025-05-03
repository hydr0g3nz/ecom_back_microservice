package valueobject

import (
	"errors"
	"strings"
)

type StockType string

const (
	StockTypeReserved StockType = "RESERVE"
	StockTypeReleased StockType = "RELEASE"
	StockTypeDeducted StockType = "DEDUCT"
)

func (s StockType) String() string {
	return string(s)
}

func (s StockType) IsValid() bool {
	statuses := [...]StockType{StockTypeReserved, StockTypeReleased, StockTypeDeducted}
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func ParseStockType(status string) (StockType, error) {
	status = strings.ToLower(status)
	if !StockType(status).IsValid() {
		return "", errors.New("invalid stock type")
	}
	return StockType(status), nil
}
