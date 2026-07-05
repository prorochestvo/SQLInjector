# CLAUDE.md

## Build & Run Commands

This module is a **library** (`package sqlinjector`, import path
`github.com/prorochestvo/sqlinjector`) — there is no `main` package and no binary to run.
All drivers are pure-Go, so everything builds and tests with `CGO_ENABLED=0`.

There is no Makefile; use the Go toolchain directly:

```bash
CGO_ENABLED=0 go build ./...                 # compile every package
CGO_ENABLED=0 go vet ./...                   # static checks
gofmt -l -w .                                # format (lists + rewrites)
CGO_ENABLED=0 go test -race ./...            # full test suite
```

Targeted runs use standard `go test` flags, e.g. `CGO_ENABLED=0 go test -race -run 'TestODataExpression/filter_eq' ./` (root) or `-v ./internal/expression/`; add `-coverprofile=cover.out` then `go tool cover -html` for coverage.

**Docker for tests.** SQLite tests run in-memory (pure Go, no Docker). The Postgres/MySQL paths (`NewSandboxOf*`, `internal/sandbox`) spin up ephemeral `testcontainers-go` containers and need a running Docker daemon (they pull `postgres:latest`/`mysql:latest` on first run). `docker-compose.yaml` is a separate local convenience — the suite does **not** read it and consumes no DSN env var.

## Architecture Overview

