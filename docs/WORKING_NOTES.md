# Cosmic Card Working Notes

## Purpose

This file is the quick-start context for continuing work on `cosmic-card-api` without re-reading the whole codebase.

Read this first, then open deeper docs only when needed.

## Current Product Scope

The backend currently supports:

- deck listing
- single-card reveal flow
- draw persistence

Product framing:

- reflection over prediction
- lightweight ritual loop
- modes:
  - `daily`
  - `guidance`
  - `support`
  - `reflection`

Full product/API direction lives in [API_SPEC.md](/Users/benz/Document/09_cosmic_card/cosmic-card-api/docs/API_SPEC.md).

## Current Architecture

Main folders:

- [cmd/api/main.go](/Users/benz/Document/09_cosmic_card/cosmic-card-api/cmd/api/main.go)
  - application composition
  - builds config, db, repos, services, handlers, router
- [internal/router/router.go](/Users/benz/Document/09_cosmic_card/cosmic-card-api/internal/router/router.go)
  - HTTP route registration only
  - should stay thin
- `internal/modules/decks/*`
  - deck listing module
- `internal/modules/draws/*`
  - draw reveal module
- [internal/config/config.go](/Users/benz/Document/09_cosmic_card/cosmic-card-api/internal/config/config.go)
  - env loading and validation
- [internal/database/postgres.go](/Users/benz/Document/09_cosmic_card/cosmic-card-api/internal/database/postgres.go)
  - postgres pool creation

Module pattern:

- `handler`
  - HTTP input/output
- `service`
  - business rules and defaults
- `repository`
  - SQL and persistence
- `model`
  - transport/domain structs

## Important Recent Decisions

Recent refactor introduced:

- dependency composition in `main`, not `router`
- `router` receives ready-made handlers
- `config.Load()` returns `(Config, error)`
- `database.NewPostgresPool()` returns `(*pgxpool.Pool, error)`
- services depend on repository interfaces, not concrete repository types

Why this matters:

- easier unit tests
- easier handler/router tests
- less hidden process exit behavior

## Current Routes

Infra:

- `GET /health`
- `GET /info`
- `GET /metrics`

API:

- `GET /api/v1/decks`
- `POST /api/v1/draws/reveal`

## What Exists vs What Is Still Missing

Already in place:

- localized deck list
- reveal one random card by mode
- persist reveal in `user_draws`

Still missing for MVP:

- explicit `drawMode` enum validation
- daily draw limit
- free vs premium rules
- history endpoint
- today-status endpoint
- entitlements endpoint
- actual automated tests

## Recommended Next Work

Suggested order:

1. add service unit tests for `draws`
2. add service unit tests for `decks`
3. add handler tests with `httptest`
4. add draw mode validation
5. add daily limit rule
6. add history and today-status APIs

## Testing Notes

Because local sandbox/cache permissions can interfere with default Go cache paths, this command is the safe one to reuse:

```bash
env GOCACHE=/private/tmp/go-build-cache go test ./...
```

Useful commands:

```bash
gofmt -w ./...
```

```bash
git log --oneline -5
```

## Commit Landmarks

Useful recent commits:

- `3edd70b` `feat: add decks api`
- `e5b98cd` `Add draw reveal flow and API spec`
- `096131c` `Refactor wiring for testable services and startup`

## Working Rules For Future Changes

Try to preserve these boundaries:

- keep `router` thin
- keep SQL in repositories
- keep defaults/validation in services
- return errors upward instead of calling `log.Fatal` in reusable code
- prefer small repository interfaces for service tests

## If Returning To This Repo Later

Read in this order:

1. this file
2. [API_SPEC.md](/Users/benz/Document/09_cosmic_card/cosmic-card-api/docs/API_SPEC.md)
3. [cmd/api/main.go](/Users/benz/Document/09_cosmic_card/cosmic-card-api/cmd/api/main.go)
4. the specific module being changed

That should be enough to resume work quickly.
