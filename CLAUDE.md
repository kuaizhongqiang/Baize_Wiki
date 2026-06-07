# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**PM: Claude** — Task management, code review, architecture oversight.
**Coder** — Implementation (see [CODING.md](CODING.md) for work standards).
**Tester** — Testing and bug reporting (see [TESTING.md](TESTING.md) for work standards).

## Project

**Baize Wiki (白泽维基)** — An AI Agent-oriented Wiki generation and consumption tool written in Go.

- **Status**: v0.1.0-alpha — 5 phases complete.
- **Goal**: A single-binary CLI that scans a source directory of documents, parses them, builds a searchable Wiki with full-text + semantic search, and exposes it via CLI and MCP.
- **License**: MIT
- **Module**: `github.com/kuaizhongqiang/baize-wiki`

## Build & Development

```bash
go build ./cmd/baize-wiki          # Build to bin/baize-wiki
go test ./...                       # Run all tests
go test ./internal/core/scanner/... # Run single package tests
go vet ./...                        # Static analysis
```

Planned Makefile targets (not yet created):

- `make build` — compile to `bin/baize-wiki`
- `make test` — all tests
- `make lint` — golangci-lint
- `make clean` — clean build artifacts
- `make cross` — cross-compile for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

Dependencies (not yet added to go.mod):

- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/viper` — config loading
- `github.com/yuin/goldmark` — Markdown parsing
- `gopkg.in/yaml.v3` — YAML frontmatter parsing
- `github.com/stretchr/testify` — test assertions
- `golang.org/x/sync/errgroup` — concurrency control

## Architecture

Three-layer architecture with strict dependency direction: **Interface Layer → App Layer → Core Domain**.

```
CLI (cobra) → App Layer (use-case orchestration) → Core Domain (business logic)
```

**Core Domain packages** (under `internal/core/`):

- `model/` — Data structures: Page, Wiki, Config, Section, Link, Frontmatter, sentinel errors
- `scanner/` — Recursive file scanning, `.baizeignore` rule matching, binary detection
- `parser/` — Markdown/frontmatter extraction, Section tree construction, batch parsing
- `generator/` — Wiki directory tree construction for Levels 1/2/3, `_index.md` generation
- `storage/` — File system read/write, meta.json persistence, atomic writes
- `index/` — Full-text search (Phase 3)
- `vector/` — Vector storage interface (Phase 4, stub)

**Other packages:**

- `cmd/baize-wiki/` — Main entry, cobra commands
- `internal/app/` — Use-case orchestration (build, info, serve, manage)
- `internal/mcp/` — MCP protocol (Phase 2)
- `internal/config/` — Config loading from `baize.yaml` + flag override via viper
- `pkg/baize/` — Public API for embedding in Go programs

### Key Concepts

**Level System** (controlled by `--level` flag):

- Level 1 (Flat): Single directory, pages merged by category into 1-10 files
- Level 2 (Structured): One level of subdirectories, pages organized by category
- Level 3 (Deep): Full directory tree (max 3 levels), fine-grained

**Scanning Strategy**:

- Phase 1 default: only `.md`/`.mdx` files
- Phase 2+ `--scan-all`: all non-binary text files (with `.baizeignore` support)
- Binary detection: check first 512 bytes for null bytes or invalid UTF-8

**Config**: `baize.yaml` with viper loading (file → env vars → flags). See [docs/data-model.md](docs/data-model.md#31-配置文件-baizeyaml) for full schema.

## Code Conventions

- Follow Go standard project layout (internal/ for private, pkg/ for public)
- Domain errors: sentinel errors; system errors: `fmt.Errorf + %w`
- Logging: `slog` standard library, passed via context
- Interactive output (progress) → stderr; `--json` results → stdout
- Tests: table-driven + `testify/assert`
- Windows-safe paths: always use `filepath.Join`
- All code, comments, commit messages in English; README/docs in bilingual or Chinese
- Module path: `github.com/kuaizhongqiang/baize-wiki`

## Phases Completed

| Phase | Focus                                                   | Status |
|-------|---------------------------------------------------------|--------|
| 1     | CLI MVP (scan → parse → generate)                      | ✅     |
| 2     | MCP Server (stdio/TCP, 6 tools)                        | ✅     |
| 3     | Full-text search (bleve) + code comments                | ✅     |
| 4     | Vector search + hybrid BM25/vector                      | ✅     |
| 5     | `[[wiki-link]]` cross-links + backlinks                 | ✅     |

## Key Design Docs

| Doc                                                         | What                                                             |
|-------------------------------------------------------------|------------------------------------------------------------------|
| [docs/architecture.md](docs/architecture.md)                | Architecture, package layout, Level system, data flow            |
| [docs/data-model.md](docs/data-model.md)                    | Core structs, config schema, error model, interface contracts    |
| [docs/cli-spec.md](docs/cli-spec.md)                        | Command definitions, flags, exit codes, JSON output format       |
| [docs/mcp-spec.md](docs/mcp-spec.md)                        | MCP tools, protocol details (Phase 2)                            |
| [docs/phase-1-plan.md](docs/phase-1-plan.md)                | Phase 1 — CLI MVP (scan → parse → generate)                      |
| [docs/phase-2-plan.md](docs/phase-2-plan.md)                | Phase 2 — MCP Server                                             |
| [docs/phase-3-plan.md](docs/phase-3-plan.md)                | Phase 3 — Full-text search + code comments (已完成)               |
| [docs/phase-4-plan.md](docs/phase-4-plan.md)                | Phase 4 — Vector search + hybrid retrieval (规划中)               |
| [docs/design-audit.md](docs/design-audit.md)                | Design audit results, residual cleanup items before coding       |
