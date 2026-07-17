package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisTimelineRepository interface {
	PushToFollowers(ctx context.Context, postID uuid.UUID, timestamp int64, followerIDs []uuid.UUID) error
	PushToCreator(ctx context.Context, postID uuid.UUID, timestamp int64, creatorID uuid.UUID) error
	GetTimeline(ctx context.Context, userID uuid.UUID, limit int) ([]string, error)
	GetCreatorTimeline(ctx context.Context, creatorID uuid.UUID, limit int) ([]string, error)
}

type DefaultRedisTimelineRepo struct{ client *redis.Client }

func NewRedisTimelineRepository(client *redis.Client) RedisTimelineRepository {
	return &DefaultRedisTimelineRepo{client: client}
}

func (r *DefaultRedisTimelineRepo) PushToFollowers(ctx context.Context, postID uuid.UUID, timestamp int64, followerIDs []uuid.UUID) error {
	pipe := r.client.Pipeline()
	postIDStr := postID.String()
	for _, fID := range followerIDs {
		key := fmt.Sprintf("timeline:fanout:%s", fID.String())
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(timestamp), Member: postIDStr})
		pipe.ZRemRangeByRank(ctx, key, 0, -801)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *DefaultRedisTimelineRepo) PushToCreator(ctx context.Context, postID uuid.UUID, timestamp int64, creatorID uuid.UUID) error {
	key := fmt.Sprintf("timeline:creator:%s", creatorID.String())
	err := r.client.ZAdd(ctx, key, redis.Z{Score: float64(timestamp), Member: postID.String()}).Err()
	if err == nil {
		r.client.ZRemRangeByRank(ctx, key, 0, -801)
	}
	return err
}

func (r *DefaultRedisTimelineRepo) GetTimeline(ctx context.Context, userID uuid.UUID, limit int) ([]string, error) {
	key := fmt.Sprintf("timeline:fanout:%s", userID.String())
	return r.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
}

func (r *DefaultRedisTimelineRepo) GetCreatorTimeline(ctx context.Context, creatorID uuid.UUID, limit int) ([]string, error) {
	key := fmt.Sprintf("timeline:creator:%s", creatorID.String())
	return r.client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
}
