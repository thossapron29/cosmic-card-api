package draws

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerRevealReturnsStructuredErrorForInvalidDrawMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := NewService(&fakeDrawRepository{})
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/draws/reveal", handler.Reveal)

	body := bytes.NewBufferString(`{"userId":"user_123","deckId":1,"drawMode":"random"}`)
	req := httptest.NewRequest(http.MethodPost, "/draws/reveal", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var response map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got %v", err)
	}

	if response["error"]["code"] != "INVALID_DRAW_MODE" {
		t.Fatalf("expected INVALID_DRAW_MODE, got %#v", response)
	}
}

func TestHandlerRevealReturnsStructuredErrorForInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := NewService(&fakeDrawRepository{})
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/draws/reveal", handler.Reveal)

	req := httptest.NewRequest(http.MethodPost, "/draws/reveal", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var response map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got %v", err)
	}

	if response["error"]["code"] != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %#v", response)
	}
}

func TestHandlerRevealReturnsStructuredNotFoundWhenNoCardIsAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := NewService(&fakeNoCardRepository{})
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/draws/reveal", handler.Reveal)

	body := bytes.NewBufferString(`{"userId":"user_123","deckId":1,"drawMode":"guidance"}`)
	req := httptest.NewRequest(http.MethodPost, "/draws/reveal", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var response map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got %v", err)
	}

	if response["error"]["code"] != "NO_CARD_AVAILABLE" {
		t.Fatalf("expected NO_CARD_AVAILABLE, got %#v", response)
	}
}

type fakeNoCardRepository struct{}

func (f *fakeNoCardRepository) FindDailyDrawByUserAndDate(ctx context.Context, userID, clientLocalDate string) (int64, error) {
	return 0, nil
}

func (f *fakeNoCardRepository) RevealRandomCard(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error) {
	return RevealDrawResponse{}, NewAppError(http.StatusNotFound, "NO_CARD_AVAILABLE", "no available card found")
}
