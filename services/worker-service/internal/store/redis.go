package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	taskQueueKey  = "worker:tasks"
	taskKeyPrefix = "task:"
	taskTTL       = time.Hour
)

type Task struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Status  string `json:"status"`
	Result  string `json:"result"`
}

type TaskStore struct {
	client *redis.Client
}

func NewTaskStore(client *redis.Client) *TaskStore {
	return &TaskStore{client: client}
}

func NewRedisClient(addr string, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func taskKey(id string) string {
	return taskKeyPrefix + id
}

func (s *TaskStore) Enqueue(ctx context.Context, task Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	pipe := s.client.TxPipeline()

	pipe.LPush(ctx, taskQueueKey, payload)
	pipe.HSet(ctx, taskKey(task.ID), map[string]any{
		"id":      task.ID,
		"type":    task.Type,
		"payload": task.Payload,
		"status":  "queued",
		"result":  "",
	})
	pipe.Expire(ctx, taskKey(task.ID), taskTTL)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *TaskStore) BlockingPop(ctx context.Context, timeout time.Duration) (*Task, error) {
	result, err := s.client.BRPop(ctx, timeout, taskQueueKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, nil
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("unmarshal task: %w", err)
	}

	return &task, nil
}

func (s *TaskStore) SetStatus(ctx context.Context, id string, status string, result string) error {
	pipe := s.client.TxPipeline()

	pipe.HSet(ctx, taskKey(id), map[string]any{
		"status": status,
		"result": result,
	})
	pipe.Expire(ctx, taskKey(id), taskTTL)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *TaskStore) Get(ctx context.Context, id string) (map[string]string, error) {
	return s.client.HGetAll(ctx, taskKey(id)).Result()
}
