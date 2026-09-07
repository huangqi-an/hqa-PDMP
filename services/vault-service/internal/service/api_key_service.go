package service

import (
	"context"
	"errors"
	"time"

	"github.com/huangqi-an/hqa-PDMP/services/vault-service/internal/crypto"
	"github.com/huangqi-an/hqa-PDMP/services/vault-service/internal/model"
	"github.com/huangqi-an/hqa-PDMP/services/vault-service/internal/repository"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

type APIKeyService struct {
	repo      repository.Repository
	encryptor *crypto.Encryptor
}

func NewAPIKeyService(
	repo repository.Repository,
	encryptor *crypto.Encryptor,
) *APIKeyService {
	return &APIKeyService{
		repo:      repo,
		encryptor: encryptor,
	}
}

type CreateAPIKeyInput struct {
	Provider string
	Name     string
	Key      string
	Notes    *string
	Tags     []string
}

type UpdateAPIKeyInput struct {
	Provider *string
	Name     *string
	Key      *string
	Notes    *string
	Tags     *[]string
}

type ListFilter struct {
	Provider string
	Query    string
	Page     int
	PageSize int
}

type APIKeyDTO struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Name      string    `json:"name"`
	MaskedKey string    `json:"maskedKey"`
	KeyHint   string    `json:"keyHint"`
	Notes     *string   `json:"notes,omitempty"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListResult struct {
	Items []APIKeyDTO `json:"items"`
	Total int64       `json:"total"`
}

func keyHint(secret string) string {
	if len(secret) <= 4 {
		return secret
	}
	return secret[len(secret)-4:]
}

func toDTO(key model.APIKey) APIKeyDTO {
	tags := []string(key.Tags)
	if tags == nil {
		tags = []string{}
	}
	return APIKeyDTO{
		ID:        key.ID,
		Provider:  key.Provider,
		Name:      key.Name,
		MaskedKey: "****" + key.KeyHint,
		KeyHint:   key.KeyHint,
		Notes:     key.Notes,
		Tags:      tags,
		CreatedAt: key.CreatedAt,
		UpdatedAt: key.UpdatedAt,
	}
}

func (s *APIKeyService) Create(ctx context.Context, userID string, input CreateAPIKeyInput) (*APIKeyDTO, error) {
	encrypted, err := s.encryptor.Encrypt(input.Key)
	if err != nil {
		return nil, err
	}
	key := &model.APIKey{
		UserID:       userID,
		Provider:     input.Provider,
		Name:         input.Name,
		EncryptedKey: encrypted,
		KeyHint:      keyHint(input.Key),
		Notes:        input.Notes,
		Tags:         pq.StringArray(input.Tags),
	}
	if err := s.repo.Create(ctx, key); err != nil {
		return nil, err
	}
	dto := toDTO(*key)
	return &dto, nil
}

func (s *APIKeyService) Get(ctx context.Context, userID string, id string) (*APIKeyDTO, error) {
	key, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, ErrAPIKeyNotFound
	}
	dto := toDTO(*key)
	return &dto, nil
}

func (s *APIKeyService) List(ctx context.Context, userID string, filter ListFilter) (*ListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	limit := pageSize
	offset := (page - 1) * pageSize

	keys, total, err := s.repo.List(ctx, userID, filter.Provider, filter.Query, limit, offset)

	if err != nil {
		return nil, err
	}
	items := make([]APIKeyDTO, 0, len(keys))
	for _, key := range keys {
		items = append(items, toDTO(key))
	}
	return &ListResult{
		Items: items,
		Total: total,
	}, nil
}

func (s *APIKeyService) Update(ctx context.Context, userID string, id string, input UpdateAPIKeyInput) (*APIKeyDTO, error) {
	key, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, ErrAPIKeyNotFound
	}
	if input.Provider != nil {
		key.Provider = *input.Provider
	}
	if input.Name != nil {
		key.Name = *input.Name
	}
	if input.Notes != nil {
		key.Notes = input.Notes
	}
	if input.Tags != nil {
		key.Tags = pq.StringArray(*input.Tags)
	}
	if input.Key != nil {
		encrypted, err := s.encryptor.Encrypt(*input.Key)
		if err != nil {
			return nil, err
		}
		key.EncryptedKey = encrypted
		key.KeyHint = keyHint(*input.Key)
	}
	if err := s.repo.Update(ctx, key); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}
	key.UpdatedAt = time.Now()
	dto := toDTO(*key)
	return &dto, nil
}

func (s *APIKeyService) Delete(ctx context.Context, userId string, id string) error {
	err := s.repo.SoftDelete(ctx, userId, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAPIKeyNotFound
	}
	return err
}

func (s *APIKeyService) Reveal(ctx context.Context, userID string, id string) (string, error) {
	key, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "", ErrAPIKeyNotFound
	}
	return s.encryptor.Decrypt(key.EncryptedKey)
}
