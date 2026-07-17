package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"twistgram-api-go/internal/model"
)

type PostOwnerInfo struct {
	UserID    uuid.UUID
	IsPrivate bool
}

type InteractionRepository interface {
	PostExists(ctx context.Context, id uuid.UUID) (bool, error)
	CommentExists(ctx context.Context, id uuid.UUID) (bool, error)
	FindCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	CreateComment(ctx context.Context, comment *model.Comment) error
	DeleteComment(ctx context.Context, id, userID uuid.UUID) error
	DeleteCommentAsPostOwner(ctx context.Context, commentID, postID uuid.UUID) error
	UpsertLike(ctx context.Context, like *model.Like) error
	DeleteLike(ctx context.Context, userID uuid.UUID, likeableType string, likeableID uuid.UUID) error
	SavedExists(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	UpsertSavedPost(ctx context.Context, saved *model.SavedPost) error
	DeleteSavedPost(ctx context.Context, userID, postID uuid.UUID) error
	ListPostComments(ctx context.Context, postID uuid.UUID, page, limit int) ([]model.Comment, int64, error)
	ListSavedPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.SavedPost, int64, error)
	IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error)
	GetPostOwner(ctx context.Context, postID uuid.UUID) (*PostOwnerInfo, error)
	IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetCommentPostID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
	CreateNotification(ctx context.Context, notification *model.Notification) error
}

type GormInteractionRepository struct{ db *gorm.DB }

func NewInteractionRepository(db *gorm.DB) InteractionRepository {
	return &GormInteractionRepository{db: db}
}

func (r *GormInteractionRepository) PostExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) CommentExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) FindCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.WithContext(ctx).First(&c, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormInteractionRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *GormInteractionRepository) DeleteComment(ctx context.Context, id, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Update("deleted_at", gorm.Expr("now()"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormInteractionRepository) DeleteCommentAsPostOwner(ctx context.Context, commentID, postID uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ? AND post_id = ? AND deleted_at IS NULL", commentID, postID).Update("deleted_at", gorm.Expr("now()"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormInteractionRepository) UpsertLike(ctx context.Context, like *model.Like) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "likeable_type"}, {Name: "likeable_id"}}, DoNothing: true}).Create(like).Error
}

func (r *GormInteractionRepository) DeleteLike(ctx context.Context, userID uuid.UUID, likeableType string, likeableID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND likeable_type = ? AND likeable_id = ?", userID, likeableType, likeableID).Delete(&model.Like{}).Error
}

func (r *GormInteractionRepository) SavedExists(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SavedPost{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) UpsertSavedPost(ctx context.Context, saved *model.SavedPost) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "post_id"}}, DoNothing: true}).Create(saved).Error
}

func (r *GormInteractionRepository) DeleteSavedPost(ctx context.Context, userID, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.SavedPost{}).Error
}

func (r *GormInteractionRepository) ListPostComments(ctx context.Context, postID uuid.UUID, page, limit int) ([]model.Comment, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ? AND deleted_at IS NULL", postID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var comments []model.Comment
	err := query.Order("created_at ASC").Offset((page - 1) * limit).Limit(limit).Find(&comments).Error
	return comments, total, err
}

func (r *GormInteractionRepository) ListSavedPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.SavedPost, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.SavedPost{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var saved []model.SavedPost
	err := query.Preload("Post").Preload("Post.Media").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&saved).Error
	return saved, total, err
}

func (r *GormInteractionRepository) IsBlockedEitherDirection(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Block{}).Where(
		"(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
		userA, userB, userB, userA,
	).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) GetPostOwner(ctx context.Context, postID uuid.UUID) (*PostOwnerInfo, error) {
	var result struct {
		UserID    uuid.UUID
		IsPrivate bool
	}
	err := r.db.WithContext(ctx).Table("posts").
		Select("posts.user_id, users.is_private").
		Joins("JOIN users ON users.id = posts.user_id").
		Where("posts.id = ? AND posts.deleted_at IS NULL", postID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	if result.UserID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &PostOwnerInfo{UserID: result.UserID, IsPrivate: result.IsPrivate}, nil
}

func (r *GormInteractionRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	return CreateNotificationHelper(ctx, r.db, notification)
}

func (r *GormInteractionRepository) IsAcceptedFollower(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").
		Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) GetCommentPostID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error) {
	var c model.Comment
	if err := r.db.WithContext(ctx).Select("post_id").First(&c, "id = ? AND deleted_at IS NULL", commentID).Error; err != nil {
		return uuid.Nil, err
	}
	return c.PostID, nil
}
