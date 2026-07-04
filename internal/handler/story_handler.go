package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type StoryHandler struct{ service *service.StoryService }

func NewStoryHandler() *StoryHandler {
	repo := repository.NewStoryRepository(config.GetDB())
	return &StoryHandler{service: service.NewStoryService(repo)}
}

func (h *StoryHandler) Create(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	var req dto.CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	res, err := h.service.Create(userID, req)
	if h.handleError(c, err) { return }
	response.Created(c, gin.H{"story": res})
}

func (h *StoryHandler) Delete(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	storyID, ok := parseStoryID(c); if !ok { return }
	if h.handleError(c, h.service.Delete(userID, storyID)) { return }
	response.Success(c, gin.H{"deleted": true})
}

func (h *StoryHandler) GetByID(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	storyID, ok := parseStoryID(c); if !ok { return }
	res, err := h.service.GetByID(userID, storyID)
	if h.handleError(c, err) { return }
	response.Success(c, gin.H{"story": res})
}

func (h *StoryHandler) Feed(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	items, err := h.service.Feed(userID)
	if h.handleError(c, err) { return }
	response.Success(c, gin.H{"feed": items})
}

func (h *StoryHandler) RecordView(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	storyID, ok := parseStoryID(c); if !ok { return }
	if h.handleError(c, h.service.RecordView(userID, storyID)) { return }
	response.Success(c, gin.H{"viewed": true})
}

func (h *StoryHandler) Viewers(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	storyID, ok := parseStoryID(c); if !ok { return }
	items, err := h.service.Viewers(userID, storyID)
	if h.handleError(c, err) { return }
	response.Success(c, gin.H{"viewers": items})
}

func (h *StoryHandler) handleError(c *gin.Context, err error) bool {
	if err == nil { return false }
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrStoryNotFound):
		response.NotFound(c, "Story not found")
	case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrUserBlocked):
		response.Forbidden(c, "You do not have access to this story")
	default:
		response.InternalError(c, "Failed to process story request")
	}
	return true
}

func parseStoryID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil { response.BadRequest(c, "Invalid story id"); return uuid.Nil, false }
	return id, true
}
