package gogi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(redisClient *redis.Client, key string) *RedisQueue {
	return &RedisQueue{
		client: redisClient,
		key:    key,
	}
}

func (q *RedisQueue) SendJob(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, q.key, data).Err()
}

func (q *RedisQueue) ReceiveJobs(ctx context.Context, handler func(Job) error) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := q.client.BLPop(ctx, 5*time.Second, q.key).Result()
				if err != nil {
					if !errors.Is(err, redis.Nil) {
						GetLogger().Debug(fmt.Sprintf("Redis BLPop error: %v", err))
					}
					continue
				}

				if len(result) < 2 {
					continue
				}

				var job Job
				if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
					GetLogger().Debug(fmt.Sprintf("Failed to unmarshal Redis job: %v", err))
					continue
				}
				if err := handler(job); err != nil {
					GetLogger().Debug(fmt.Sprintf("Redis job handler error: %v", err))
				}
			}
		}
	}()
	return nil
}
