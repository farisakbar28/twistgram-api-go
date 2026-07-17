package constants

// Follow status values
const (
	StatusAccepted = "accepted"
	StatusPending  = "pending"
)

// Likeable types
const (
	LikeablePost    = "post"
	LikeableComment = "comment"
)

// Notification types
const (
	NotifTypeLike          = "like"
	NotifTypeComment       = "comment"
	NotifTypeFollow        = "follow"
	NotifTypeFollowRequest = "follow_request"
	NotifTypeMention       = "mention"
	NotifTypeStoryReply    = "story_reply"
)

// OTP types
const (
	OTPSignup        = "signup"
	OTPRecovery      = "recovery"
	OTPEmailRecovery = "email_recovery"
)

// Report target types
const (
	ReportTargetUser    = "user"
	ReportTargetPost    = "post"
	ReportTargetComment = "comment"
)

// Report reasons
const (
	ReasonSpam         = "spam"
	ReasonInappropriate = "inappropriate"
	ReasonHarassment   = "harassment"
	ReasonFakeAccount  = "fake_account"
	ReasonOther        = "other"
)

// Report status
const (
	ReportStatusPending     = "pending"
	ReportStatusReviewed    = "reviewed"
	ReportStatusActionTaken = "action_taken"
	ReportStatusDismissed   = "dismissed"
)

// Media types
const (
	MediaTypeImage = "image"
	MediaTypeVideo = "video"
	MediaTypeText  = "text"
)

// Story visibility
const (
	VisibilityAllFollowers = "all_followers"
	VisibilityCloseFriends = "close_friends"
)

// Collection names for saved posts
const (
	DefaultCollection = "All"
)

// Password validation
const (
	MinPasswordLength = 8
)

// Pagination defaults
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
	MaxSearchLimit = 50
)

// Rate limits
const (
	RateLimitPublic  = 5
	RateBurstPublic  = 10
	RateLimitProtected = 30
	RateBurstProtected = 60
)
