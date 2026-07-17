package dto

import "time"

type PostMediaRequest struct {
	MediaURL      string  `json:"media_url" binding:"required"`
	MediaType     string  `json:"media_type" binding:"required"`
	OrderIndex    int     `json:"order_index"`
	MusicTrackURL *string `json:"music_track_url"`
}

type CreatePostRequest struct {
	Caption          *string            `json:"caption"`
	Location         *string            `json:"location"`
	CommentsDisabled *bool              `json:"comments_disabled"`
	Media            []PostMediaRequest `json:"media" binding:"required,min=1,max=10"`
	TaggedUserIDs    []string           `json:"tagged_user_ids"`
	Hashtags         []string           `json:"hashtags"`
}

type UpdatePostRequest struct {
	Caption *string `json:"caption"`
}

type PostMediaResponse struct {
	MediaID       string  `json:"media_id,omitempty"`
	Type          string  `json:"type"`
	URL           string  `json:"url"`
	AspectRatio   *string `json:"aspect_ratio,omitempty"`
}

type PostUserResponse struct {
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	AvatarURL  *string `json:"avatar_url,omitempty"`
	IsVerified bool    `json:"is_verified"`
}

type PostMetricsResponse struct {
	LikesCount    int  `json:"likes_count"`
	CommentsCount int  `json:"comments_count"`
	HasLiked      bool `json:"has_liked"`
	HasSaved      bool `json:"has_saved"`
}

type PostResponse struct {
	PostID           string                `json:"post_id"`
	PostType         string                `json:"post_type"`
	User             PostUserResponse      `json:"user"`
	Caption          *string               `json:"caption,omitempty"`
	Location         *string               `json:"location,omitempty"`
	IsArchived       bool                  `json:"is_archived,omitempty"`
	CommentsDisabled bool                  `json:"comments_disabled,omitempty"`
	MediaURL         *string               `json:"media_url,omitempty"`    
	AspectRatio      *string               `json:"aspect_ratio,omitempty"` 
	MediaItems       []PostMediaResponse   `json:"media_items,omitempty"`  
	Metrics          PostMetricsResponse   `json:"metrics"`
	CreatedAt        time.Time             `json:"created_at"`
}

type FeedResponse struct {
	Status string `json:"status"`
	Data   struct {
		Posts      []PostResponse `json:"posts"`
		Pagination struct {
			NextCursor string `json:"next_cursor,omitempty"`
			HasMore    bool   `json:"has_more"`
		} `json:"pagination"`
	} `json:"data"`
}
