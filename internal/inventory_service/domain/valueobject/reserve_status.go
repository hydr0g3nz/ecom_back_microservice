package valueobject

import (
	"errors"
	"strings"
)

type ReserveStatus string

const (
	ReserveStatusReserved  ReserveStatus = "RESERVED"
	ReserveStatusCompleted ReserveStatus = "COMPLETED"
	ReserveStatusCancelled ReserveStatus = "CANCELLED"
)

func (s ReserveStatus) String() string {
	return string(s)
}

func (s ReserveStatus) IsValid() bool {
	statuses := [...]ReserveStatus{ReserveStatusReserved, ReserveStatusCompleted, ReserveStatusCancelled}
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func ParseReserveStatus(status string) (ReserveStatus, error) {
	status = strings.ToLower(status)
	if !ReserveStatus(status).IsValid() {
		return "", errors.New("invalid reserve status")
	}
	return ReserveStatus(status), nil
}
