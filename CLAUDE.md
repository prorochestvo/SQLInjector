# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

Targeted test runs (real package paths):
```bash
# Root package, single top-level test
CGO_ENABLED=0 go test -race -run TestODataExpression ./

# A subtest
CGO_ENABLED=0 go test -race -run 'TestODataExpression/filter_eq' ./

# An internal package, verbose
CGO_ENABLED=0 go test -race -v ./internal/expression/

# Coverage
CGO_ENABLED=0 go test -race -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

**Docker requirement for tests.** The SQLite-backed tests run fully in-memory (pure Go,
no Docker). The PostgreSQL/MySQL paths — `NewSandboxOf*` and anything exercising the
`internal/sandbox` pool — spin up **ephemeral containers via `testcontainers-go`** and
therefore need a **running Docker daemon**; they pull `postgres:latest` / `mysql:latest`
on first run. `docker-compose.yaml` is a separate convenience for standing up a local
`postgres:alpine` (database `sqlinjector`, port `5432`); the test suite does **not** read
it (no DSN env var is consumed) — it manages its own containers.

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
| `connection.go` | Opens sessions: `NewPostgreSQL` / `NewMySQL` / `NewSQLite3` return a `Vault` (wraps `internal.Vault`). Connection-pool tuning via `Parameter`s (`MaxOpenConnection`, `MaxIdleConnection`, `ConnectionMaxLifetime`, `ConnectionMaxIdleTime`). `Stats` / `Burden` report pool health. |
| `expression.go` | Query builder. `Where`/`Or`/`GroupBy`/`OrderBy`/`Limit`/`Offset`/`Relation` produce `Expression`s that emit `qm.QueryMod`. `ODataExpression(*url.Values, ...)` parses OData-style query params (`$filter`, `$sort`, `$Limit`, `$Offset`) into a combined expression. Operator/direction constants (`Equal`, `In`, `Contains`, `Ascending`, …) re-export the `internal/expression` enums. |
| `migration.go` | `Migrater` (`Up`/`Down`/`Plan`/`State`/`Clean`, `SetTableName`). Migration sources: `NewFileMigration` (folder), `NewEmbedMigration` (`embed.FS`), `NewMemoryMigration` (raw up/down strings), `NewStructMigration` (generate a `CREATE TABLE` from a SQLBoiler struct via reflection). `MultipleMigration` merges and ID-sorts. |
| `repository.go` | `Repository[DATAKEY, DATASET]` generic CRUD + batch interface, plus `DummyRepository`, an in-memory implementation (with `OnBefore*`/`OnAfter*` hooks) used as a test double. `Expression`s are evaluated in-memory by the sandbox imitator. |
| `sandbox.go` | `NewSandboxOfPostgreSQL` / `NewSandboxOfMySQL` (Docker containers via a shared `internal/sandbox` pool) and `NewSandboxOfSQLite3` (in-memory). Each creates an isolated schema/database, runs the supplied migrations, and returns a ready `Vault` for tests. |
| `transaction.go` | `Commit[T]` / `Rollback[T]` run one or more typed `Action[T]` inside a single transaction; `TransactionCommit` / `TransactionRollback` are the untyped variants over `internal/transaction`. |
| `internal/vault.go` | Core interfaces: `Vault`, `Transaction`, `Executor`, `Extractor` — the `database/sql` surface the rest of the library programs against. |
| `internal/dialect.go` | `Dialect` enum (`postgres`, `mysql`, `sqlite`) — the `sql.Open` driver names. |
| `internal/expression/` | Per-clause builders (`where`, `orderby`, `groupby`, `limit`, `offset`, `select`, `table`, `relation`) and the `combiner` that flattens them into `qm.QueryMod`s. Also the OData filter parser. |
| `internal/schema/` | The migration engine. `instruction.go` parses `.sql` files (direction markers), reflects struct→table DDL, and computes MD5 identities; `migration.go` implements `State`/`Plan`/`Up`/`Down` against the tracking table. |
| `internal/sandbox/` | `testcontainers-go` wrappers (`container.go`, `pool.go`, `vault.go`) and `imitator.go` — an in-memory evaluator that applies `where`/`groupby`/`orderby` expressions to a `map`, backing `DummyRepository`. |
| `internal/transaction/` | `executor.go` — runs a slice of actions in one `*sql.Tx`, committing or rolling back as a unit. |

### Key Patterns

- **Vault abstraction** — every consumer programs against `internal.Vault` (the
  `database/sql` Begin/Exec/Query/Close surface), never a concrete `*sql.DB`. The root
  `Vault` interface just re-exports it, so connections, migrations, transactions and
  sandboxes are all interchangeable and a `*sql.Tx` can stand in wherever an executor is
  expected.
- **Expression → QueryMod** — each clause type implements `Expression` (`QueryMod() []qm.QueryMod`).
  `combiner` concatenates them, so the public builders, OData parsing, and the in-memory
  imitator all share one representation. Note: `Or` only fuses operands that are concrete
  `*expression.Where`; mixed types fall back to AND-combining.
- **Generic repository + in-memory double** — `Repository[DATAKEY constraints.Ordered, DATASET any]`
  is the persistence contract; `DummyRepository` satisfies it entirely in memory (keyed by a
  reflected `id`-style field, filtered via the sandbox imitator) so service/handler code can
  be tested without a database.
- **Disposable sandboxes** — tests get a fully-migrated throwaway database per call:
  PostgreSQL/MySQL via a pooled `testcontainers` container (isolated by a generated
  schema/database name), SQLite via an in-memory connection. The caller must close the
  returned `Vault` to free Docker resources.

### Database

Supported engines and drivers (all **pure-Go — CGO is not required**, builds run with
`CGO_ENABLED=0`):

- PostgreSQL — `github.com/lib/pq` (driver name `postgres`)
- MySQL — `github.com/go-sql-driver/mysql` (driver name `mysql`)
- SQLite3 — `github.com/glebarez/sqlite` (driver name `sqlite`), a pure-Go,
  `modernc.org/sqlite`-based driver — this is what keeps the module CGO-free; do **not**
  switch to the CGO `mattn/go-sqlite3` driver.

**Schema & migrations.** Migrations are driven by `Migrater`, not an external tool:

- An `Instruction` carries an `ID` plus `up`/`down` SQL; its identity is the **MD5 of
  `up + "\n" + down`**. `.sql` files are parsed with direction markers
  `--- #migrate:up` / `--- #migrate:down` (prefix const `MigrationCommandPrefix`); other
  `--` lines are comments. For file/embed sources the **filename is the migration ID**, and
  instructions apply in **lexicographic ID order** — there is no enforced
  `YYYYMM`/counter convention, so name files so they sort correctly.
