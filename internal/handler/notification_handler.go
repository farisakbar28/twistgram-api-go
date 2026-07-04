package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type NotificationHandler struct{ service *service.NotificationService }

func NewNotificationHandler() *NotificationHandler {
	repo := repository.NewNotificationRepository(config.GetDB())
	return &NotificationHandler{service: service.NewNotificationService(repo)}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	page := queryIntDM(c, "page", 1)
	limit := queryIntDM(c, "limit", 20)
	items, total, err := h.service.List(userID, page, limit)
	if h.handleError(c, err) { return }
	response.WithPagination(c, gin.H{"notifications": items}, &response.Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages(total, limit)})
}

func (h *NotificationHandler) Read(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	notificationID, ok := parseNotificationID(c); if !ok { return }
	if h.handleError(c, h.service.MarkRead(userID, notificationID)) { return }
	response.Success(c, gin.H{"read": true})
}

func (h *NotificationHandler) handleError(c *gin.Context, err error) bool {
	if err == nil { return false }
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

func parseNotificationID(c *gin.Context) (uuid.UUID, bool) { id, err := uuid.Parse(c.Param("id")); if err != nil { response.BadRequest(c, "Invalid notification id"); return uuid.Nil, false }; return id, true }
