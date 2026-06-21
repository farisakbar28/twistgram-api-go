package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response envelope
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Meta holds pagination metadata
type Meta struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

// PaginatedResponse wraps data with pagination meta
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Success sends a 200 success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Success",
		Data:    data,
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Created successfully",
		Data:    data,
	})
}

// Error sends an error response with the given status code and message
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// BadRequest sends a 400 error
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 error
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 error
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 error
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError sends a 500 error
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// WithPagination sends a success response with pagination metadata
func WithPagination(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Message: "Success",
		Data:    data,
		Meta:    meta,
	})
}
