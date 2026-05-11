package draws

import (
	"errors"
	"fmt"
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

func (h *Handler) GetHistory(c *gin.Context) {
	limit := 20
	if rawLimit := c.Query("limit"); rawLimit != "" {
		var parsed int
		if _, err := fmt.Sscanf(rawLimit, "%d", &parsed); err != nil {
			writeAppError(c, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "limit must be a valid integer"))
			return
		}
		limit = parsed
	}

	res, err := h.service.GetHistory(
		c.Request.Context(),
		c.Query("userId"),
		c.DefaultQuery("locale", "en"),
		c.Query("cursor"),
		limit,
	)
	if err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			writeAppError(c, appErr)
			return
		}

		writeAppError(c, NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get draw history"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   res.Data,
		"paging": res.Paging,
	})
}

func (h *Handler) GetTodayStatus(c *gin.Context) {
	res, err := h.service.GetTodayStatus(
		c.Request.Context(),
		c.Query("userId"),
		c.Query("clientLocalDate"),
	)
	if err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			writeAppError(c, appErr)
			return
		}

		writeAppError(c, NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get today status"))
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
