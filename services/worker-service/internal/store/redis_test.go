package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestTaskStore(t *testing.T) (*TaskStore, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})

	return NewTaskStore(client), mr
}

func TestEnqueueWritesTaskToRedis(t *testing.T) {
	store, _ := newTestTaskStore(t)
	ctx := context.Background()

	task := Task{
		ID:      "task-1",
		Type:    "echo",
		Payload: "hello redis",
		Status:  "queued",
	}

	if err := store.Enqueue(ctx, task); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	data, err := store.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if data["status"] != "queued" {
		t.Fatalf("status = %q, want queued", data["status"])
	}

	if data["payload"] != "hello redis" {
		t.Fatalf("payload = %q, want hello redis", data["payload"])
	}

	queueLength := store.client.LLen(ctx, taskQueueKey).Val()
	if queueLength != 1 {
		t.Fatalf("queue length = %d, want 1", queueLength)
	}
}

func TestBlockingPopReturnsAndRemovesTask(t *testing.T) {
	store, _ := newTestTaskStore(t)
	ctx := context.Background()

	task := Task{
		ID:      "task-2",
		Type:    "echo",
		Payload: "second",
		Status:  "queued",
	}

	if err := store.Enqueue(ctx, task); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	popCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	got, err := store.BlockingPop(popCtx, time.Second)
	if err != nil {
		t.Fatalf("BlockingPop() error = %v", err)
	}

	if got == nil {
		t.Fatal("BlockingPop() = nil, want task")
	}

	if got.ID != "task-2" {
		t.Fatalf("got ID = %q, want task-2", got.ID)
	}

	queueLength := store.client.LLen(ctx, taskQueueKey).Val()
	if queueLength != 0 {
		t.Fatalf("queue length = %d, want 0", queueLength)
	}
}

func TestBlockingPopReturnsNilWhenEmpty(t *testing.T) {
	store, _ := newTestTaskStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	got, err := store.BlockingPop(ctx, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("BlockingPop() error = %v", err)
	}

	if got != nil {
		t.Fatalf("BlockingPop() = %#v, want nil", got)
	}
}

func TestSetStatusUpdatesStatusAndResult(t *testing.T) {
	store, _ := newTestTaskStore(t)
	ctx := context.Background()

	task := Task{
		ID:      "task-3",
		Type:    "echo",
		Payload: "third",
		Status:  "queued",
	}

	if err := store.Enqueue(ctx, task); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	if err := store.SetStatus(ctx, task.ID, "done", "processed: third"); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	data, err := store.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if data["status"] != "done" {
		t.Fatalf("status = %q, want done", data["status"])
	}

	if data["result"] != "processed: third" {
		t.Fatalf("result = %q, want processed: third", data["result"])
	}
}
