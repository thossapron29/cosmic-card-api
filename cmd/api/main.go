package main

import (
	"log"

	"github.com/yourname/cosmic-card-api/internal/config"
	"github.com/yourname/cosmic-card-api/internal/database"
	"github.com/yourname/cosmic-card-api/internal/modules/decks"
	"github.com/yourname/cosmic-card-api/internal/modules/draws"
	"github.com/yourname/cosmic-card-api/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	deckRepo := decks.NewRepository(db)
	deckService := decks.NewService(deckRepo)
	deckHandler := decks.NewHandler(deckService)

	drawRepo := draws.NewRepository(db)
	drawService := draws.NewService(drawRepo)
	drawHandler := draws.NewHandler(drawService)

	r := router.New(cfg, db, router.APIHandlers{
		Decks: deckHandler,
		Draws: drawHandler,
	})

	addr := ":" + cfg.Port

	log.Println("Cosmic Card API running on", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
