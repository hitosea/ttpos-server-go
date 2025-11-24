# Task Completion Checklist (ttpos-server-go)

> General Checklist for Serena & Developers before finishing Dev / Fix / Integration tasks.

## General (All Tasks)

- **Specs & Requirements**
  - [ ] Relevant Spec / Issue found or created in `docs/shared/specs/story-*/`.
  - [ ] Business requirements in `requirements.md` covered.
  - [ ] Technical `design.md` updated to match implementation.
  - [ ] Current task in `tasks.md` marked as completed with notes.

- **Code Standards**
  - [ ] Go code follows `go-main.mdc` / `ttpos-bmp/.cursor/rules/go-rules.mdc`:
    - main: Clear layering (API/Service/Repo). No direct GORM in Service. API doesn't depend on Repo.
    - ttpos-bmp: `dao/do/entity/service` are auto-generated, not manually edited.
  - [ ] PHP code follows `php.mdc` (Legacy: Controller for params/resp, Service for logic/tx, Model for ORM).
  - [ ] Vue code follows `vue.mdc` (Legacy: Composition API + TS + async/await + Pinia).
  - [ ] Naming, Constants, Error Handling, I18n rules aligned (e.g., `dto.LocaleResponse` for Go Main).

- **Linting & Formatting**
  - [ ] Go: `go fmt ./...` executed.
  - [ ] Go: `go vet ./...` passed (if needed).
  - [ ] PHP: PSR-2 / Project style compliant.
  - [ ] Frontend: ESLint / Prettier passed.

- **Testing**
  - [ ] Unit Tests: Added/Updated for core logic.
  - [ ] `go test ./...` passed in relevant modules.
  - [ ] Integration / API tests passed (if available).
  - [ ] Key scenarios verified locally.

## Database Tasks

- [ ] Schema Changes:
  - [ ] Migration script created:
    - PHP: `admin/database/migrations/` (via `php think migrate:create`).
    - GoFrame: `ttpos-bmp/manifest/sql/` (via project rules).
  - [ ] Field Standards:
    - Time: `int` type, `_time` suffix.
    - Money: `decimal` (e.g., `decimal(20,8)`).
    - Required: `uuid`, `create_time`, `update_time`, `delete_time`.
  - [ ] Idempotency checks (avoid duplicate table/column errors).
  - [ ] Migration verified locally / test env.
  - [ ] Go/PHP Models synced with schema.

## API / Microservice Integration

- [ ] External APIs:
  - [ ] Routes/Params/Resp aligned with `docs/shared/api/*.md` or Spec.
  - [ ] Unified Response format (Go Main: `code/message/data`, data is Object).
  - [ ] Swagger / Apidoc comments complete.
- [ ] Microservices (ttpos-bmp / gRPC / Nacos):
  - [ ] Protobuf follows `proto-rules.mdc` (Naming, Suffixes).
  - [ ] Generated code unmodified. Service registered to Nacos.
  - [ ] Client calls verified in Main or other modules.

## Documentation & Knowledge

- [ ] Docs Updated:
  - [ ] Spec: `docs/shared/specs/story-*/{requirements,design,tasks}.md`.
  - [ ] Architecture: `docs/human/architecture/*.md` (only for arch changes).
  - [ ] API Docs: `docs/shared/api/*.md`.
  - [ ] Troubleshooting: `docs/shared/troubleshooting/*.md` (for complex fixes).
- [ ] Learnings:
  - [ ] Drafted in `docs/agent/templates/graphiti-episode.md`.
  - [ ] Saved to Graphiti (if valuable).

## Commit & Review

- [ ] Git Commit Message follows `version.mdc` (e.g., `feat(order): ...`).
- [ ] PR focused on single requirement.
- [ ] Major changes synced in Team Activity Log (handled by upper-layer tools).