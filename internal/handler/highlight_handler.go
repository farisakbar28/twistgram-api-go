package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type HighlightHandler struct{ service *service.HighlightService }

func NewHighlightHandlerWithService(svc *service.HighlightService) *HighlightHandler {
	return &HighlightHandler{service: svc}
}

func (h *HighlightHandler) Create(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	var req dto.CreateHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.service.Create(c.Request.Context(), userID, req)
	if h.handleError(c, err) {
		return
	}
	response.Created(c, gin.H{"highlight": res})
}

func (h *HighlightHandler) List(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), userID)
	if h.handleError(c, err) {
		return
	}
	response.Success(c, gin.H{"highlights": items})
}

func (h *HighlightHandler) Update(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	highlightID, ok := parseHighlightID(c)
	if !ok {
		return
	}
	var req dto.UpdateHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleError(c, h.service.Update(c.Request.Context(), userID, highlightID, req)) {
		return
	}
	response.Success(c, gin.H{"updated": true})
}

func (h *HighlightHandler) Delete(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	highlightID, ok := parseHighlightID(c)
	if !ok {
		return
	}
	if h.handleError(c, h.service.Delete(c.Request.Context(), userID, highlightID)) {
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *HighlightHandler) AddStory(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	highlightID, ok := parseHighlightID(c)
	if !ok {
		return
	}
	var req dto.AddStoryToHighlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if h.handleError(c, h.service.AddStory(c.Request.Context(), userID, highlightID, req)) {
		return
	}
	response.Success(c, gin.H{"added": true})
}

func (h *HighlightHandler) RemoveStory(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	highlightID, ok := parseHighlightID(c)
	if !ok {
		return
	}
	storyID, ok := parseStoryID(c)
	if !ok {
		return
	}
	if h.handleError(c, h.service.RemoveStory(c.Request.Context(), userID, highlightID, storyID)) {
		return
	}
	response.Success(c, gin.H{"removed": true})
}

func (h *HighlightHandler) handleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrHighlightNotFound):
		response.NotFound(c, "Highlight not found")
	case errors.Is(err, service.ErrForbidden):
		response.Forbidden(c, "You do not own this highlight")
	default:
		response.InternalError(c, "Failed to process highlight request")
	}
	return true
}

func parseHighlightID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid highlight id")
		return uuid.Nil, false
	}
	return id, true
}
