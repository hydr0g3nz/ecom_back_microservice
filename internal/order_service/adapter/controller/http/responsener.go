package http

import (
	"encoding/json"
	"net/http"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/dto"
)

// Responsener handles HTTP responses
type Responsener struct{}

// NewResponsner creates a new responsener
func NewResponsner() *Responsener {
	return &Responsener{}
}

// JSON sends a JSON response
func (r *Responsener) JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but continue
			// TODO: implement proper logging
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}

// Success sends a success response
func (r *Responsener) Success(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusOK, dto.StatusResponse{
		Success: true,
		Message: message,
	})
}

// Created sends a resource created response
func (r *Responsener) Created(w http.ResponseWriter, data interface{}) {
	r.JSON(w, http.StatusCreated, data)
}

// NoContent sends a no content response
func (r *Responsener) NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a bad request response
func (r *Responsener) BadRequest(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusBadRequest, dto.ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

// NotFound sends a not found response
func (r *Responsener) NotFound(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusNotFound, dto.ErrorResponse{
		Code:    http.StatusNotFound,
		Message: message,
	})
}

// InternalServerError sends an internal server error response
func (r *Responsener) InternalServerError(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusInternalServerError, dto.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: message,
	})
}

// Unauthorized sends an unauthorized response
func (r *Responsener) Unauthorized(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusUnauthorized, dto.ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: message,
	})
}

// Forbidden sends a forbidden response
func (r *Responsener) Forbidden(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusForbidden, dto.ErrorResponse{
		Code:    http.StatusForbidden,
		Message: message,
	})
}

// UnprocessableEntity sends an unprocessable entity response
func (r *Responsener) UnprocessableEntity(w http.ResponseWriter, message string) {
	r.JSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
		Code:    http.StatusUnprocessableEntity,
		Message: message,
	})
}