- Applied migrations are tracked in a table (default `_migrations`, override via
  `Migrater.SetTableName`) holding `id, md5, applied_at`. `Up` applies every pending
  instruction inside **one transaction** and records its hash; `Down` reverses the most
  recent applied tail; `Clean` reverses everything.
- **Drift detection**: `State` classifies instructions as applied `[X]`, pending `[ ]`, or
  `ERR` (an ID present in the table whose stored MD5 no longer matches the code — i.e. a
  migration was edited after being applied). `Plan` errors out if the database contains
  migration IDs the code doesn't know about.
- `NewStructMigration` generates `CREATE TABLE` DDL from a SQLBoiler struct by reflecting
  `boil` field tags; per-dialect SQL type mapping lives in `internal/schema/instruction.go`
  (`fieldType.toSqlType`).

> Caution: the migration-table `INSERT`/`DELETE` statements are string-formatted, not
> parameterized (IDs/MD5s are stripped of quotes first). Migration IDs are developer-controlled,
> not user input — keep it that way.

### Key Dependencies

- `github.com/volatiletech/sqlboiler/v4` — ORM and `queries/qm` query mods that the
  expression builder targets.
- `github.com/volatiletech/null/v8` — nullable column types, recognised by the
  struct→table DDL generator.
- `github.com/lib/pq` — PostgreSQL driver (pure Go).
- `github.com/go-sql-driver/mysql` — MySQL driver (pure Go).
- `github.com/glebarez/sqlite` — pure-Go SQLite driver (`modernc.org/sqlite`), no CGO.
- `github.com/testcontainers/testcontainers-go` + `github.com/docker/go-connections` —
  ephemeral PostgreSQL/MySQL containers for the sandbox (test-time, needs Docker).
