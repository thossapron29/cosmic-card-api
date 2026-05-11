package draws

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

type fakeDrawRepository struct {
	response           RevealDrawResponse
	err                error
	gotReq             RevealDrawRequest
	dailyDrawID        int64
	dailyErr           error
	callCount          int
	dailyCallCount     int
	gotUserID          string
	gotClientLocalDate string
}

func (f *fakeDrawRepository) FindDailyDrawByUserAndDate(ctx context.Context, userID, clientLocalDate string) (int64, error) {
	f.dailyCallCount++
	f.gotUserID = userID
	f.gotClientLocalDate = clientLocalDate

	if f.dailyErr != nil {
		return 0, f.dailyErr
	}

	return f.dailyDrawID, nil
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
	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Code != "INVALID_REQUEST" || appErr.Message != "userId is required" {
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
	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Code != "INVALID_REQUEST" || appErr.Message != "deckId is required" {
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

	if repo.dailyCallCount != 1 {
		t.Fatalf("expected daily lookup to be called once, got %d", repo.dailyCallCount)
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

func TestServiceRevealRejectsInvalidDrawMode(t *testing.T) {
	repo := &fakeDrawRepository{}
	service := NewService(repo)

	_, err := service.Reveal(context.Background(), RevealDrawRequest{
		UserID:   "user_123",
		DeckID:   7,
		DrawMode: "random",
	})

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %v", err)
	}

	if appErr.Status != http.StatusBadRequest || appErr.Code != "INVALID_DRAW_MODE" {
		t.Fatalf("expected invalid draw mode error, got %#v", appErr)
	}

	if repo.callCount != 0 {
		t.Fatalf("expected reveal not to be called, got %d calls", repo.callCount)
	}
}

func TestServiceRevealRejectsDuplicateDailyDraw(t *testing.T) {
	repo := &fakeDrawRepository{dailyDrawID: 99}
	service := NewService(repo)

	_, err := service.Reveal(context.Background(), RevealDrawRequest{
		UserID:          "user_123",
		DeckID:          7,
		DrawMode:        "daily",
		ClientLocalDate: "2026-05-11",
	})

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %v", err)
	}

	if appErr.Status != http.StatusConflict || appErr.Code != "DAILY_DRAW_ALREADY_USED" {
		t.Fatalf("expected daily conflict error, got %#v", appErr)
	}

	if repo.callCount != 0 {
		t.Fatalf("expected reveal not to be called, got %d calls", repo.callCount)
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

	if repo.dailyCallCount != 0 {
		t.Fatalf("expected no daily lookup for non-daily mode, got %d", repo.dailyCallCount)
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

func TestServiceRevealPropagatesDailyLookupError(t *testing.T) {
	repoErr := errors.New("daily lookup failed")
	repo := &fakeDrawRepository{dailyErr: repoErr}
	service := NewService(repo)

	_, err := service.Reveal(context.Background(), RevealDrawRequest{
		UserID:   "user_123",
		DeckID:   7,
		DrawMode: "daily",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error %v, got %v", repoErr, err)
	}
}
