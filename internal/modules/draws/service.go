package draws

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type DrawRepository interface {
	FindDailyDrawByUserAndDate(ctx context.Context, userID, clientLocalDate string) (int64, error)
	CountDrawsByUserModeAndDate(ctx context.Context, userID, drawMode, clientLocalDate string) (int, error)
	FindDrawHistory(ctx context.Context, userID, locale string, limit int, cursor int64) ([]DrawHistoryItem, error)
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

var freeDrawLimitsByMode = map[string]int{
	"guidance":   2,
	"support":    1,
	"reflection": 1,
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

func (s *Service) GetHistory(ctx context.Context, userID, locale, cursor string, limit int) (DrawHistoryResponse, error) {
	if userID == "" {
		return DrawHistoryResponse{}, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "userId is required")
	}

	if locale == "" {
		locale = "en"
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	var cursorID int64
	if cursor != "" {
		parsed, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return DrawHistoryResponse{}, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "cursor must be a valid draw id")
		}
		cursorID = parsed
	}

	items, err := s.repo.FindDrawHistory(ctx, userID, locale, limit+1, cursorID)
	if err != nil {
		return DrawHistoryResponse{}, err
	}

	res := DrawHistoryResponse{
		Data:   items,
		Paging: DrawHistoryPaging{},
	}

	if len(items) > limit {
		lastVisible := items[limit-1]
		res.Data = items[:limit]
		res.Paging.NextCursor = fmt.Sprintf("%d", lastVisible.DrawID)
	}

	return res, nil
}

func (s *Service) GetTodayStatus(ctx context.Context, userID, clientLocalDate string) (TodayStatusResponse, error) {
	if userID == "" {
		return TodayStatusResponse{}, NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "userId is required")
	}

	if clientLocalDate == "" {
		clientLocalDate = TodayString()
	}

	dailyDrawID, err := s.repo.FindDailyDrawByUserAndDate(ctx, userID, clientLocalDate)
	if err != nil {
		return TodayStatusResponse{}, err
	}

	guidanceRemaining, err := s.remainingFreeDraws(ctx, userID, "guidance", clientLocalDate)
	if err != nil {
		return TodayStatusResponse{}, err
	}

	supportRemaining, err := s.remainingFreeDraws(ctx, userID, "support", clientLocalDate)
	if err != nil {
		return TodayStatusResponse{}, err
	}

	reflectionRemaining, err := s.remainingFreeDraws(ctx, userID, "reflection", clientLocalDate)
	if err != nil {
		return TodayStatusResponse{}, err
	}

	res := TodayStatusResponse{
		ClientLocalDate: clientLocalDate,
		Daily: TodayStatusDaily{
			Available: dailyDrawID == 0,
		},
		Guidance:   TodayStatusModeLimit{RemainingFreeDraws: guidanceRemaining},
		Support:    TodayStatusModeLimit{RemainingFreeDraws: supportRemaining},
		Reflection: TodayStatusModeLimit{RemainingFreeDraws: reflectionRemaining},
	}

	if dailyDrawID != 0 {
		res.Daily.DrawID = dailyDrawID
	}

	return res, nil
}

func (s *Service) remainingFreeDraws(ctx context.Context, userID, drawMode, clientLocalDate string) (int, error) {
	limit := freeDrawLimitsByMode[drawMode]

	used, err := s.repo.CountDrawsByUserModeAndDate(ctx, userID, drawMode, clientLocalDate)
	if err != nil {
		return 0, err
	}

	remaining := limit - used
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}