- `github.com/twinj/uuid` — UUID generation.
- `github.com/stretchr/testify` — test assertions.
- `golang.org/x/exp/constraints` — `Ordered` constraint for generic repository keys.
- Go version: `1.22.4`

### Error Handling

There is no `PublicError`-style user/internal split — this is a library, and all surfaces
return plain `error`. Conventions in use:

- Wrap with context via `fmt.Errorf("...: %w", err)`, preserving the cause for `errors.Is`/`As`.
- Aggregate cleanup/teardown failures with `errors.Join` (e.g. a failed operation joined with
  the `Rollback()`/`Close()` error in `defer`), so a secondary failure is never swallowed.
- No exported sentinel error values or custom error types; callers match on wrapped causes.

## Code Organization Principles

These rules govern *where code lives*. Apply them by default; treat a violation as
something to flag, not silently accept.

### Package placement follows consumption, not aspiration

- Code shared by **multiple** binaries/entry points belongs in the shared tree
  (`internal/`).
- Code with **exactly one** consumer belongs **next to that consumer**
  (`cmd/<binary>/`), not in the shared tree.
- Prefer the private location (`internal/`) over a public one (`pkg/`) unless there
  is a **real external (out-of-module) consumer**. Don't promise a public API
  surface the project doesn't actually provide.
- **Why:** the shared tree is for the genuinely shared layers of one app. Putting
  single-consumer code (or a separate app) there bloats it and implies a contract
  that doesn't exist; a `pkg/` package nobody outside the module imports is dead
  weight. Before placing or keeping a package in the shared or public tree, check
  who actually imports it — one consumer means co-locate, no external module means
  keep it private. Never keep something in the shared tree just because it's
  "reusable in principle"; treat such a move as its own deliberate refactor.

### Deduplication is not a goal in itself

- Distinguish **coincidental similarity** (looks alike today but must be free to
  diverge) from a **genuine cross-cutting invariant**. Coincidental similarity →
  duplicate the few trivial lines and let each site evolve. A true invariant →
  centralize it once, where it belongs.
- Do **not** build a shared `bootstrap` / `startup` / `wiring` layer for multiple
  binaries just because their startup looks similar — inline it per entry point
  (`cmd/<binary>/main.go`) so each stays free to diverge (different DBs,
  dependencies, config).
