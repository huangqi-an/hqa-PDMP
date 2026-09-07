package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/crypto"
	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/model"
)

type fakeRepo struct {
	createCalls int
	createdKey  *model.APIKey
	createErr   error

	findByIDResult *model.APIKey
	findByIDErr    error

	listResult   []model.APIKey
	listTotal    int64
	listErr      error
	listProvider string
	listQuery    string
	listLimit    int
	listOffset   int

	updateErr     error
	softDeleteErr error
}

func (f *fakeRepo) Create(ctx context.Context, key *model.APIKey) error {
	f.createCalls++
	f.createdKey = key
	return f.createErr
}

func (f *fakeRepo) FindByID(ctx context.Context, userID string, id string) (*model.APIKey, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeRepo) List(
	ctx context.Context,
	userID string,
	provider string,
	query string,
	limit int,
	offset int,
) ([]model.APIKey, int64, error) {
	f.listProvider = provider
	f.listQuery = query
	f.listLimit = limit
	f.listOffset = offset

	return f.listResult, f.listTotal, f.listErr
}

func (f *fakeRepo) Update(ctx context.Context, key *model.APIKey) error {
	return f.updateErr
}

func (f *fakeRepo) SoftDelete(ctx context.Context, userID string, id string) error {
	return f.softDeleteErr
}

func newTestService() (*APIKeyService, *fakeRepo, *crypto.Encryptor) {
	encryptor, _ := crypto.New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	repo := &fakeRepo{}

	return NewAPIKeyService(repo, encryptor), repo, encryptor
}

func TestCreateEncryptsAPIKey(t *testing.T) {
	svc, repo, _ := newTestService()

	input := CreateAPIKeyInput{
		Provider: "openai",
		Name:     "default",
		Key:      "sk-test-1234abcd",
		Tags:     []string{"test"},
	}

	got, err := svc.Create(context.Background(), "user-123", input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.createCalls != 1 {
		t.Fatalf("createCalls = %d, want 1", repo.createCalls)
	}

	if repo.createdKey == nil {
		t.Fatal("createdKey should not be nil")
	}

	if repo.createdKey.EncryptedKey == input.Key {
		t.Fatal("EncryptedKey should not equal plaintext")
	}

	if repo.createdKey.KeyHint != "abcd" {
		t.Fatalf("KeyHint = %q, want abcd", repo.createdKey.KeyHint)
	}

	if got.MaskedKey != "****abcd" {
		t.Fatalf("MaskedKey = %q, want ****abcd", got.MaskedKey)
	}
}

func TestListUsesDefaultPagination(t *testing.T) {
	svc, repo, _ := newTestService()

	_, err := svc.List(context.Background(), "user-123", ListFilter{
		Page:     0,
		PageSize: 0,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if repo.listLimit != 20 {
		t.Fatalf("limit = %d, want 20", repo.listLimit)
	}

	if repo.listOffset != 0 {
		t.Fatalf("offset = %d, want 0", repo.listOffset)
	}
}

func TestGetReturnsNotFoundWhenRepoReturnsNil(t *testing.T) {
	svc, repo, _ := newTestService()
	repo.findByIDResult = nil
	repo.findByIDErr = nil

	_, err := svc.Get(context.Background(), "user-123", "key-123")

	if !errors.Is(err, ErrAPIKeyNotFound) {
		t.Fatalf("Get() error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestRevealDecryptsStoredKey(t *testing.T) {
	svc, repo, encryptor := newTestService()

	plaintext := "sk-secret-123456"
	encrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	repo.findByIDResult = &model.APIKey{
		ID:           "key-123",
		UserID:       "user-123",
		EncryptedKey: encrypted,
	}

	got, err := svc.Reveal(context.Background(), "user-123", "key-123")
	if err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}

	if got != plaintext {
		t.Fatalf("Reveal() = %q, want %q", got, plaintext)
	}
}

func TestDeleteMapsRecordNotFound(t *testing.T) {
	svc, repo, _ := newTestService()
	repo.softDeleteErr = gorm.ErrRecordNotFound

	err := svc.Delete(context.Background(), "user-123", "key-123")

	if !errors.Is(err, ErrAPIKeyNotFound) {
		t.Fatalf("Delete() error = %v, want ErrAPIKeyNotFound", err)
	}
}
