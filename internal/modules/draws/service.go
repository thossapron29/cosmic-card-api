package draws

import (
	"context"
	"net/http"
)

type DrawRepository interface {
	FindDailyDrawByUserAndDate(ctx context.Context, userID, clientLocalDate string) (int64, error)
	RevealRandomCard(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error)
}

type Service struct {
	repo DrawRepository
}

func NewService(repo DrawRepository) *Service {
	return &Service{repo: repo}
}

var allowedDrawModes = map[string]struct{}{
	"daily":      {},
	"guidance":   {},
	"support":    {},
	"reflection": {},
}

func (s *Service) Reveal(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error) {
	if req.UserID == "" {
		return RevealDrawResponse{}, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "userId is required")
	}

	if req.DeckID == 0 {
		return RevealDrawResponse{}, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "deckId is required")
	}

	if req.Locale == "" {
		req.Locale = "en"
	}

	if req.DrawMode == "" {
		req.DrawMode = "daily"
	}

	if _, ok := allowedDrawModes[req.DrawMode]; !ok {
		return RevealDrawResponse{}, NewAppError(http.StatusBadRequest, "INVALID_DRAW_MODE", "drawMode must be one of: daily, guidance, support, reflection")
	}

	if req.ClientLocalDate == "" {
		req.ClientLocalDate = TodayString()
	}

	if req.DrawMode == "daily" {
		existingDrawID, err := s.repo.FindDailyDrawByUserAndDate(ctx, req.UserID, req.ClientLocalDate)
		if err != nil {
			return RevealDrawResponse{}, err
		}

		// Without entitlements in place yet, all users are treated as free users.
		if existingDrawID != 0 {
			return RevealDrawResponse{}, NewAppError(http.StatusConflict, "DAILY_DRAW_ALREADY_USED", "Daily card has already been used for this date.")
		}
	}

	return s.repo.RevealRandomCard(ctx, req)
}
