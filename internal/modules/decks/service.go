package decks

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDecks(ctx context.Context, locale string) ([]Deck, error) {
	if locale == "" {
		locale = "en"
	}

	return s.repo.FindActiveDecks(ctx, locale)
}
