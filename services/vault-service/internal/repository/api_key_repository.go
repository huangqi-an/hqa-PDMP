package repository

import (
	"context"
	"errors"
	"time"

	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/model"
	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(
	ctx context.Context,
	key *model.APIKey,
) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *APIKeyRepository) FindByID(ctx context.Context, userID string, id string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&key).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (r *APIKeyRepository) List(ctx context.Context, userID string, provider string, query string, limit int, offset int) ([]model.APIKey, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.APIKey{}).Where("user_id = ?", userID)
	if provider != "" {
		q = q.Where("provider = ?", provider)
	}
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("name ILIKE ? OR provider ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var keys []model.APIKey
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&keys).Error

	return keys, total, err
}

func (r *APIKeyRepository) Update(ctx context.Context, key *model.APIKey) error {
	updates := map[string]any{
		"provider":      key.Provider,
		"name":          key.Name,
		"encrypted_key": key.EncryptedKey,
		"key_hint":      key.KeyHint,
		"notes":         key.Notes,
		"tags":          key.Tags,
		"updated_at":    time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&model.APIKey{}).Where("id = ? AND user_id = ?", key.ID, key.UserID).Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *APIKeyRepository) SoftDelete(ctx context.Context, userID string, id string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.APIKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
