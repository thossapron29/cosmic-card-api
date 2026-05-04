package main

import (
	"log"

	"github.com/yourname/cosmic-card-api/internal/config"
	"github.com/yourname/cosmic-card-api/internal/database"
	"github.com/yourname/cosmic-card-api/internal/router"
)

func main() {
	cfg := config.Load()

	db := database.NewPostgresPool(cfg.DatabaseURL)
	defer db.Close()

	r := router.New(cfg, db)

	addr := ":" + cfg.Port

	log.Println("Cosmic Card API running on", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
