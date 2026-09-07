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
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			consume(ctx, workerID, taskStore)
		}(i)
	}

	wg.Wait()
}

func consume(ctx context.Context, workerID int, taskStore *store.TaskStore) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %d stopped", workerID)
			return
		default:
		}

		task, err := taskStore.BlockingPop(ctx, 5*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Printf("worker %d pop task failed: %v", workerID, err)
			continue
		}

		if task == nil {
			continue
		}

		log.Printf("worker %d processing task %s", workerID, task.ID)

		if err := taskStore.SetStatus(ctx, task.ID, "running", ""); err != nil {
			log.Printf("worker %d set running failed: %v", workerID, err)
			continue
		}

		result, err := processTask(*task)
		if err != nil {
			_ = taskStore.SetStatus(ctx, task.ID, "failed", err.Error())
			continue
		}

		if err := taskStore.SetStatus(ctx, task.ID, "done", result); err != nil {
			log.Printf("worker %d set done failed: %v", workerID, err)
		}
	}
}

func processTask(task store.Task) (string, error) {
	// 先用 sleep 模拟耗时任务
	time.Sleep(2 * time.Second)

	return fmt.Sprintf("processed: %s", task.Payload), nil
}
