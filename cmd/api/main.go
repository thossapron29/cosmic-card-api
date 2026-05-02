package main

import (
	"log"

	"github.com/yourname/cosmic-card-api/internal/config"
	"github.com/yourname/cosmic-card-api/internal/router"
)

func main() {
	cfg := config.Load()

	r := router.New(cfg)

	addr := ":" + cfg.Port

	log.Println("Cosmic Card API running on", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
