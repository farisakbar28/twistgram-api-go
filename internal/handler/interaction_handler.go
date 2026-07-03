package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type InteractionHandler struct{ service *service.InteractionService }

func NewInteractionHandler() *InteractionHandler {
	repo := repository.NewInteractionRepository(config.GetDB())
	return &InteractionHandler{service: service.NewInteractionService(repo)}
}

func (h *InteractionHandler) LikePost(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	postID, ok := parsePostID(c); if !ok { return }
	res, err := h.service.LikePost(userID, postID)
	if h.handleError(c, err) { return }
	response.Success(c, gin.H{"like": res})
}

func (h *InteractionHandler) UnlikePost(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	postID, ok := parsePostID(c); if !ok { return }
	if h.handleError(c, h.service.UnlikePost(userID, postID)) { return }
	response.Success(c, gin.H{"unliked": true})
}

func (h *InteractionHandler) Comment(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	postID, ok := parsePostID(c); if !ok { return }
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "Invalid request body"); return }
	res, err := h.service.CreateComment(userID, postID, req)
	if h.handleError(c, err) { return }
	response.Created(c, gin.H{"comment": res})
}

func (h *InteractionHandler) DeleteComment(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	commentID, ok := parseCommentParam(c); if !ok { return }
	if h.handleError(c, h.service.DeleteComment(userID, commentID)) { return }
	response.Success(c, gin.H{"deleted": true})
}

func (h *InteractionHandler) LikeComment(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	commentID, ok := parseCommentParam(c); if !ok { return }
	res, err := h.service.LikeComment(userID, commentID)
	if h.handleError(c, err) { return }
	response.Success(c, gin.H{"like": res})
}

func (h *InteractionHandler) SavePost(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	postID, ok := parsePostID(c); if !ok { return }
	res, err := h.service.SavePost(userID, postID)
	if h.handleError(c, err) { return }
	response.Success(c, gin.H{"save": res})
}

func (h *InteractionHandler) UnsavePost(c *gin.Context) {
	userID, ok := authUser(c); if !ok { return }
	postID, ok := parsePostID(c); if !ok { return }
	if h.handleError(c, h.service.UnsavePost(userID, postID)) { return }
	response.Success(c, gin.H{"unsaved": true})
}

func (h *InteractionHandler) handleError(c *gin.Context, err error) bool {
	if err == nil { return false }
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrPostNotFound), errors.Is(err, service.ErrInteractionNotFound):
		response.NotFound(c, "Resource not found")
	case errors.Is(err, service.ErrForbidden):
		response.Forbidden(c, "You do not have access to this resource")
	case errors.Is(err, service.ErrUserBlocked):
		response.Forbidden(c, "Action blocked")
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.NotFound(c, "Resource not found")
	default:
		response.InternalError(c, "Failed to process interaction request")
	}
	return true
}

func parseCommentParam(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("comment_id"))
	if err != nil { response.BadRequest(c, "Invalid comment id"); return uuid.Nil, false }
	return id, true
}

func parseCommentID(c *gin.Context) (uuid.UUID, bool) { id, err := uuid.Parse(c.Param("id")); if err != nil { response.BadRequest(c, "Invalid comment id"); return uuid.Nil, false }; return id, true }

func _page(c *gin.Context) int { p, _ := strconv.Atoi(c.DefaultQuery("page", "1")); if p < 1 { return 1 }; return p }
