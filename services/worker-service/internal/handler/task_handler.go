package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/huangqi-an/hqa-PDMP/services/worker-service/internal/store"
)

type TaskHandler struct {
	store *store.TaskStore
}

func NewTaskHandler(store *store.TaskStore) *TaskHandler {
	return &TaskHandler{store: store}
}

type CreateTaskRequest struct {
	Type    string `json:"type" binding:"required"`
	Payload string `json:"payload" binding:"required"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1002,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	task := store.Task{
		ID:      uuid.NewString(),
		Type:    req.Type,
		Payload: req.Payload,
		Status:  "queued",
	}

	if err := h.store.Enqueue(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    9999,
			"message": "任务入队失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"id":     task.ID,
			"status": task.Status,
		},
	})
}

func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")

	data, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    9999,
			"message": "查询任务失败",
			"data":    nil,
		})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    1006,
			"message": "任务不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}
