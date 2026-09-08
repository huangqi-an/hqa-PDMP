package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	taskKeyPrefix = "task:"
	taskTTL       = time.Hour
	taskStreamKey = "worker:task-stream"
	taskGroupName = "worker-group"
)

type Task struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Status  string `json:"status"`
	Result  string `json:"result"`
}

type StreamTask struct {
	Task     Task
	StreamID string
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

	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: taskStreamKey,
		Values: map[string]any{
			"task": payload,
		},
	})
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

func (s *TaskStore) EnsureGroup(ctx context.Context) error {
	err := s.client.XGroupCreateMkStream(
		ctx,
		taskStreamKey,
		taskGroupName,
		"$",
	).Err()

	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil
}

func (s *TaskStore) ReadGroup(
	ctx context.Context,
	consumer string,
	timeout time.Duration,
) (*StreamTask, error) {
	streams, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    taskGroupName,
		Consumer: consumer,
		Streams:  []string{taskStreamKey, ">"},
		Count:    1,
		Block:    timeout,
	}).Result()

	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil
	}

	msg := streams[0].Messages[0]

	payload, ok := msg.Values["task"].(string)
	if !ok {
		return nil, fmt.Errorf("stream task payload is not string")
	}

	var task Task
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return nil, fmt.Errorf("unmarshal stream task: %w", err)
	}

	return &StreamTask{
		Task:     task,
		StreamID: msg.ID,
	}, nil
}

func (s *TaskStore) Ack(ctx context.Context, streamID string) error {
	return s.client.XAck(
		ctx,
		taskStreamKey,
		taskGroupName,
		streamID,
	).Err()
}
