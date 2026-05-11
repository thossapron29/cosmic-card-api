# Cosmic Card API Spec

## Purpose

This document translates the `01_Product Concept` into an API-oriented MVP scope for the current backend.

Product goal:

- Help users pause, draw a card, receive a short message, and reflect.
- Prioritize clarity, calm, and emotional resonance over prediction.
- Keep the experience lightweight, intentional, and repeatable.

## Product-to-API Summary

From the product concept, the backend needs to support four core user moments:

- Daily ritual: one intentional daily card.
- Ask the Universe: on-demand guidance draw tied to a question.
- Emotional support: quick comfort-focused draw.
- Reflection: end-of-day reflective draw.

That means the API should provide:

- Deck discovery.
- Card reveal by mode.
- Draw persistence for history and habit loops.
- Rules for free vs premium access.
- Enough metadata for the client to render a meaningful reveal experience.

## Current Backend Status

Already present in codebase:

- `GET /health`
- `GET /info`
- `GET /metrics`
- `GET /api/v1/decks`
- `POST /api/v1/draws/reveal`

Current implemented behavior:

- Deck list is locale-aware.
- Draw reveal picks one random active card from a deck.
- Draw modes currently supported in query logic:
  - `daily`
  - `guidance`
  - `support`
  - `reflection`
- Each reveal stores a record in `user_draws`.

Current gap:

- The API can reveal a card, but it does not yet enforce most product rules around ritual, entitlements, limits, or history.

## MVP API Scope

### 1. Deck Catalog

#### `GET /api/v1/decks?locale=en`

Purpose:

- Show available decks for the app home and deck picker.

Response shape:

```json
{
  "data": [
    {
      "id": 1,
      "code": "cosmic-guidance",
      "name": "Cosmic Guidance",
      "shortDescription": "Daily clarity and gentle perspective",
      "coverImage": "deck/cosmic-guidance.png",
      "iconName": "sparkles",
      "isPremium": false
    }
  ]
}
```

MVP rules:

- Return active decks only.
- Support localized deck labels.
- Include `isPremium` so the client can gate access.

Recommended additions:

- `cardCount`
- `modes`
- `sortOrder`
- `isActive`

### 2. Reveal a Card

#### `POST /api/v1/draws/reveal`

Purpose:

- Execute the core ritual loop: Ask -> Draw -> Reveal.

Request:

```json
{
  "userId": "user_123",
  "deckId": 1,
  "drawMode": "daily",
  "locale": "en",
  "questionText": "What should I focus on today?",
  "clientLocalDate": "2026-05-11"
}
```

Response:

```json
{
  "data": {
    "drawId": 42,
    "card": {
      "id": 101,
      "code": "inner-light",
      "title": "Inner Light",
      "shortMessage": "The answer becomes clearer when you stop forcing it.",
      "meaning": "Step back and give your thoughts room to settle. Clarity will come from softness, not pressure.",
      "reflectionPrompt": "Where in your life would gentleness help more than control?",
      "shareText": "Today I drew Inner Light.",
      "illustrationKey": "inner-light-v1",
      "energyType": "calm"
    },
    "deck": {
      "id": 1,
      "code": "cosmic-guidance",
      "name": "Cosmic Guidance"
    }
  }
}
```

MVP rules:

- `userId` is required.
- `deckId` is required.
- `locale` defaults to `en`.
- `drawMode` defaults to `daily`.
- `clientLocalDate` defaults to server date when omitted.
- Draw only from active cards in the selected deck and locale.
- Persist every successful reveal to `user_draws`.

Business rules still needed:

- Enforce one `daily` draw per user per local date for free users.
- Enforce per-mode limits for free users.
- Block premium-only decks for non-premium users.
- Validate allowed `drawMode` values explicitly.
- Return a product-friendly error when no card is available for the requested mode.

Recommended error contract:

```json
{
  "error": {
    "code": "DAILY_DRAW_ALREADY_USED",
    "message": "Daily card has already been used for this date."
  }
}
```

Suggested error codes:

- `INVALID_REQUEST`
- `INVALID_DRAW_MODE`
- `DECK_NOT_FOUND`
- `DECK_PREMIUM_REQUIRED`
- `DRAW_LIMIT_REACHED`
- `DAILY_DRAW_ALREADY_USED`
- `NO_CARD_AVAILABLE`

### 3. Draw History

#### `GET /api/v1/draws?userId=user_123&limit=20&cursor=...`

Purpose:

- Support reflection and revisit past messages.

Response:

```json
{
  "data": [
    {
      "drawId": 42,
      "drawMode": "daily",
      "questionText": "What should I focus on today?",
      "clientLocalDate": "2026-05-11",
      "revealedAt": "2026-05-11T08:00:00Z",
      "deck": {
        "id": 1,
        "code": "cosmic-guidance",
        "name": "Cosmic Guidance"
      },
      "card": {
        "id": 101,
        "code": "inner-light",
        "title": "Inner Light",
        "shortMessage": "The answer becomes clearer when you stop forcing it."
      }
    }
  ],
  "paging": {
    "nextCursor": "..."
  }
}
```

Why it matters:

- Reflection mode and retention both depend on users being able to return to prior draws.

### 4. Today Status

#### `GET /api/v1/draws/today-status?userId=user_123&clientLocalDate=2026-05-11`

Purpose:

- Let the client know whether today’s ritual is available and what other limits apply.

Response:

```json
{
  "data": {
    "clientLocalDate": "2026-05-11",
    "daily": {
      "available": false,
      "drawId": 42
    },
    "guidance": {
      "remainingFreeDraws": 2
    },
    "support": {
      "remainingFreeDraws": 1
    },
    "reflection": {
      "remainingFreeDraws": 1
    }
  }
}
```

Why it matters:

- The product is ritual-based, so the app should know whether the daily card is available before the user taps into the flow.

### 5. User Entitlements

#### `GET /api/v1/users/{userId}/entitlements`

Purpose:

- Tell the client whether the user can access premium decks and unlimited draws.

Response:

```json
{
  "data": {
    "plan": "free",
    "isPremium": false,
    "features": {
      "unlimitedDraws": false,
      "premiumDecks": false
    }
  }
}
```

Why it matters:

- The product concept explicitly uses a freemium model.

## Required Domain Rules

### Draw modes

The product concept implies these canonical modes:

- `daily`
- `guidance`
- `support`
- `reflection`

Recommendation:

- Treat mode values as an enum in service validation.

### Deck access

Launch decks from concept:

- `cosmic-guidance`
- `self-love-healing`
- `decision-clarity`
- `growth-opportunity`
- `universe-signs`

Rules:

- Some decks can be free.
- Some decks can be premium-only.
- Client should not decide access on its own; API must enforce it.

### Card structure

Each card should expose:

- `title`
- `shortMessage`
- `meaning`
- `reflectionPrompt` optional
- `shareText` optional
- `illustrationKey`
- `energyType`

### Localization

Decks and cards are already locale-based in current queries.

MVP rule:

- If the requested locale does not exist, either:
  - fall back to `en`, or
  - return a clear validation error.

Recommendation:

- Prefer fallback to `en` for product smoothness.

## Suggested Database / Model Needs

Based on current code and product concept, backend likely needs these entities:

- `decks`
- `deck_translations`
- `cards`
- `card_translations`
- `user_draws`
- `users`
- `subscriptions` or `user_entitlements`

Useful `user_draws` fields beyond current usage:

- `id`
- `user_id`
- `deck_id`
- `card_id`
- `draw_mode`
- `question_text`
- `locale_at_time`
- `client_local_date`
- `created_at`
- `is_premium_draw`

Potential future entities:

- `journal_entries`
- `favorites`
- `user_preferences`

## Priority Build Plan

### Phase 1: Complete the launch ritual

- Keep `GET /api/v1/decks`
- Harden `POST /api/v1/draws/reveal`
- Add explicit draw mode validation
- Add premium deck access validation
- Add daily draw uniqueness rule
- Add structured error codes

### Phase 2: Support retention

- Add `GET /api/v1/draws`
- Add `GET /api/v1/draws/today-status`

### Phase 3: Support monetization

- Add `GET /api/v1/users/{userId}/entitlements`
- Add premium limits and premium deck unlocking

## Implementation Notes For This Repo

Recommended near-term changes in current codebase:

- `internal/modules/draws/service.go`
  - validate `drawMode` against allowed values
  - check daily-limit and free-limit rules before reveal

- `internal/modules/draws/repository.go`
  - add query for existing daily draw by `user_id + draw_mode + client_local_date`
  - add history listing query
  - add today-status aggregate query

- `internal/modules/decks/repository.go`
  - optionally expose more deck metadata for client rendering and gating

- `internal/router/router.go`
  - add history, today-status, and entitlements routes

## Non-Goals For MVP

Do not build yet unless product scope changes:

- Fortune-telling logic
- Complex spreads or multi-card tarot systems
- Heavy personalization models
- Social feed features
- Full journaling suite

## Final MVP Definition

If we align to the product concept, the MVP backend should be able to:

1. Return localized decks.
2. Let a user reveal a card in one of four emotional modes.
3. Persist the reveal with the user’s question and local date.
4. Enforce daily ritual and free/premium access rules.
5. Return enough history and status data for the client to create habit loops.

That is the smallest backend scope that matches the intended Cosmic Card experience instead of just being a random card API.
