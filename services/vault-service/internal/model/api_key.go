package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type APIKey struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string         `gorm:"type:text;not null;index:idx_api_keys_user_deleted" json:"userId"`
	Provider     string         `gorm:"type:text;not null" json:"provider"`
	Name         string         `gorm:"type:text;not null" json:"name"`
	EncryptedKey string         `gorm:"type:text;not null" json:"-"`
	KeyHint      string         `gorm:"type:text;not null" json:"keyHint"`
	Notes        *string        `gorm:"type:text" json:"notes,omitempty"`
	Tags         pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"tags"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index:idx_api_keys_user_deleted" json:"-"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
