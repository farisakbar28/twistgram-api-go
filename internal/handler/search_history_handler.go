package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type SearchHistoryHandler struct{ service *service.SearchHistoryService }

func NewSearchHistoryHandlerWithService(svc *service.SearchHistoryService) *SearchHistoryHandler {
	return &SearchHistoryHandler{service: svc}
}

func (h *SearchHistoryHandler) List(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.service.ListHistory(c.Request.Context(), userID, limit)
	if err != nil {
		response.InternalError(c, "Failed to load search history")
		return
	}
	response.Success(c, gin.H{"history": items})
}

func (h *SearchHistoryHandler) Save(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	query := c.Query("q")
	queryType := c.DefaultQuery("type", "user")
	if err := h.service.SaveSearch(c.Request.Context(), userID, query, queryType); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}
	response.Success(c, gin.H{"saved": true})
}

func (h *SearchHistoryHandler) DeleteItem(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid history id")
		return
	}
	if err := h.service.DeleteHistoryItem(c.Request.Context(), userID, itemID); err != nil {
		response.InternalError(c, "Failed to delete history item")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *SearchHistoryHandler) DeleteAll(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	if err := h.service.DeleteAllHistory(c.Request.Context(), userID); err != nil {
		response.InternalError(c, "Failed to delete history")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
