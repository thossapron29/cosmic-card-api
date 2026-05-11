package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourname/cosmic-card-api/internal/config"
)

type APIHandlers struct {
	Decks DeckHandler
	Draws DrawHandler
}

type DeckHandler interface {
	GetDecks(c *gin.Context)
}

type DrawHandler interface {
	Reveal(c *gin.Context)
	GetHistory(c *gin.Context)
	GetTodayStatus(c *gin.Context)
}

func New(cfg config.Config, db *pgxpool.Pool, handlers APIHandlers) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(500, gin.H{
				"status":   "error",
				"database": "down",
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "up",
		})
	})

	r.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     cfg.AppName,
			"version": cfg.AppVersion,
			"env":     cfg.AppEnv,
		})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api/v1")

	if handlers.Decks != nil {
		api.GET("/decks", handlers.Decks.GetDecks)
	}

	if handlers.Draws != nil {
		api.GET("/draws", handlers.Draws.GetHistory)
		api.GET("/draws/today-status", handlers.Draws.GetTodayStatus)
		api.POST("/draws/reveal", handlers.Draws.Reveal)
	}

	return r
}
