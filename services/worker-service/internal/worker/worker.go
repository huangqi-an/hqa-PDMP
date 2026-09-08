package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/huangqi-an/hqa-PDMP/services/worker-service/internal/store"
)

func Run(ctx context.Context, taskStore *store.TaskStore, workerCount int) {
	if err := taskStore.EnsureGroup(ctx); err != nil {
		log.Printf("worker ensure group failed: %v", err)
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			consume(ctx, fmt.Sprintf("worker-%d", workerID), taskStore)
		}(i)
	}

	wg.Wait()
}

func consume(ctx context.Context, consumer string, taskStore *store.TaskStore) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %s stopped", consumer)
			return
		default:
		}

		streamTask, err := taskStore.ReadGroup(ctx, consumer, 5*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Printf("worker %s read task failed: %v", consumer, err)
			continue
		}

		if streamTask == nil {
			continue
		}

		task := streamTask.Task

		log.Printf("worker %s processing task %s", consumer, task.ID)

		if err := taskStore.SetStatus(ctx, task.ID, "running", ""); err != nil {
			log.Printf("worker %s set running failed: %v", consumer, err)
			continue
		}

		result, err := processTask(task)
		if err != nil {
			_ = taskStore.SetStatus(ctx, task.ID, "failed", err.Error())
		} else {
			if err := taskStore.SetStatus(ctx, task.ID, "done", result); err != nil {
				log.Printf("worker %s set done failed: %v", consumer, err)
			}
		}

		if err := taskStore.Ack(ctx, streamTask.StreamID); err != nil {
			log.Printf("worker %s ack failed: %v", consumer, err)
		}
	}
}

func processTask(task store.Task) (string, error) {
	// 先用 sleep 模拟耗时任务
	time.Sleep(2 * time.Second)

	return fmt.Sprintf("processed: %s", task.Payload), nil
}
