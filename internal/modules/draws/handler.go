package draws

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Reveal(c *gin.Context) {
	var req RevealDrawRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeAppError(c, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "invalid request body"))
		return
	}

	res, err := h.service.Reveal(c.Request.Context(), req)
	if err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			writeAppError(c, appErr)
			return
		}

		if errors.Is(err, pgx.ErrNoRows) {
			writeAppError(c, NewAppError(http.StatusNotFound, "NO_CARD_AVAILABLE", "no available card found"))
			return
		}

		writeAppError(c, NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reveal card"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}

func writeAppError(c *gin.Context, err *AppError) {
	c.JSON(err.Status, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}