SQLInjector is a thin toolkit layered on top of the
[SQLBoiler](https://github.com/volatiletech/sqlboiler) ORM. It bundles four concerns a
SQLBoiler-based app repeatedly needs: opening/configuring database sessions across three
dialects, applying schema migrations, building queries (including OData `$filter` parsing)
into SQLBoiler `qm.QueryMod`s, and running actions inside a single commit/rollback
transaction. A generic in-memory repository and disposable database "sandboxes" round it
out so downstream packages can unit-test without a real database. The public surface lives
in the root package; `internal/` holds the dialect-specific machinery.

### Components

| File / package | Role |
|----------------|------|
| `connection.go` | Opens dialect sessions (`NewPostgreSQL`/`NewMySQL`/`NewSQLite3`) returning a `Vault`; pool tuning via `Parameter`s. |
| `expression.go` | Query builder: clause builders + `ODataExpression` (parses `$filter`/`$sort`/`$Limit`/`$Offset`) into `qm.QueryMod`. |
| `migration.go` | `Migrater` plus migration sources (file, embed, memory, struct-reflection); `MultipleMigration` merges + ID-sorts. |
| `repository.go` | Generic `Repository[K,V]` CRUD/batch contract + in-memory `DummyRepository` test double. |
| `sandbox.go` | Disposable migrated databases: Postgres/MySQL via Docker pool, SQLite in-memory. Caller must Close the `Vault`. |
| `transaction.go` | Runs typed `Action[T]`s inside one commit/rollback transaction. |
| `internal/vault.go` | Core interfaces (`Vault`/`Transaction`/`Executor`/`Extractor`) — the `database/sql` surface the library targets. |
| `internal/dialect.go` | `Dialect` enum -> `sql.Open` driver names. |
| `internal/expression/` | Per-clause builders, the `combiner` that flattens them to `qm.QueryMod`, and the OData filter parser. |
| `internal/schema/` | Migration engine: `.sql`/struct parsing, MD5 identities, State/Plan/Up/Down against the tracking table. |
| `internal/sandbox/` | `testcontainers-go` wrappers + `imitator.go` (in-memory expression evaluator backing `DummyRepository`). |
| `internal/transaction/` | Runs an action slice in one `*sql.Tx`, committing/rolling back as a unit. |

### Key Patterns

- **Vault abstraction** — everything programs against `internal.Vault` (Begin/Exec/Query/Close), never a concrete `*sql.DB`, so connections, migrations, transactions and sandboxes are interchangeable and a `*sql.Tx` can stand in for any executor.
- **Expression -> QueryMod** — each clause implements `Expression` (`QueryMod() []qm.QueryMod`); `combiner` concatenates them, so builders, OData parsing and the in-memory imitator share one representation. Gotcha: `Or` only fuses concrete `*expression.Where` operands; mixed types fall back to AND-combining.
- **Generic repository + in-memory double** — `Repository[K constraints.Ordered, V any]` is the persistence contract; `DummyRepository` satisfies it in memory (keyed by a reflected `id`-style field, filtered via the sandbox imitator) so service/handler code is tested without a DB.
- **Disposable sandboxes** — a fully-migrated throwaway DB per call (pooled `testcontainers` container, isolated by a generated schema/db name, or in-memory SQLite); the caller must Close the returned `Vault` to free Docker resources.

### Database

Three engines, all **pure-Go (no CGO)**: PostgreSQL (`lib/pq`), MySQL (`go-sql-driver/mysql`), SQLite3 (`glebarez/sqlite`, a `modernc.org/sqlite` driver). The `glebarez` choice is what keeps the build CGO-free — never swap to CGO `mattn/go-sqlite3` (see Constraints).

**Schema & migrations** (driven by `Migrater`, no external tool):

- An `Instruction` = `ID` + up/down SQL; identity is the **MD5 of `up + "\n" + down`**. `.sql` files use `--- #migrate:up` / `--- #migrate:down` markers (`MigrationCommandPrefix`); other `--` lines are comments. For file/embed sources the **filename is the migration ID** and instructions apply in **lexicographic ID order** — no enforced `YYYYMM`/counter convention, so name files to sort correctly.
- Applied migrations are tracked in `_migrations` (override via `SetTableName`; columns `id, md5, applied_at`). `Up` applies all pending in **one transaction**; `Down` reverses the latest applied tail; `Clean` reverses everything.
- **Drift**: `State` marks each instruction applied `[X]` / pending `[ ]` / `ERR` (stored MD5 no longer matches code = migration edited after apply). `Plan` errors if the DB holds migration IDs the code doesn't know.
- `NewStructMigration` reflects `boil` field tags into `CREATE TABLE` DDL; per-dialect type mapping in `internal/schema/instruction.go` (`fieldType.toSqlType`).

> Caution: the migration-table `INSERT`/`DELETE` statements are string-formatted, not
> parameterized (IDs/MD5s are stripped of quotes first). Migration IDs are developer-controlled,
> not user input — keep it that way.

### Key Dependencies

Direct deps live in `go.mod`: SQLBoiler (+ `queries/qm`, the expression builder's target), `volatiletech/null` (nullable types the DDL generator recognises), the three pure-Go drivers, `testcontainers-go`, `testify`, and `golang.org/x/exp/constraints` (the `Ordered` key constraint). The pure-Go SQLite choice is the load-bearing one — see Constraints.

### Error Handling

Library surfaces return plain `error` (no `PublicError` user/internal split). Wrap with `fmt.Errorf("...: %w", err)` to preserve the cause for `errors.Is`/`As`; aggregate teardown failures with `errors.Join` (e.g. an op error joined with the `Rollback()`/`Close()` error in `defer`) so a secondary failure is never swallowed. No exported sentinels or custom error types — callers match on wrapped causes.

## Code Organization

Public surface lives in the root package; `internal/` holds the dialect-specific machinery. Prefer `internal/` over a public `pkg/` — there is no external (out-of-module) consumer to promise an API to. Organize by concern (flat, per-concern split), not by launcher. Distinguish coincidental similarity (duplicate the few trivial lines, let each site diverge) from a genuine cross-cutting invariant (centralize once).

## File Declaration Order

Public surface at the top, private internals at the bottom — a reader sees everything important first. Per file built around one object:

1. Exported `const`/`var` + `New<Object>` constructor(s).
2. The struct definition.
3. Its methods (prefer alphabetical).
4. Unexported `const`/`var`.
5. Auxiliary unexported support structs.
6. Unexported methods/functions (prefer alphabetical).

Multiple structs: same layout, primary struct first (two large objects usually means split the file). No-object files (free functions + a config type): same spirit — exported on top, then unexported. Treat a violating file as something to fix.

## Constraints

- **Forbidden imports**: never introduce `github.com/mattn/go-sqlite3` (or any other
  CGO-dependent driver) — it would force `CGO_ENABLED=1` and break the pure-Go build. The
  SQLite path must stay on `github.com/glebarez/sqlite`. More generally, keep `go.mod` free
  of CGO-dependent modules.
- **Testing**: Use `github.com/stretchr/testify`; run tests with `-race`; parallel
  subtests preferred where there's no shared mutable state. Tests that touch
  PostgreSQL/MySQL need a running Docker daemon (`testcontainers-go`); SQLite tests do not.
- **One `Test*` per method, scenarios as subtests**: each tested method/function gets
  exactly one top-level test function named after it (e.g. `TestEncode` for `Encode`),
  and every scenario for that method lives as a `t.Run("descriptive name", ...)`
  subtest inside it. Do **not** create separate top-level tests like
  `TestEncode_EmptyInput`, `TestEncode_Unicode`, `TestEncode_Error` — these belong
  as subtests of a single `TestEncode`. Methods on a type follow the same rule with
  the standard `TestType_Method` form (e.g. `TestUser_Validate`).
  ```go
  func TestEncode(t *testing.T) {
      t.Parallel()
      t.Run("empty input returns empty string", func(t *testing.T) { t.Parallel(); /* ... */ })
      t.Run("returns error on invalid byte", func(t *testing.T) { t.Parallel(); /* ... */ })
  }
  ```
- **No CGO**: `CGO_ENABLED=0` must be set for all build and test commands. The whole point
  of the SQLite driver choice is to keep this true.
- **Compile-time interface checks**: Every mock/stub struct in test files must have a
  `var _ interfaceName = &mockStruct{}` assertion at the top of the file.
- **No section-divider comments**: Do not use `// --- section ---` or `// ----` style
  separator comments. Let the code structure speak for itself.
- **No skipped errors**: Never use `_` to discard error return values in production or
  test code. Always capture the error and assert/check it. The only exceptions are
  `fmt.Fprint*` writes to loggers, `Rollback()` calls in error-recovery paths, and
  resource `.Close()` in `t.Cleanup` / `defer`.
- **Comments**: all comments are in English and start with a lowercase first word
  (e.g. `// wrap the driver error so callers can match on it`).
- **Godoc on exported identifiers**: every exported Type/Func/Method/Var/Const gets a doc comment starting with its name and ending with a period; exactly one `// Package <name> ...` per package. Skip it if it would only restate the signature. Document concurrency guarantees, lifecycle contracts ("caller must Close"), and error-sentinel conditions; preserve existing WHY-comments verbatim; don't bulk-comment private helpers.
- **Scratch in `./tmp/`**: keep throwaway artifacts, fixtures, and intermediate files in
  `./tmp/` rather than the repo root.

## Planning Workflow

Non-trivial work gets a Markdown plan file in `plans/` before code. One plan per concern; if implementation diverges, update the plan before completing it.

- **Active** `plans/NNN-slug.md` — next `NNN` = highest existing prefix across `plans/`, `completed/`, `history/`, +1.
- **Completed** `plans/completed/YYMMDD.NNNN.slug.md` — `NNNN` resets daily. Move here (`mv`) only once acceptance criteria are met and `make test` passes.
- **Archived** `plans/history/` — abandoned/superseded plans keep their original `NNN-` name.

Plan sections: Overview, Assumptions, Tasks (each: Description, Acceptance Criteria, Pitfalls & edge cases, Complexity Easy/Medium/Hard), Execution Order, Risks, Trade-offs.

## Agent Pipeline

Non-trivial tasks run architect -> engineer -> reviewer, no stage skipped:

1. `gocode-architect` — writes/updates the plan file in `plans/` (see Planning Workflow).
2. `gocode-engineer` — implements the plan and writes tests for new code. `make test` must be green before review; if red, hand logs to `gocode-testdoctor` first.
3. `gocode-reviewer` x3 in parallel (one message, three tool calls), each a self-contained prompt naming its lens and what to SKIP: (A) correctness & tests, (B) security & operations, (C) performance & architecture. Deliverable per lens: P0/P1/P2/P3 findings with `file:line` + patch sketch, <=600 words. The full three-lens fan-out is mandatory on the FIRST review.

`gocode-testdoctor` is invoked on demand whenever tests fail; it makes the minimal fix, no redesign.

The orchestrator (main session) synthesises the three reports, resolves conflicts (names the rejected suggestion; user has final say), and gates completion. P0/P1 go back to the engineer; after a fix, re-review is a SINGLE pass scoped to the changed lines, not another fan-out. Once every P0/P1 is resolved, move the plan to `plans/completed/`.
