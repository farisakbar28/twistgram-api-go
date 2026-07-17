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

type SocialHandler struct {
	socialService *service.SocialService
}

func NewSocialHandlerWithService(socialService *service.SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

func (h *SocialHandler) Follow(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	targetID, ok := parseIDParam(c)
	if !ok {
		return
	}
	result, err := h.socialService.Follow(c.Request.Context(), viewerID, targetID)
	if h.handleSocialError(c, err) {
		return
	}
	response.Success(c, gin.H{"follow": result})
}

func (h *SocialHandler) Unfollow(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	targetID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.Unfollow(c.Request.Context(), viewerID, targetID)) {
		return
	}
	response.Success(c, gin.H{"unfollowed": true})
}

func (h *SocialHandler) Followers(c *gin.Context) {
	userID, ok := parseIDParam(c)
	if !ok {
		return
	}
	items, meta, err := h.socialService.ListFollowers(c.Request.Context(), userID, queryInt(c, "page", 1), queryInt(c, "limit", 20))
	if h.handleSocialError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"followers": items}, meta)
}

func (h *SocialHandler) Following(c *gin.Context) {
	userID, ok := parseIDParam(c)
	if !ok {
		return
	}
	items, meta, err := h.socialService.ListFollowing(c.Request.Context(), userID, queryInt(c, "page", 1), queryInt(c, "limit", 20))
	if h.handleSocialError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"following": items}, meta)
}

func (h *SocialHandler) RemoveFollower(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	followerID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.RemoveFollower(c.Request.Context(), viewerID, followerID)) {
		return
	}
	response.Success(c, gin.H{"removed": true})
}

func (h *SocialHandler) FollowRequests(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	items, meta, err := h.socialService.ListIncomingFollowRequests(c.Request.Context(), viewerID, queryInt(c, "page", 1), queryInt(c, "limit", 20))
	if h.handleSocialError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"requests": items}, meta)
}

func (h *SocialHandler) ApproveFollowRequest(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	requesterID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.ApproveFollowRequest(c.Request.Context(), viewerID, requesterID)) {
		return
	}
	response.Success(c, gin.H{"approved": true})
}

func (h *SocialHandler) DeclineFollowRequest(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	requesterID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.DeclineFollowRequest(c.Request.Context(), viewerID, requesterID)) {
		return
	}
	response.Success(c, gin.H{"declined": true})
}

func (h *SocialHandler) Block(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	targetID, ok := parseIDParam(c)
	if !ok {
		return
	}
	result, err := h.socialService.Block(c.Request.Context(), viewerID, targetID)
	if h.handleSocialError(c, err) {
		return
	}
	response.Success(c, gin.H{"block": result})
}

func (h *SocialHandler) Unblock(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	targetID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.Unblock(c.Request.Context(), viewerID, targetID)) {
		return
	}
	response.Success(c, gin.H{"unblocked": true})
}

func (h *SocialHandler) GetBlockedUsers(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	users, err := h.socialService.GetBlockedUsers(c.Request.Context(), viewerID)
	if h.handleSocialError(c, err) {
		return
	}
	response.Success(c, gin.H{"users": users})
}

func (h *SocialHandler) Report(c *gin.Context) {
	reporterID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	var req dto.ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	result, err := h.socialService.Report(c.Request.Context(), reporterID, req)
	if h.handleSocialError(c, err) {
		return
	}
	response.Created(c, gin.H{"report": result})
}

func (h *SocialHandler) AddCloseFriend(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	targetID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.AddCloseFriend(c.Request.Context(), viewerID, targetID)) {
		return
	}
	response.Success(c, gin.H{"close_friend": true})
}

func (h *SocialHandler) RemoveCloseFriend(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	targetID, ok := parseIDParam(c)
	if !ok {
		return
	}
	if h.handleSocialError(c, h.socialService.RemoveCloseFriend(c.Request.Context(), viewerID, targetID)) {
		return
	}
	response.Success(c, gin.H{"close_friend": false})
}

func (h *SocialHandler) ListCloseFriends(c *gin.Context) {
	viewerID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}
	items, meta, err := h.socialService.ListCloseFriends(c.Request.Context(), viewerID, queryInt(c, "page", 1), queryInt(c, "limit", 20))
	if h.handleSocialError(c, err) {
		return
	}
	response.WithPagination(c, gin.H{"close_friends": items}, meta)
}

func parseIDParam(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("identifier"))
	if err != nil {
		response.BadRequest(c, "Invalid user id")
		return uuid.Nil, false
	}
	return id, true
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func (h *SocialHandler) handleSocialError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, "Invalid request data")
	case errors.Is(err, service.ErrSelfAction):
		response.BadRequest(c, "Cannot perform this action on yourself")
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, "User not found")
	case errors.Is(err, service.ErrTargetNotFound):
		response.NotFound(c, "Report target not found")
	case errors.Is(err, service.ErrInvalidTarget):
		response.BadRequest(c, "Invalid report target type")
	case errors.Is(err, service.ErrInvalidReason):
		response.BadRequest(c, "Invalid report reason")
	case errors.Is(err, service.ErrBlocked):
		response.Forbidden(c, "Action blocked by user privacy settings")
	case errors.Is(err, service.ErrFollowNotFound):
		response.NotFound(c, "Follow request not found")
	default:
		response.InternalError(c, "Failed to process social request")
	}
	return true
}
