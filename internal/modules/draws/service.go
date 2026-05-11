package draws

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Reveal(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error) {
	if req.UserID == "" {
		return RevealDrawResponse{}, errors.New("userId is required")
	}

	if req.DeckID == 0 {
		return RevealDrawResponse{}, errors.New("deckId is required")
	}

	if req.Locale == "" {
		req.Locale = "en"
	}

	if req.DrawMode == "" {
		req.DrawMode = "daily"
	}

	if req.ClientLocalDate == "" {
		req.ClientLocalDate = TodayString()
	}

	return s.repo.RevealRandomCard(ctx, req)
}
