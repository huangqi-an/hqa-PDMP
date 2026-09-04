package main

import (
	"log"

	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/config"
	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/database"
	"github.com/114514-art/hqa-PDMP/services/vault-service/internal/router"
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
	r := router.New(db)
	log.Printf("vault-service listening on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
