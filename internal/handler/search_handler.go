package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type SearchHandler struct{ service *service.SearchService }

func NewSearchHandlerWithService(svc *service.SearchService) *SearchHandler {
	return &SearchHandler{service: svc}
}

func (h *SearchHandler) Search(c *gin.Context) {
	viewerID := middleware.GetUserID(c)
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	res, err := h.service.Search(c.Request.Context(), viewerID, query, limit)
	if err != nil {
		response.InternalError(c, "Failed to search")
		return
	}
	response.Success(c, res)
}

func (h *SearchHandler) HashtagPosts(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		response.BadRequest(c, "Hashtag tag is required")
		return
	}
	viewerID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	posts, total, err := h.service.GetHashtagPosts(c.Request.Context(), tag, viewerID, page, limit)
	if err != nil {
		response.InternalError(c, "Failed to load hashtag posts")
		return
	}
	totalPages := 0
	if limit > 0 && total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	response.WithPagination(c, gin.H{"posts": posts}, &response.Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages})
}
