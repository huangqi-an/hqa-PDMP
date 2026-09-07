package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/huangqi-an/hqa-PDMP/services/worker-service/internal/config"
	"github.com/huangqi-an/hqa-PDMP/services/worker-service/internal/handler"
	"github.com/huangqi-an/hqa-PDMP/services/worker-service/internal/store"
	"github.com/huangqi-an/hqa-PDMP/services/worker-service/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := store.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatal(err)
	}

	taskStore := store.NewTaskStore(redisClient)
	taskHandler := handler.NewTaskHandler(taskStore)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/tasks")
	api.POST("", taskHandler.Create)
	api.GET("/:id", taskHandler.Get)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go worker.Run(ctx, taskStore, cfg.WorkerCount)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("worker-service listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down worker-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}
}
