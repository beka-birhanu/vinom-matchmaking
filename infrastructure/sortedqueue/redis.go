package sortedqueue

import (
	"context"
	"time"

	"github.com/beka-birhanu/vinom-matchmaking/service/i"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// RedisUniqueSortedQueue manages a sorted queue in Redis with TTL support.
type RedisUniqueSortedQueue struct {
	client *redis.Client
	locker *redsync.Redsync
	ttl    time.Duration
}

// NewRedisUniqueSortedQueue initializes a RedisSortedQueue with the provided Redis client and TTL.
func NewRedisUniqueSortedQueue(client *redis.Client, ttlSeconds int) (i.SortedQueue, error) {
	queue := &RedisUniqueSortedQueue{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
	pool := goredis.NewPool(client)
	queue.locker = redsync.New(pool)
	return queue, nil
}

// Enqueue adds a member to the sorted queue with a given score and sets expiration if necessary.
func (rsq *RedisUniqueSortedQueue) Enqueue(ctx context.Context, queueKey string, score float64, member string) error {
	_, err := rsq.client.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: member}).Result()
	if err != nil {
		return err
	}

	// Set expiration only if it's not already set
	ttl, err := rsq.client.TTL(ctx, queueKey).Result()
	if err == nil && ttl == -1 {
		_ = rsq.client.Expire(ctx, queueKey, rsq.ttl).Err()
	}

	return nil
}

// DequeTops removes and retrieves up to `amount` members with the lowest scores.
func (rsq *RedisUniqueSortedQueue) DequeTops(ctx context.Context, queueKey string, amount int64) ([]string, error) {
	mutex := rsq.locker.NewMutex(queueKey + ":match_lock")
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.Unlock()
	}()

	var members []string
	if rsq.client.ZCard(ctx, queueKey).Val() >= amount {
		for _, p := range rsq.client.ZPopMin(ctx, queueKey, amount).Val() {
			members = append(members, p.Member.(string))
		}
	}

	return members, nil
}

// Count returns the number of members in the sorted queue.
func (rsq *RedisUniqueSortedQueue) Count(ctx context.Context, queueKey string) (int64, error) {
	mutex := rsq.locker.NewMutex(queueKey + ":match_lock")
	if err := mutex.Lock(); err != nil {
		return 0, err
	}
	defer func() {
		_, _ = mutex.Unlock()
	}()

	return rsq.client.ZCard(ctx, queueKey).Val(), nil
}
