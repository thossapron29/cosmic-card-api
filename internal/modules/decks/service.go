package decks

import "context"

type DeckRepository interface {
	FindActiveDecks(ctx context.Context, locale string) ([]Deck, error)
}

type Service struct {
	repo DeckRepository
}

func NewService(repo DeckRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDecks(ctx context.Context, locale string) ([]Deck, error) {
	if locale == "" {
		locale = "en"
	}

	return s.repo.FindActiveDecks(ctx, locale)
}
