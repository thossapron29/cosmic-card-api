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

func New(cfg config.Config, db *pgxpool.Pool) *gin.Engine {
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

	return r
}
