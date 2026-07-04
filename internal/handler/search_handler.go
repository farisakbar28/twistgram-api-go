package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"twistgram-api-go/internal/config"
	"twistgram-api-go/internal/middleware"
	"twistgram-api-go/internal/repository"
	"twistgram-api-go/internal/service"
	"twistgram-api-go/pkg/response"
)

type SearchHandler struct{ service *service.SearchService }

func NewSearchHandler() *SearchHandler {
	repo := repository.NewSearchRepository(config.GetDB())
	return &SearchHandler{service: service.NewSearchService(repo)}
}

func (h *SearchHandler) Search(c *gin.Context) {
	viewerID := middleware.GetUserID(c)
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	res, err := h.service.Search(viewerID, query, limit)
	if err != nil { response.InternalError(c, "Failed to search"); return }
	response.Success(c, res)
}
