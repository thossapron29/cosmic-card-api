package decks

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindActiveDecks(ctx context.Context, locale string) ([]Deck, error) {
	query := `
		SELECT
			d.id,
			d.code,
			dt.name,
			COALESCE(dt.short_description, '') AS short_description,
			COALESCE(d.cover_image, '') AS cover_image,
			COALESCE(d.icon_name, '') AS icon_name,
			d.is_premium
		FROM decks d
		JOIN deck_translations dt ON dt.deck_id = d.id
		WHERE d.is_active = true
		  AND dt.locale = $1
		ORDER BY d.sort_order ASC, d.id ASC
	`

	rows, err := r.db.Query(ctx, query, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decks := make([]Deck, 0)

	for rows.Next() {
		var deck Deck

		if err := rows.Scan(
			&deck.ID,
			&deck.Code,
			&deck.Name,
			&deck.ShortDescription,
			&deck.CoverImage,
			&deck.IconName,
			&deck.IsPremium,
		); err != nil {
			return nil, err
		}

		decks = append(decks, deck)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return decks, nil
}
