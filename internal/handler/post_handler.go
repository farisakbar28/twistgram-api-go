package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type PostHandler struct{ postService *service.PostService }

func NewPostHandlerWithService(svc *service.PostService) *PostHandler {
	return &PostHandler{postService: svc}
}

func (h *PostHandler) Create(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	res, err := h.postService.CreatePost(userID, req)
	if err != nil {
		response.InternalError(c, "Failed to create post")
		return
	}
	response.Created(c, gin.H{"post": res})
}

func (h *PostHandler) Feed(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	pg := page(c)
	lm := limit(c)
	items, total, err := h.postService.Feed(userID, pg, lm)
	if err != nil {
		response.InternalError(c, "Failed to load feed")
		return
	}
	response.WithPagination(c, gin.H{"posts": items}, buildMeta(pg, lm, total))
}
func (h *PostHandler) MyPosts(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	pg := page(c)
	lm := limit(c)
	items, total, err := h.postService.MyPosts(userID, pg, lm)
	if err != nil {
		response.InternalError(c, "Failed to load posts")
		return
	}
	response.WithPagination(c, gin.H{"posts": items}, buildMeta(pg, lm, total))
}

func (h *PostHandler) UserPosts(c *gin.Context) {
	viewerID, ok := authUser(c)
	if !ok {
		return
	}
	targetID, ok := parsePostUserParam(c)
	if !ok {
		return
	}
	pg := page(c)
	lm := limit(c)
	items, total, err := h.postService.UserPosts(viewerID, targetID, pg, lm)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "User not found")
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not have access to this user's posts")
		} else {
			response.InternalError(c, "Failed to load user posts")
		}
		return
	}
	response.WithPagination(c, gin.H{"posts": items}, buildMeta(pg, lm, total))
}

func (h *PostHandler) EditCaption(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	id, ok := parsePostID(c)
	if !ok {
		return
	}
	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if err := h.postService.EditCaption(userID, id, req); err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			response.BadRequest(c, "Invalid caption")
		} else if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not own this post")
		} else {
			response.InternalError(c, "Failed to update post")
		}
		return
	}
	response.Success(c, gin.H{"updated": true})
}

func (h *PostHandler) Archive(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	id, ok := parsePostID(c)
	if !ok {
		return
	}
	if err := h.postService.Archive(userID, id, true); err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not own this post")
		} else {
			response.InternalError(c, "Failed to archive post")
		}
		return
	}
	response.Success(c, gin.H{"archived": true})
}
func (h *PostHandler) Unarchive(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	id, ok := parsePostID(c)
	if !ok {
		return
	}
	if err := h.postService.Archive(userID, id, false); err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not own this post")
		} else {
			response.InternalError(c, "Failed to unarchive post")
		}
		return
	}
	response.Success(c, gin.H{"unarchived": true})
}
func (h *PostHandler) Delete(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	id, ok := parsePostID(c)
	if !ok {
		return
	}
	if err := h.postService.Delete(userID, id); err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not own this post")
		} else {
			response.InternalError(c, "Failed to delete post")
		}
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *PostHandler) GetByID(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	postID, ok := parsePostID(c)
	if !ok {
		return
	}
	res, err := h.postService.GetByID(userID, postID)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else {
			response.InternalError(c, "Failed to load post")
		}
		return
	}
	response.Success(c, gin.H{"post": res})
}

func (h *PostHandler) RemoveTag(c *gin.Context) {
	userID, ok := authUser(c)
	if !ok {
		return
	}
	postID, ok := parsePostID(c)
	if !ok {
		return
	}
	taggedUserIDStr := c.Param("taggedUserId")
	taggedUserID, err := uuid.Parse(taggedUserIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid tagged user id")
		return
	}
	if err := h.postService.RemoveTag(userID, postID, taggedUserID); err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.NotFound(c, "Post not found")
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(c, "You do not own this post")
		} else {
			response.InternalError(c, "Failed to remove tag")
		}
		return
	}
	response.Success(c, gin.H{"removed": true})
}

func authUser(c *gin.Context) (uuid.UUID, bool) {
	id := middleware.GetUserID(c)
	if id == "" {
		response.Unauthorized(c, "User not authenticated")
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		response.Unauthorized(c, "Invalid authenticated user")
		return uuid.Nil, false
	}
	return parsed, true
}
func parsePostID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid post id")
		return uuid.Nil, false
	}
	return id, true
}
func parsePostUserParam(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("identifier"))
	if err != nil {
		response.BadRequest(c, "Invalid user id")
		return uuid.Nil, false
	}
	return id, true
}
func page(c *gin.Context) int {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if p < 1 {
		return 1
	}
	return p
}
func limit(c *gin.Context) int {
	l, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if l < 1 {
		return 20
	}
	if l > 100 {
		return 100
	}
	return l
}
func buildMeta(page, limit int, total int64) *response.Meta {
	totalPages := 0
	if limit > 0 && total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &response.Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}
