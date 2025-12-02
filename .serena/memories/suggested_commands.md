# Suggested Commands & Scripts (ttpos-server-go)

> For Serena / Developers to quickly find common commands in this repo.

## Go Main Project `main/`

- **Install Dependencies**
  - `cd main && go mod tidy`
- **Run Main Service (Dev)**
  - `cd main && go run main.go`
- **Run All Tests**
  - `cd main && go test ./...`
- **Generate Swagger Docs**
  - Install: `go install github.com/swaggo/swag/cmd/swag@latest`
  - Generate: `cd main && swag init`
- **Build Production (Script)**
  - Example: `./scripts/build.sh` (if exists)

## Go Middleware `ttpos-bmp/`

- **Install Dependencies**
  - `cd ttpos-bmp && go mod tidy`
- **Run Service**
  - Typical: `cd ttpos-bmp && go run main.go` (Check `main.go` location)
- **DB Migration (BMP Specific)**
  - Scripts in `ttpos-bmp/manifest/sql/`. Check `ttpos-bmp/Makefile` and `MIGRATION_QUICK_START.md`.

## Legacy PHP Admin `admin/`

- **Install Dependencies**
  - `cd admin && composer install`
- **DB Migration**
  - `cd admin && php think migrate:run`
- **Create Migration**
  - `cd admin && php think migrate:create CreateUsersTable`
- **Rollback Migration**
  - `cd admin && php think migrate:rollback`

## WebSocket / redis-proxy (Go)

> Check actual directories like `websocket/`, `redis-proxy/`.

- **Install Dependencies**
  - `cd websocket && go mod tidy`
  - `cd redis-proxy && go mod tidy`
- **Run Service**
  - `cd websocket && go run main.go`
  - `cd redis-proxy && go run main.go`

## Docker & Docker Compose

- **Dev Start (Example)**
  - `docker-compose -f docker-compose.dev.yml up -d`
- **Redis Cluster (Example)**
  - `docker-compose -f docker-compose.dev.redis.yml up -d`
- **Prod Deploy (Example)**
  - `docker-compose -f docker-compose.production.yml up -d`

## Testing

- **Go Unit Tests (Main)**
  - `cd main && go test ./...`
- **Coverage Report**
  - `cd main && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- **Integration Tests**
  - Example: `go test -tags=integration ./tests/...`

## Frontend (Legacy Vue Admin)

> Located in `admin/views/` (Vue 3 + TS + Vite + Element Plus).

- **Install Node Deps**
  - `npm install` (Check `package.json`)
- **Dev / Build**
  - Check `scripts` in `package.json` (e.g., `npm run dev` / `npm run build`).

## General Utils (macOS Darwin)

- Files: `ls`, `cd`, `find`, `grep` / `rg` (ripgrep)
- Git: `git status`, `git diff`, `git log`, `git checkout -b <branch>`, `git commit -m "feat: ..."`, `git push`

## Post-Task Routine

- **Go / PHP / Frontend**
  - Format code (`go fmt ./...`, PSR-2 tools, ESLint/Prettier).
  - Run relevant tests (`go test`, etc.).
  - Update docs (`docs/shared/specs/...`, API docs).
  - Commit message follows `.cursor/rules/version.mdc`.