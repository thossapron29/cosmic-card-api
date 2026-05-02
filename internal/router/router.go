package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourname/cosmic-card-api/internal/config"
)

func New(cfg config.Config) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
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
