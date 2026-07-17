package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"twistgram-api-go/internal/dto"
	"twistgram-api-go/internal/model"
	"twistgram-api-go/internal/repository"
)

var ErrStoryNotFound = errors.New("story not found")

const storyTTL = 24 * time.Hour

type StoryService struct{ repo repository.StoryRepository }

func NewStoryService(repo repository.StoryRepository) *StoryService { return &StoryService{repo: repo} }

func (s *StoryService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateStoryRequest) (*dto.StoryResponse, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	mt := strings.TrimSpace(strings.ToLower(req.MediaType))
	if mt != "image" && mt != "video" && mt != "text" {
		return nil, ErrInvalidInput
	}
	if mt == "text" && (req.TextContent == nil || strings.TrimSpace(*req.TextContent) == "") {
		return nil, ErrInvalidInput
	}
	if (mt == "image" || mt == "video") && (req.MediaURL == nil || strings.TrimSpace(*req.MediaURL) == "") {
		return nil, ErrInvalidInput
	}
	story := &model.Story{
		UserID:        userID,
		MediaURL:      req.MediaURL,
		MediaType:     mt,
		TextContent:   req.TextContent,
		MusicTrackURL: req.MusicTrackURL,
		ExpiresAt:     time.Now().Add(storyTTL),
	}
	tags := make([]model.StoryTag, 0)
	for _, idStr := range req.TaggedUserIDs {
		parsed, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil || parsed == userID {
			continue
		}
		tags = append(tags, model.StoryTag{TaggedUserID: parsed})
	}
	if err := s.repo.CreateStoryWithTags(ctx, story, tags); err != nil {
		return nil, err
	}

	for _, tag := range tags {
		_ = s.repo.CreateNotification(ctx, &model.Notification{
			RecipientID: tag.TaggedUserID,
			ActorID:     userID,
			Type:        "mention",
			ReferenceID: &story.ID,
		})
	}

	return buildStoryResponse(story), nil
}

func (s *StoryService) GetByID(ctx context.Context, viewerID, storyID uuid.UUID) (*dto.StoryResponse, error) {
	story, err := s.repo.GetStoryByID(ctx, storyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStoryNotFound
	}
	if err != nil {
		return nil, err
	}
	if viewerID != story.UserID {
		blocked, err := s.repo.IsBlockedEitherDirection(ctx, viewerID, story.UserID)
		if err != nil {
			return nil, err
		}
		if blocked {
			return nil, ErrForbidden
		}
		following, err := s.repo.IsAcceptedFollower(ctx, viewerID, story.UserID)
		if err != nil {
			return nil, err
		}
		if !following {
			return nil, ErrForbidden
		}
	}
	return buildStoryResponse(story), nil
}

func (s *StoryService) Delete(ctx context.Context, userID, storyID uuid.UUID) error {
	if userID == uuid.Nil || storyID == uuid.Nil {
		return ErrInvalidInput
	}
	ownerID, err := s.repo.GetStoryOwner(ctx, storyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrStoryNotFound
	}
	if err != nil {
		return err
	}
	if ownerID != userID {
		return ErrForbidden
	}

	if err := s.repo.DeleteStory(ctx, storyID, userID); err != nil {
		return err
	}
	return nil
}

func (s *StoryService) Feed(ctx context.Context, userID uuid.UUID) ([]dto.StoryFeedItem, error) {
	stories, err := s.repo.ListActiveFeedStories(ctx, userID)
	if err != nil {
		return nil, err
	}
	grouped := map[uuid.UUID]*dto.StoryFeedItem{}
	order := []uuid.UUID{}
	for _, story := range stories {
		if _, ok := grouped[story.UserID]; !ok {
			grouped[story.UserID] = &dto.StoryFeedItem{UserID: story.UserID.String(), Username: story.User.Username, AvatarURL: story.User.AvatarURL, Stories: []dto.StoryResponse{}}
			order = append(order, story.UserID)
		}
		grouped[story.UserID].Stories = append(grouped[story.UserID].Stories, *buildStoryResponse(&story))
	}
	out := make([]dto.StoryFeedItem, 0, len(order))
	for _, uid := range order {
		item := grouped[uid]
		st := item.Stories
		for i, j := 0, len(st)-1; i < j; i, j = i+1, j-1 {
			st[i], st[j] = st[j], st[i]
		}
		out = append(out, *item)
	}
	return out, nil
}

func (s *StoryService) RecordView(ctx context.Context, viewerID, storyID uuid.UUID) error {
	if viewerID == uuid.Nil || storyID == uuid.Nil {
		return ErrInvalidInput
	}
	story, err := s.repo.GetStoryByID(ctx, storyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrStoryNotFound
	}
	if err != nil {
		return err
	}
	if viewerID == story.UserID {
		return nil
	}
	blocked, err := s.repo.IsBlockedEitherDirection(ctx, viewerID, story.UserID)
	if err != nil {
		return err
	}
	if blocked {
		return ErrForbidden
	}
	return s.repo.RecordView(ctx, &model.StoryView{StoryID: storyID, ViewerID: viewerID, ViewedAt: time.Now()})
}

func (s *StoryService) Viewers(ctx context.Context, userID, storyID uuid.UUID) ([]dto.StoryViewerResponse, error) {
	ownerID, err := s.repo.GetStoryOwner(ctx, storyID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStoryNotFound
	}
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, ErrForbidden
	}
	views, err := s.repo.ListViewers(ctx, storyID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.StoryViewerResponse, 0, len(views))
	for _, v := range views {
		out = append(out, dto.StoryViewerResponse{ViewerID: v.ViewerID.String(), Name: v.Viewer.Name, Username: v.Viewer.Username, AvatarURL: v.Viewer.AvatarURL, ViewedAt: v.ViewedAt})
	}
	return out, nil
}

func buildStoryResponse(s *model.Story) *dto.StoryResponse {
	return &dto.StoryResponse{ID: s.ID.String(), UserID: s.UserID.String(), MediaURL: s.MediaURL, MediaType: s.MediaType, TextContent: s.TextContent, MusicTrackURL: s.MusicTrackURL, ExpiresAt: s.ExpiresAt, CreatedAt: s.CreatedAt}
}
