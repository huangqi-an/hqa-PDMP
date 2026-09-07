package main

import (
	"log"

	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/config"
	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/crypto"
	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/database"
	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/handler"
	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/repository"
	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/router"
	"github.com/haungqi-an/hqa-PDMP/services/vault-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	encryptor, err := crypto.New(cfg.EncryptionKey)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewAPIKeyRepository(db)
	svc := service.NewAPIKeyService(repo, encryptor)
	keyHandler := handler.NewAPIKeyHandler(svc)

	r := router.New(db, cfg.JWTSecret, keyHandler)
	log.Printf("vault-service listening on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
