APP_NAME=cosmic-card-api
MAIN=./cmd/api

.PHONY: run
run:
	go run $(MAIN)

.PHONY: build
build:
	go build -o bin/$(APP_NAME) $(MAIN)

.PHONY: start
start: build
	./bin/$(APP_NAME)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: dev
dev:
	air