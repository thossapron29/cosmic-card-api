package decks

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDecks(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")

	decks, err := h.service.GetDecks(c.Request.Context(), locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get decks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": decks,
	})
}
