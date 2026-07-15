.PHONY: dev-backend dev-frontend dev build-backend build-frontend build test test-short test-race migrate-up migrate-down migrate-create migrate-status

# ─── Load .env ─────────────────────────────────────────────

include .env

# ─── Development ───────────────────────────────────────────

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && bun run dev

dev:
	start cmd /c "cd backend && air"
	start cmd /c "cd frontend && bun run dev"

# ─── Build ─────────────────────────────────────────────────

build-frontend:
	cd frontend && bun run build

copy-frontend:
	if exist "backend\cmd\server\static" rmdir /s /q "backend\cmd\server\static"
	xcopy /e /i "frontend\build" "backend\cmd\server\static" > nul

build-backend:
	cd backend && go build -o ../build/server ./cmd/server

build: build-frontend copy-frontend build-backend

# ─── Test ──────────────────────────────────────────────────

test:
	cd backend && go test ./... -v -count=1

test-short:
	cd backend && go test ./... -short -count=1

test-race:
	cd backend && go test ./... -race -count=1

# ─── Database (goose) ──────────────────────────────────────

migrate-up:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-create:
	cd backend && goose -dir migrations create $(name) sql

migrate-status:
	cd backend && goose -dir migrations postgres "$(DATABASE_URL)" status

# ─── Tools ─────────────────────────────────────────────────

air-install:
	go install github.com/air-verse/air@latest

# ─── Clean ─────────────────────────────────────────────────

clean:
	rm -rf build/
