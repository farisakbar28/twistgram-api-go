package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type NotificationHandler struct{ service *service.NotificationService }

func NewNotificationHandlerWithService(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

func (h *NotificationHandler) ReadAll(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.InternalError(c, "Failed to mark notifications read")
		return
	}
	response.Success(c, gin.H{"read_all": true})
}

func (h *NotificationHandler) Create(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	var req dto.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if err := h.service.Create(c.Request.Context(), userID, req); err != nil {
		response.InternalError(c, "Failed to create notification")
		return
	}
	response.Created(c, gin.H{"created": true})
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	page := queryIntDM(c, "page", 1)
	limit := queryIntDM(c, "limit", 20)
	items, total, err := h.service.List(c.Request.Context(), userID, page, limit)
	if h.handleError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"notifications": items}, &response.Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages(total, limit)})
}

func (h *NotificationHandler) Read(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	notificationID, ok := parseNotificationID(c)
	if !ok {
		return
	}
	if h.handleError(c, h.service.MarkRead(c.Request.Context(), userID, notificationID)) {
		return
	}
	response.Success(c, gin.H{"read": true})
}

func (h *NotificationHandler) handleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrNotificationNotFound):
		response.NotFound(c, "Notification not found")
	default:
		response.InternalError(c, "Failed to process notification request")
	}
	return true
}

func parseNotificationID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid notification id")
		return uuid.Nil, false
	}
	return id, true
}
