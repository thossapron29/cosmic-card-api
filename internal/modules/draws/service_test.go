package draws

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDrawRepository struct {
	response  RevealDrawResponse
	err       error
	gotReq    RevealDrawRequest
	callCount int
}

func (f *fakeDrawRepository) RevealRandomCard(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error) {
	f.gotReq = req
	f.callCount++

	if f.err != nil {
		return RevealDrawResponse{}, f.err
	}

	return f.response, nil
}

func TestServiceRevealRequiresUserID(t *testing.T) {
	repo := &fakeDrawRepository{}
	service := NewService(repo)

	_, err := service.Reveal(context.Background(), RevealDrawRequest{DeckID: 1})
	if err == nil || err.Error() != "userId is required" {
		t.Fatalf("expected userId validation error, got %v", err)
	}

	if repo.callCount != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.callCount)
	}
}

func TestServiceRevealRequiresDeckID(t *testing.T) {
	repo := &fakeDrawRepository{}
	service := NewService(repo)

	_, err := service.Reveal(context.Background(), RevealDrawRequest{UserID: "user_123"})
	if err == nil || err.Error() != "deckId is required" {
		t.Fatalf("expected deckId validation error, got %v", err)
	}

	if repo.callCount != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.callCount)
	}
}

func TestServiceRevealAppliesDefaultsBeforeCallingRepository(t *testing.T) {
	repo := &fakeDrawRepository{
		response: RevealDrawResponse{DrawID: 42},
	}
	service := NewService(repo)

	result, err := service.Reveal(context.Background(), RevealDrawRequest{
		UserID: "user_123",
		DeckID: 7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.DrawID != 42 {
		t.Fatalf("expected repository response, got %#v", result)
	}

	if repo.callCount != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.callCount)
	}

	if repo.gotReq.Locale != "en" {
		t.Fatalf("expected default locale en, got %q", repo.gotReq.Locale)
	}

	if repo.gotReq.DrawMode != "daily" {
		t.Fatalf("expected default draw mode daily, got %q", repo.gotReq.DrawMode)
	}

	expectedDate := time.Now().Format("2006-01-02")
	if repo.gotReq.ClientLocalDate != expectedDate {
		t.Fatalf("expected clientLocalDate %q, got %q", expectedDate, repo.gotReq.ClientLocalDate)
	}
}

func TestServiceRevealPreservesProvidedValues(t *testing.T) {
	repo := &fakeDrawRepository{}
	service := NewService(repo)

	req := RevealDrawRequest{
		UserID:          "user_123",
		DeckID:          7,
		DrawMode:        "support",
		Locale:          "th",
		QuestionText:    "How should I feel today?",
		ClientLocalDate: "2026-05-11",
	}

	_, err := service.Reveal(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.gotReq != req {
		t.Fatalf("expected request to be preserved, got %#v", repo.gotReq)
	}
}

func TestServiceRevealPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("repository failed")
	repo := &fakeDrawRepository{err: repoErr}
	service := NewService(repo)

	_, err := service.Reveal(context.Background(), RevealDrawRequest{
		UserID: "user_123",
		DeckID: 7,
		Locale: "en",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error %v, got %v", repoErr, err)
	}
}
