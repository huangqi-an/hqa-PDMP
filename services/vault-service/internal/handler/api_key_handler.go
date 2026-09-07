package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/huangqi-an/hqa-PDMP/services/vault-service/internal/service"
)

type APIKeyHandler struct {
	service *service.APIKeyService
}

func NewAPIKeyHandler(service *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

type CreateRequest struct {
	Provider string   `json:"provider" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Key      string   `json:"key" binding:"required"`
	Notes    *string  `json:"notes"`
	Tags     []string `json:"tags"`
}
type UpdateRequest struct {
	Provider *string   `json:"provider"`
	Name     *string   `json:"name"`
	Key      *string   `json:"key"`
	Notes    *string   `json:"notes"`
	Tags     *[]string `json:"tags"`
}

func success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func fail(c *gin.Context, status int, code int, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}

func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func handleServiceError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrAPIKeyNotFound) {
		fail(c, http.StatusNotFound, 1006, "密钥不存在")
		return
	}

	log.Printf("vault-service internal error: %v", err)
	fail(c, http.StatusInternalServerError, 9999, "服务器内部错误")
}

func (h *APIKeyHandler) List(c *gin.Context) {
	userID := c.GetString("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filter := service.ListFilter{
		Provider: c.Query("provider"),
		Query:    c.Query("q"),
		Page:     page,
		PageSize: pageSize,
	}
	result, err := h.service.List(c.Request.Context(), userID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	success(c, result)
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	userID := c.GetString("userID")
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 1002, "参数错误")
		return
	}
	key, err := h.service.Create(c.Request.Context(), userID, service.CreateAPIKeyInput{
		Provider: req.Provider,
		Name:     req.Name,
		Key:      req.Key,
		Notes:    req.Notes,
		Tags:     req.Tags,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	created(c, key)
}

func (h *APIKeyHandler) Get(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	key, err := h.service.Get(c.Request.Context(), userID, id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	success(c, key)
}

func (h *APIKeyHandler) Update(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 1002, "参数错误")
		return
	}
	key, err := h.service.Update(c.Request.Context(), userID, id, service.UpdateAPIKeyInput{
		Provider: req.Provider,
		Name:     req.Name,
		Key:      req.Key,
		Notes:    req.Notes,
		Tags:     req.Tags,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	success(c, key)
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), userID, id); err != nil {
		handleServiceError(c, err)
		return
	}
	success(c, nil)
}

func (h *APIKeyHandler) Reveal(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	plaintext, err := h.service.Reveal(c.Request.Context(), userID, id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	success(c, gin.H{
		"key": plaintext,
	})
}
