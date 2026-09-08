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

	streamLength := store.client.XLen(ctx, taskStreamKey).Val()
	if streamLength != 1 {
		t.Fatalf("stream length = %d, want 1", streamLength)
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

func TestReadGroupReturnsTaskAndAckRemovesPending(t *testing.T) {
	store, _ := newTestTaskStore(t)
	ctx := context.Background()

	if err := store.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup() error = %v", err)
	}

	task := Task{
		ID:      "task-2",
		Type:    "echo",
		Payload: "second",
		Status:  "queued",
	}

	if err := store.Enqueue(ctx, task); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	got, err := store.ReadGroup(ctx, "test-consumer", time.Second)
	if err != nil {
		t.Fatalf("ReadGroup() error = %v", err)
	}

	if got == nil {
		t.Fatal("ReadGroup() = nil, want task")
	}

	if got.Task.ID != "task-2" {
		t.Fatalf("got ID = %q, want task-2", got.Task.ID)
	}

	pending, err := store.client.XPending(ctx, taskStreamKey, taskGroupName).Result()
	if err != nil {
		t.Fatalf("XPending() error = %v", err)
	}

	if pending.Count != 1 {
		t.Fatalf("pending count = %d, want 1", pending.Count)
	}

	if err := store.Ack(ctx, got.StreamID); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}

	pending, err = store.client.XPending(ctx, taskStreamKey, taskGroupName).Result()
	if err != nil {
		t.Fatalf("XPending() after ack error = %v", err)
	}

	if pending.Count != 0 {
		t.Fatalf("pending count after ack = %d, want 0", pending.Count)
	}
}

func TestReadGroupReturnsNilWhenEmpty(t *testing.T) {
	store, _ := newTestTaskStore(t)

	if err := store.EnsureGroup(context.Background()); err != nil {
		t.Fatalf("EnsureGroup() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got, err := store.ReadGroup(ctx, "test-consumer", time.Second)
	if err != nil {
		t.Fatalf("ReadGroup() error = %v", err)
	}

	if got != nil {
		t.Fatalf("ReadGroup() = %#v, want nil", got)
	}
}
