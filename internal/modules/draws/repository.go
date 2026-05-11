package draws

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindDailyDrawByUserAndDate(ctx context.Context, userID, clientLocalDate string) (int64, error) {
	query := `
		SELECT id
		FROM user_draws
		WHERE user_id = $1
		  AND draw_mode = 'daily'
		  AND client_local_date = $2::date
		ORDER BY id DESC
		LIMIT 1
	`

	var drawID int64

	err := r.db.QueryRow(ctx, query, userID, clientLocalDate).Scan(&drawID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return drawID, nil
}

func (r *Repository) RevealRandomCard(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error) {
	query := `
		WITH selected_card AS (
			SELECT
				c.id,
				c.code,
				c.energy_type,
				COALESCE(c.illustration_key, '') AS illustration_key,
				ct.title,
				ct.short_message,
				COALESCE(ct.meaning, '') AS meaning,
				COALESCE(ct.reflection_prompt, '') AS reflection_prompt,
				COALESCE(ct.share_text, '') AS share_text,
				d.id AS deck_id,
				d.code AS deck_code,
				dt.name AS deck_name
			FROM cards c
			JOIN card_translations ct ON ct.card_id = c.id
			JOIN decks d ON d.id = c.deck_id
			JOIN deck_translations dt ON dt.deck_id = d.id
			WHERE c.deck_id = $1
			  AND c.is_active = true
			  AND ct.locale = $2
			  AND dt.locale = $2
			  AND (
				($3 = 'daily' AND c.allow_daily_draw = true)
				OR ($3 = 'guidance' AND c.allow_guidance_draw = true)
				OR ($3 = 'support' AND c.allow_support_draw = true)
				OR ($3 = 'reflection' AND c.allow_reflection_draw = true)
			  )
			ORDER BY random() * c.weight DESC
			LIMIT 1
		),
		inserted_draw AS (
			INSERT INTO user_draws (
				user_id,
				card_id,
				draw_mode,
				question_text,
				locale_at_time,
				deck_id,
				client_local_date
			)
			SELECT
				$4,
				id,
				$3,
				NULLIF($5, ''),
				$2,
				deck_id,
				NULLIF($6, '')::date
			FROM selected_card
			RETURNING id
		)
		SELECT
			inserted_draw.id AS draw_id,
			selected_card.id AS card_id,
			selected_card.code,
			selected_card.title,
			selected_card.short_message,
			selected_card.meaning,
			selected_card.reflection_prompt,
			selected_card.share_text,
			selected_card.illustration_key,
			selected_card.energy_type,
			selected_card.deck_id,
			selected_card.deck_code,
			selected_card.deck_name
		FROM inserted_draw
		CROSS JOIN selected_card
	`

	var res RevealDrawResponse

	err := r.db.QueryRow(
		ctx,
		query,
		req.DeckID,
		req.Locale,
		req.DrawMode,
		req.UserID,
		req.QuestionText,
		req.ClientLocalDate,
	).Scan(
		&res.DrawID,
		&res.Card.ID,
		&res.Card.Code,
		&res.Card.Title,
		&res.Card.ShortMessage,
		&res.Card.Meaning,
		&res.Card.ReflectionPrompt,
		&res.Card.ShareText,
		&res.Card.IllustrationKey,
		&res.Card.EnergyType,
		&res.Deck.ID,
		&res.Deck.Code,
		&res.Deck.Name,
	)

	return res, err
}

func TodayString() string {
	return time.Now().Format("2006-01-02")
}