- Before extracting a helper, check whether the only thing being shared is already
  captured elsewhere (e.g. already a one-line call) — if so, don't wrap it. An
  abstraction can re-introduce the very complexity it pretends to hide (e.g. a
  returning constructor needs error-cleanup that an inline fatal-and-exit path
  simply doesn't).
- **Why:** premature extraction imposes a contract where code should diverge. Dedup
  earns its place only when it names a non-obvious invariant, removes a real
  divergence risk, or cuts genuine cognitive load — not because two snippets look
  alike.

### Business logic is organized by concern, not by launcher

- Business-logic packages are judged by being **simple and isolated**, regardless of
  which binary runs them or how they are launched ("how it starts is not the
  package's concern"). Keep a flat, per-concern split.
- Do **not** reorganize business logic by runtime-vs-operator, by deployment, or by
  consuming binary.
- **Why:** grouping by launcher couples organization to deployment, which changes;
  cohesion by concern is stabler. Isolation + simplicity is the real quality bar.

## File Declaration Order

Order the top-level declarations in each `*.go` file so the important, public surface
is at the top and private internals are hidden at the bottom. A reader should see
everything important first; scanning the file should not require digging.

For a file built around one object:

1. Exported `const` and `var`, plus the `New<Object>` constructor(s). These come first
   because they are what you need to create and use the object — the first thing a
   reader looks for.
2. The object's struct definition.
3. The object's methods (prefer alphabetical order; not mandatory).
4. Unexported `const` and `var`.
5. Auxiliary/helper structs (unexported support types) — placed between the unexported
   vars/consts and the unexported methods.
6. Unexported methods/functions (prefer alphabetical order; not mandatory).

- **Multiple structs in one file:** keep the same layout but put the primary ("main")
  struct first. A combined layout is acceptable but very rare — two large objects in one
  file usually means the file should be split into two.
- **Files with no object** (free functions plus a config/data type): apply the same
  spirit — exported type(s) and function(s) on top, then unexported consts/vars, then
  auxiliary structs, then unexported helper functions.

Treat a file that violates this order as something to fix.

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

      t.Run("empty input returns empty string", func(t *testing.T) {
          t.Parallel()
          // ...
      })

      t.Run("unicode is preserved", func(t *testing.T) {
          t.Parallel()
          // ...
      })

      t.Run("returns error on invalid byte", func(t *testing.T) {
          t.Parallel()
          // ...
      })
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
- **Godoc on exported identifiers**: Every exported identifier (Type, Func, Method,
  Var, Const) gets a doc comment that starts with the identifier name and ends with
  a period — e.g. `// Encode returns the base64-encoded form of v.` Each package
  has exactly one `// Package <name> ...` declaration. Skip the comment entirely if
  it would only restate the signature — no `// Foo is a Foo.` fluff. Document
  concurrency guarantees, constructor lifecycle contracts ("caller must Close"), and
  error sentinel conditions. Preserve existing WHY-comments verbatim; do not overwrite
  a substantive comment with a generic restatement. Unexported symbols only get
  comments when intent is non-obvious — do not bulk-add comments to private helpers.
- **Scratch in `./tmp/`**: keep throwaway artifacts, fixtures, and intermediate files in
  `./tmp/` rather than the repo root.

## Planning Workflow

All non-trivial work is tracked as a Markdown plan file before implementation begins.

### Directory layout

```
plans/
├── NNN-task-slug.md     # active / in-progress plans (e.g. 001-fix-auth.md)
├── completed/           # plans for fully shipped tasks (e.g. 260422.0001.fix-auth.md)
└── history/             # archived / cancelled plans
```

### File naming

- **Active plans (`plans/`)** — zero-padded sequential index + kebab-case slug:
  `NNN-description.md` (e.g. `001-fix-unauthorized-middleware.md`, `002-add-rate-limiting.md`).
  Pick the next number by checking the highest existing prefix across `plans/`, `plans/completed/`,
  and `plans/history/`.

- **Completed plans (`plans/completed/`)** — date prefix + zero-padded daily index (4 digits) + slug:
  `YYMMDD.NNNN.description.md` (e.g. `260422.0001.fix-unauthorized-middleware.md`).
  `NNNN` resets to `0001` each day and increments for each additional completion on that day.

- **Archived plans (`plans/history/`)** — keep the original `NNN-` filename from `plans/`.

### Lifecycle

1. **Create** — before touching code, produce a plan file in `plans/` using the `NNN-slug.md`
   naming convention described above.
2. **Implement** — work through the tasks defined in the plan. The plan file stays in
   `plans/` while work is in progress.
3. **Complete** — once every acceptance criterion is met and `make test` passes, rename
   and move the file to `plans/completed/` using the date-based convention:
   ```bash
   mv plans/001-fix-auth.md plans/completed/260422.0001.fix-auth.md
   ```
4. **Archive** — if a plan is abandoned or superseded without being fully implemented or
   if we need to save intermediate data or task execution logs, move it to `plans/history/` instead.

### Plan file format

Every plan file follows this structure:

```markdown
# Task Breakdown

## Overview
## Assumptions
## Tasks
### Task N: <Title>
- Description:
- Acceptance Criteria:
- Pitfalls & edge cases:
- Complexity: Easy / Medium / Hard
## Execution Order
## Risks
## Trade-offs
```

### Rules

- **One plan per concern.** Don't bundle unrelated changes in a single plan file.
- **Plan before code.** Claude must create (or confirm an existing) plan file before
  writing or modifying any source files.
- **Keep plans honest.** If implementation diverges from the plan, update the plan file
  before moving it to `completed/`.
- **Slug matches intent.** The description part of the filename should be readable at a glance:
  `002-add-rate-limiting.md`, `003-migrate-sqlite-to-postgres.md`, not `004-task.md`.

## Agent Pipeline

All non-trivial tasks follow a three-stage pipeline using specialized agents. The
review stage fans out to **three `gocode-reviewer` instances running in parallel**,
each with a distinct lens. A separate `gocode-testdoctor` agent is invoked
on-demand whenever tests fail, at any stage.

```
User describes task
    ↓
1. gocode-architect
    → Creates plan file at plans/NNN-slug.md (see Planning Workflow)
    ↓
2. gocode-engineer
    → Implements the tasks defined in the plan
    ↓
3. gocode-reviewer × 3 (run in parallel — single message, three tool calls)
    Lens A: correctness & tests — bugs, races, edge cases, error paths,
            context propagation, resource cleanup, test coverage,
            test structure (one Test* per method with subtests),
            scenario completeness, fixtures
    Lens B: security & operations — input validation, auth boundaries,
            secrets handling, injection (SQL, command, template),
            observability (logs, metrics, traces), log volume,
            operator/runbook UX
    Lens C: performance & architecture — allocations, blocking I/O,
            goroutine/resource leaks, layer boundaries, dependency
            direction, API contracts (breaking changes, exported
            surface stability), interface scope, future-proofing
    ↓
   Orchestrator synthesises all three reports, deduplicates findings,
   resolves conflicts (e.g. one reviewer flags as P0 what another
   accepts as a trade-off), and presents the merged punch list to the user.
    ↓
  ❌ P0/P1 found?  → Back to gocode-engineer with the consolidated findings.
                             After fix, run ONE targeted reviewer pass on the changed
                             lines (not all 3 again) before re-approval.
  ⚠️  Tests failing?        → gocode-testdoctor diagnoses and patches, then rerun the
                             targeted reviewer pass.
  ✅ All three approve?     → Orchestrator moves the plan: mv plans/NNN-slug.md
                             plans/completed/YYMMDD.NNNN.slug.md
```

### Agent responsibilities

| Agent | Owns | Output |
|-------|------|--------|
| `gocode-architect` | Planning, decomposition, trade-offs | New plan file in `plans/` |
| `gocode-engineer` | Implementation, tests for new code | Code + tests in the repo |
| `gocode-reviewer` (×3, parallel) | Lens-specific verdicts, priority-ranked findings, patch sketches | Three independent review reports |
| `gocode-testdoctor` | Triage of failing tests, minimal patches | Code/test fixes, re-run of `make test` |

The orchestrating agent (the main Claude session driving the pipeline) owns
synthesis: merging the three reports, resolving conflicting verdicts, deciding
which findings to act on, and moving the plan to `completed/` once everyone
signs off.

Priority scale used by reviewers: **P0 / P1 / P2 / P3**.

### Rules

- **No skipping stages.** Every task starts with the architect and ends with the three-reviewer fan-out.
- **Plan file first.** The architect MUST produce a plan file before any code is written. If a plan already exists for the task, update it rather than creating a new one.
- **Three reviewers, three lenses, one message.** All three `gocode-reviewer` agents are launched in a single tool-call batch (multiple `Agent` blocks in one message) so they run in parallel. Each prompt names the lens explicitly and tells the agent what to SKIP (the other lenses) to avoid duplicated work.
- **No solo reviewer pass on first review.** Even for small changes the full three-lens fan-out is required, because the lenses catch genuinely different classes of issue (Lens A won't see ops/log-volume problems; Lens C won't see test gaps). Skipping lenses is what the orchestrator does AFTER a P0/P1 fix, not BEFORE the first verdict.
- **Lens prompts are self-contained.** Each reviewer's prompt must include: (1) the lens name, (2) what to focus on, (3) what to SKIP (so it doesn't restate other lenses), (4) the file list, (5) the deliverable shape (P0 / P1 / P2 / P3 with `file:line` + patch sketch), (6) the word cap (typically 600 words).
- **Re-review after fixes is single-pass.** Once an engineer addresses P0/P1 findings, the orchestrator runs ONE reviewer pass scoped to the changed lines, not the full fan-out. Re-running all three each iteration is expensive and rediscovers nothing.
- **Conflict resolution is explicit.** When reviewers disagree (one says P0, another says trade-off), the orchestrator chooses, names the rejected suggestion, and explains the reasoning to the user before moving on. The user has final say.
- **Orchestrator gates completion.** The plan moves to `plans/completed/` only after every reviewer's P0 and P1 findings are addressed (either fixed, or explicitly accepted with rationale). The rename uses the standard `YYMMDD.NNNN.slug.md` format.
- **`make test` must pass** before review begins. If it fails, hand the logs to `gocode-testdoctor` first — reviewers should not waste time on a red tree.
- **Testdoctor is scoped.** It patches tests or the minimal production code needed to make the failure go away. It does not redesign or refactor.
</content>
</invoke>
