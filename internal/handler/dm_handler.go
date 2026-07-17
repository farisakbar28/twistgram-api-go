package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type DMHandler struct{ service *service.DMService }

func NewDMHandlerWithService(svc *service.DMService) *DMHandler {
	return &DMHandler{service: svc}
}

func (h *DMHandler) StartConversation(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	var req dto.StartConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		response.BadRequest(c, "Invalid target id")
		return
	}
	res, err := h.service.StartConversation(userID, targetID)
	if handleDMError(c, err) {
		return
	}
	response.Created(c, gin.H{"conversation": res})
}

func (h *DMHandler) ListConversations(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	pg := queryIntDM(c, "page", 1)
	lm := queryIntDM(c, "limit", 20)
	items, total, err := h.service.ListConversations(userID, pg, lm)
	if handleDMError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"conversations": items}, &response.Meta{Page: pg, Limit: lm, Total: total, TotalPages: totalPages(total, lm)})
}

func (h *DMHandler) ListMessages(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	conversationID, ok := parseConversationID(c)
	if !ok {
		return
	}
	pg := queryIntDM(c, "page", 1)
	lm := queryIntDM(c, "limit", 20)
	items, total, err := h.service.ListMessages(userID, conversationID, pg, lm)
	if handleDMError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"messages": items}, &response.Meta{Page: pg, Limit: lm, Total: total, TotalPages: totalPages(total, lm)})
}

func (h *DMHandler) SendMessage(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	conversationID, ok := parseConversationID(c)
	if !ok {
		return
	}
	var req dto.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.service.SendMessage(userID, conversationID, req)
	if handleDMError(c, err) {
		return
	}
	response.Created(c, gin.H{"message": res})
}

func (h *DMHandler) ListMessageRequests(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	pg := queryIntDM(c, "page", 1)
	lm := queryIntDM(c, "limit", 20)
	items, total, err := h.service.ListMessageRequests(userID, pg, lm)
	if handleDMError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"conversations": items}, &response.Meta{Page: pg, Limit: lm, Total: total, TotalPages: totalPages(total, lm)})
}

func handleDMError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrForbidden):
		response.Forbidden(c, "You do not have access to this conversation")
	case errors.Is(err, service.ErrConversationNotFound):
		response.NotFound(c, "Conversation not found")
	default:
		response.InternalError(c, "Failed to process direct message request")
	}
	return true
}

func parseConversationID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid conversation id")
		return uuid.Nil, false
	}
	return id, true
}
func queryIntDM(c *gin.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < 1 {
		return fallback
	}
	return v
}
func totalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}
