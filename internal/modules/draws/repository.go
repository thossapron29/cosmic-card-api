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

func (r *Repository) CountDrawsByUserModeAndDate(ctx context.Context, userID, drawMode, clientLocalDate string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_draws
		WHERE user_id = $1
		  AND draw_mode = $2
		  AND client_local_date = $3::date
	`

	var count int

	err := r.db.QueryRow(ctx, query, userID, drawMode, clientLocalDate).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) FindDrawHistory(ctx context.Context, userID, locale string, limit int, cursor int64) ([]DrawHistoryItem, error) {
	query := `
		SELECT
			ud.id AS draw_id,
			ud.draw_mode,
			COALESCE(ud.question_text, '') AS question_text,
			COALESCE(ud.client_local_date::text, '') AS client_local_date,
			ud.drawn_at,
			d.id AS deck_id,
			d.code AS deck_code,
			COALESCE(dt.name, dt_en.name) AS deck_name,
			c.id AS card_id,
			c.code AS card_code,
			COALESCE(ct.title, ct_en.title, tt.name, tt_en.name, c.code) AS card_title,
			COALESCE(ct.short_message, ct_en.short_message, tt.description, tt_en.description, '') AS short_message
		FROM user_draws ud
		JOIN decks d ON d.id = ud.deck_id
		JOIN cards c ON c.id = ud.card_id
		LEFT JOIN deck_translations dt
			ON dt.deck_id = d.id
		   AND dt.locale = $2
		LEFT JOIN deck_translations dt_en
			ON dt_en.deck_id = d.id
		   AND dt_en.locale = 'en'
		LEFT JOIN card_translations ct
			ON ct.card_id = c.id
		   AND ct.locale = $2
		LEFT JOIN card_translations ct_en
			ON ct_en.card_id = c.id
		   AND ct_en.locale = 'en'
		LEFT JOIN theme_translations tt
			ON tt.theme_id = c.theme_id
		   AND tt.locale = $2
		LEFT JOIN theme_translations tt_en
			ON tt_en.theme_id = c.theme_id
		   AND tt_en.locale = 'en'
		WHERE ud.user_id = $1
		  AND ($4 = 0 OR ud.id < $4)
		ORDER BY ud.id DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, userID, locale, limit, cursor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DrawHistoryItem, 0)

	for rows.Next() {
		var item DrawHistoryItem
		var revealedAt time.Time

		if err := rows.Scan(
			&item.DrawID,
			&item.DrawMode,
			&item.QuestionText,
			&item.ClientLocalDate,
			&revealedAt,
			&item.Deck.ID,
			&item.Deck.Code,
			&item.Deck.Name,
			&item.Card.ID,
			&item.Card.Code,
			&item.Card.Title,
			&item.Card.ShortMessage,
		); err != nil {
			return nil, err
		}

		item.RevealedAt = revealedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) RevealRandomCard(ctx context.Context, req RevealDrawRequest) (RevealDrawResponse, error) {
	query := `
		WITH selected_card AS (
			SELECT
				c.id,
				c.code,
				c.energy_type,
				COALESCE(c.illustration_key, '') AS illustration_key,
				COALESCE(ct.title, ct_en.title, tt.name, tt_en.name, c.code) AS title,
				COALESCE(ct.short_message, ct_en.short_message, tt.description, tt_en.description, '') AS short_message,
				COALESCE(ct.meaning, ct_en.meaning, tt.description, tt_en.description, '') AS meaning,
				COALESCE(ct.reflection_prompt, ct_en.reflection_prompt, 'What part of this message wants your attention today?') AS reflection_prompt,
				COALESCE(ct.share_text, ct_en.share_text, 'Today I drew ' || COALESCE(ct.title, ct_en.title, tt.name, tt_en.name, c.code) || '.') AS share_text,
				d.id AS deck_id,
				d.code AS deck_code,
				COALESCE(dt.name, dt_en.name) AS deck_name
			FROM cards c
			LEFT JOIN card_translations ct ON ct.card_id = c.id AND ct.locale = $2
			LEFT JOIN card_translations ct_en ON ct_en.card_id = c.id AND ct_en.locale = 'en'
			LEFT JOIN theme_translations tt ON tt.theme_id = c.theme_id AND tt.locale = $2
			LEFT JOIN theme_translations tt_en ON tt_en.theme_id = c.theme_id AND tt_en.locale = 'en'
			JOIN decks d ON d.id = c.deck_id
			LEFT JOIN deck_translations dt ON dt.deck_id = d.id AND dt.locale = $2
			LEFT JOIN deck_translations dt_en ON dt_en.deck_id = d.id AND dt_en.locale = 'en'
			WHERE c.deck_id = $1
			  AND c.is_active = true
			  AND (ct.id IS NOT NULL OR ct_en.id IS NOT NULL OR tt.id IS NOT NULL OR tt_en.id IS NOT NULL)
			  AND (dt.id IS NOT NULL OR dt_en.id IS NOT NULL)
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
