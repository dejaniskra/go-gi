package gogi

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisJobQueue struct {
	client *redis.Client
	name   string
}

func NewRedisQueue(redisClient *RedisClient, name string) *RedisJobQueue {
	return &RedisJobQueue{
		client: redisClient.Writer, // always write to writer side
		name:   name,
	}
}

func (r *RedisJobQueue) SendJob(ctx context.Context, job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return r.client.LPush(ctx, r.name, data).Err()
}

func (r *RedisJobQueue) ReceiveJobs(ctx context.Context, handler func(*Job) error) error {
	for {
		result, err := r.client.RPop(ctx, r.name).Result()
		if err == redis.Nil {
			return nil
		} else if err != nil {
			return err
		}
		var job Job
		if err := json.Unmarshal([]byte(result), &job); err != nil {
			return err
		}
		if err := handler(&job); err != nil {
			return err
		}
	}
}
