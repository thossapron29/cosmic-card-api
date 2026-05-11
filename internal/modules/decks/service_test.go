package decks

import (
	"context"
	"errors"
	"testing"
)

type fakeDeckRepository struct {
	decks     []Deck
	err       error
	gotLocale string
	callCount int
}

func (f *fakeDeckRepository) FindActiveDecks(ctx context.Context, locale string) ([]Deck, error) {
	f.gotLocale = locale
	f.callCount++

	if f.err != nil {
		return nil, f.err
	}

	return f.decks, nil
}

func TestServiceGetDecksDefaultsLocaleToEnglish(t *testing.T) {
	repo := &fakeDeckRepository{
		decks: []Deck{{ID: 1, Code: "cosmic-guidance"}},
	}
	service := NewService(repo)

	result, err := service.GetDecks(context.Background(), "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.gotLocale != "en" {
		t.Fatalf("expected locale to default to en, got %q", repo.gotLocale)
	}

	if repo.callCount != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.callCount)
	}

	if len(result) != 1 || result[0].Code != "cosmic-guidance" {
		t.Fatalf("expected returned decks from repository, got %#v", result)
	}
}

func TestServiceGetDecksUsesProvidedLocale(t *testing.T) {
	repo := &fakeDeckRepository{}
	service := NewService(repo)

	_, err := service.GetDecks(context.Background(), "th")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.gotLocale != "th" {
		t.Fatalf("expected locale th, got %q", repo.gotLocale)
	}
}

func TestServiceGetDecksPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("repository failed")
	repo := &fakeDeckRepository{err: repoErr}
	service := NewService(repo)

	_, err := service.GetDecks(context.Background(), "en")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error %v, got %v", repoErr, err)
	}
}
