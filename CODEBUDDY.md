# CODEBUDDY.md

This file provides guidance to CodeBuddy Code when working with code in this repository.

## Project Overview

**Baize Wiki (白泽维基)** — An AI Agent-oriented Wiki generation and consumption tool written in Go.

- **Module**: `github.com/kuaizhongqiang/baize-wiki` (Go 1.24.2)
- **Status**: Phase 1 (MVP) — design docs audited and approved, Go code not yet implemented. Start with M1 (project scaffold).
- **Goal**: Single-binary CLI that scans a source directory of documents (`.md`/`.mdx`), parses them, and generates a structured Wiki output (folder + MD files) at configurable complexity Levels 1/2/3.
- **License**: MIT

## Build & Development Commands

```bash
go build ./cmd/baize-wiki                    # Build to bin/baize-wiki
go test ./...                                 # Run all tests
go test ./internal/core/scanner/... -v        # Run single package tests (verbose)
go test ./internal/core/scanner/... -run TestScanDir -v  # Run a specific test function
go vet ./...                                  # Static analysis
```

Makefile targets (planned, not yet created):
- `make build`, `make test`, `make lint` (golangci-lint), `make clean`, `make cross`

## Architecture

Three-layer architecture with strict dependency direction: **Interface Layer → App Layer → Core Domain** (upper layers depend on lower ones; lower layers never depend on upper layers).

```
CLI (cobra) → App Layer (use-case orchestration) → Core Domain (business logic)
```

### Core Domain Packages (internal/core/)

| Package | Purpose | Phase |
|---------|---------|-------|
| `model/` | Data structures: Page, Wiki, Config, Section, Link, Frontmatter, sentinel errors | M2 |
| `scanner/` | Recursive file scanning, `.baizeignore` rule matching, binary detection (first 512 bytes: null byte or invalid UTF-8) | M3 |
| `parser/` | Markdown/frontmatter extraction (yaml.v3), Section tree construction, batch parsing (goldmark) | M4 |
| `generator/` | Wiki directory tree construction for Levels 1/2/3, `_index.md` generation, Level 1 merge algorithm (group by category → sort by weight/filename → concatenate → split at 50KB) | M5 |
| `storage/` | File system read/write, `meta.json` persistence, atomic writes | M5 |
| `index/` | Full-text search (bleve) — Phase 3 | |
| `vector/` | Vector storage interface stub — Phase 4 | |

### Other Packages

| Package | Purpose | Phase |
|---------|---------|-------|
| `cmd/baize-wiki/` | Main entry, cobra commands (init, build, info, search, mcp) | M1 |
| `internal/app/` | Use-case orchestration (build, info, search, serve, manage) | M1+ |
| `internal/mcp/` | MCP protocol server (stdio/TCP, JSON-RPC 2.0, MCP 2025-03-26) — Phase 2 | |
| `internal/config/` | Config loading from `baize.yaml` via viper (file → env vars → flags) | M1 |
| `pkg/baize/` | Public API for embedding in Go programs — Phase 2+ | |

### Core Data Flow

```
Source Dir → Scanner → File List → Parser → Pages → Generator → Wiki Dir (+ .baize/meta.json)
```

### Level System (controlled by `--level` flag)

- **Level 1 (Flat)**: Single directory, pages merged by category into 1-10 files (split at 50KB)
- **Level 2 (Structured)**: One level of subdirectories, pages organized by category
- **Level 3 (Deep)**: Full directory tree (max 3 levels), fine-grained

### Scanning Strategy

- Phase 1 default: only `.md`/`.mdx` files
- Phase 2+ `--scan-all`: all non-binary text files (with `.baizeignore` support)
- Binary detection: check first 512 bytes for null bytes or invalid UTF-8 via `utf8.Valid`

## Key Design Docs

| Doc | Content |
|-----|---------|
| [docs/architecture.md](docs/architecture.md) | Architecture, package layout, Level system, data flow, non-functional goals |
| [docs/data-model.md](docs/data-model.md) | Core structs, config schema, interface contracts, sentinel errors |
| [docs/cli-spec.md](docs/cli-spec.md) | Command definitions (init/build/info/search/mcp), flags, exit codes, JSON output, env vars |
| [docs/mcp-spec.md](docs/mcp-spec.md) | 6 MCP tools (wiki_build, wiki_search, wiki_read, wiki_list, wiki_add, wiki_stats), error codes, resources |
| [docs/phase-1-plan.md](docs/phase-1-plan.md) | 7 milestones with task decomposition, dependencies, risks |
| [docs/design-audit.md](docs/design-audit.md) | Design audit results (90% readiness), 5 residual micro-fixes before coding |

## Phase 1 Milestones

| # | Milestone | What |
|---|-----------|------|
| M1 | Project scaffold | Go module, Makefile, goreleaser, cobra CLI stubs |
| M2 | Core model | model/ package, config, errors + tests |
| M3 | Scanner | scanner/ package (ignore rules, binary detection) + tests |
| M4 | Parser | parser/ package (frontmatter, markdown, plain text) + tests |
| M5 | Generator | generator/ (Level 1/2/3) + storage/ + tests |
| M6 | CLI integration | init/build/info commands wired end-to-end |
| M7 | Polish | README, examples, CI, benchmarks, tag v0.1.0-alpha |

Full build strategy: no incremental/cache logic; each `build` is a complete rebuild.

## Code Conventions

- Follow Go standard project layout (`internal/` for private, `pkg/` for public)
- Domain errors: sentinel errors; system errors: `fmt.Errorf + %w`
- Logging: `slog` standard library, passed via context
- Interactive output (progress) → stderr; `--json` results → stdout
- Tests: table-driven + `testify/assert`
- Windows-safe paths: always use `filepath.Join`
- All code, comments, commit messages in English
- Module path: `github.com/kuaizhongqiang/baize-wiki`

## Test Data

Test files live in `testdata/source/`:
- `docs/` — Markdown with/without frontmatter, plain text (.txt)
- `scripts/` — C# source files for `--scan-all` mode testing
- `binary-test/` — Binary file with null bytes, `.DS_Store`, `Thumbs.db` for scanner tests

Use `scripts/prepare-testdata.sh` to regenerate test data from Unity projects.

## Ignore Rules

`.baizeignore` at project root uses gitignore-compatible patterns. Standard ignores: `.git/`, `node_modules/`, `vendor/`, `__pycache__/`, `dist/`, `build/`, `bin/`, `*.exe`, `*.dll`, `*.so`, `.DS_Store`, `Thumbs.db`, `.idea/`, `.vscode/`.
