package repository

import (
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
	PostExists(id uuid.UUID) (bool, error)
	CommentExists(id uuid.UUID) (bool, error)
	FindCommentByID(id uuid.UUID) (*model.Comment, error)
	CreateComment(comment *model.Comment) error
	DeleteComment(id, userID uuid.UUID) error
	DeleteCommentAsPostOwner(commentID, postID uuid.UUID) error
	UpsertLike(like *model.Like) error
	DeleteLike(userID uuid.UUID, likeableType string, likeableID uuid.UUID) error
	SavedExists(userID, postID uuid.UUID) (bool, error)
	UpsertSavedPost(saved *model.SavedPost) error
	DeleteSavedPost(userID, postID uuid.UUID) error
	ListPostComments(postID uuid.UUID, page, limit int) ([]model.Comment, int64, error)
	ListSavedPosts(userID uuid.UUID, page, limit int) ([]model.SavedPost, int64, error)
	IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error)
	GetPostOwner(postID uuid.UUID) (*PostOwnerInfo, error)
	IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error)
	GetCommentPostID(commentID uuid.UUID) (uuid.UUID, error)
	CreateNotification(notification *model.Notification) error
}

type GormInteractionRepository struct{ db *gorm.DB }

func NewInteractionRepository(db *gorm.DB) InteractionRepository { return &GormInteractionRepository{db: db} }

func (r *GormInteractionRepository) PostExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) CommentExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) FindCommentByID(id uuid.UUID) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.First(&c, "id = ? AND deleted_at IS NULL", id).Error; err != nil { return nil, err }
	return &c, nil
}

func (r *GormInteractionRepository) CreateComment(comment *model.Comment) error { return r.db.Create(comment).Error }

func (r *GormInteractionRepository) DeleteComment(id, userID uuid.UUID) error {
	res := r.db.Model(&model.Comment{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).Update("deleted_at", gorm.Expr("now()"))
	if res.Error != nil { return res.Error }
	if res.RowsAffected == 0 { return gorm.ErrRecordNotFound }
	return nil
}

func (r *GormInteractionRepository) DeleteCommentAsPostOwner(commentID, postID uuid.UUID) error {
	res := r.db.Model(&model.Comment{}).Where("id = ? AND post_id = ? AND deleted_at IS NULL", commentID, postID).Update("deleted_at", gorm.Expr("now()"))
	if res.Error != nil { return res.Error }
	if res.RowsAffected == 0 { return gorm.ErrRecordNotFound }
	return nil
}

func (r *GormInteractionRepository) UpsertLike(like *model.Like) error {
	return r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "likeable_type"}, {Name: "likeable_id"}}, DoNothing: true}).Create(like).Error
}

func (r *GormInteractionRepository) DeleteLike(userID uuid.UUID, likeableType string, likeableID uuid.UUID) error {
	return r.db.Where("user_id = ? AND likeable_type = ? AND likeable_id = ?", userID, likeableType, likeableID).Delete(&model.Like{}).Error
}

func (r *GormInteractionRepository) SavedExists(userID, postID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.SavedPost{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) UpsertSavedPost(saved *model.SavedPost) error {
	return r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "post_id"}}, DoNothing: true}).Create(saved).Error
}

func (r *GormInteractionRepository) DeleteSavedPost(userID, postID uuid.UUID) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.SavedPost{}).Error
}

func (r *GormInteractionRepository) ListPostComments(postID uuid.UUID, page, limit int) ([]model.Comment, int64, error) {
	var total int64
	query := r.db.Model(&model.Comment{}).Where("post_id = ? AND deleted_at IS NULL", postID)
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var comments []model.Comment
	err := query.Order("created_at ASC").Offset((page-1)*limit).Limit(limit).Find(&comments).Error
	return comments, total, err
}

func (r *GormInteractionRepository) ListSavedPosts(userID uuid.UUID, page, limit int) ([]model.SavedPost, int64, error) {
	var total int64
	query := r.db.Model(&model.SavedPost{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	var saved []model.SavedPost
	err := query.Preload("Post").Preload("Post.Media").Order("created_at DESC").Offset((page-1)*limit).Limit(limit).Find(&saved).Error
	return saved, total, err
}

func (r *GormInteractionRepository) IsBlockedEitherDirection(userA, userB uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Block{}).Where(
		"(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
		userA, userB, userB, userA,
	).Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) GetPostOwner(postID uuid.UUID) (*PostOwnerInfo, error) {
	var result struct {
		UserID    uuid.UUID
		IsPrivate bool
	}
	err := r.db.Table("posts").
		Select("posts.user_id, users.is_private").
		Joins("JOIN users ON users.id = posts.user_id").
		Where("posts.id = ? AND posts.deleted_at IS NULL", postID).
		Scan(&result).Error
	if err != nil { return nil, err }
	if result.UserID == uuid.Nil { return nil, gorm.ErrRecordNotFound }
	return &PostOwnerInfo{UserID: result.UserID, IsPrivate: result.IsPrivate}, nil
}

func (r *GormInteractionRepository) CreateNotification(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

func (r *GormInteractionRepository) IsAcceptedFollower(followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ? AND status = ?", followerID, followingID, "accepted").
		Count(&count).Error
	return count > 0, err
}

func (r *GormInteractionRepository) GetCommentPostID(commentID uuid.UUID) (uuid.UUID, error) {
	var c model.Comment
	if err := r.db.Select("post_id").First(&c, "id = ? AND deleted_at IS NULL", commentID).Error; err != nil {
		return uuid.Nil, err
	}
	return c.PostID, nil
}
