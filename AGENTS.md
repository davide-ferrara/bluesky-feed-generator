# AGENTS.md - Guidelines for AI Coding Agents

This file provides instructions for AI agents working in this repository.

---

## 1. Build & Run Commands

### feed-generator
```bash
cd feed-generator
make run        # Run with defaults
make build      # Build binary to ./bin/feedgen
make test       # Run tests
make clean      # Remove binaries and temp files
```

### feed-service
```bash
cd feed-service
make run        # Build and run
make build      # Build binary to ./bin/feed-service
make test       # Run tests
make templ      # Generate Templ templates
make dev        # Development with hot reload (air)
make clean      # Remove binaries and temp files
make deps       # Download and tidy dependencies
```

### Running a Single Test
```bash
go test -v -run TestName ./package/path
```

---

## 2. Project Structure

```
bsky-schwartz/
├── pkg/schwartz/     # Shared types (Post, ValueAnalysis, SchwartzValues)
├── db/               # Database operations (SQLite)
├── feed-generator/   # CLI for collecting/analyzing posts
└── feed-service/     # HTTP server for serving feeds
```

---

## 3. Code Style Guidelines

### Imports

Group imports in this order:
1. Standard library
2. External packages (bluesky-social/indigo, gin, etc.)
3. Local packages (bsky-schwartz/db, bsky-schwartz/pkg/schwartz)

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/bluesky-social/indigo/api/bsky"
    "github.com/gin-gonic/gin"

    dbpkg "bsky-schwartz/db"
    "bsky-schwartz/pkg/schwartz"
)
```

Use alias `dbpkg` for the local db package to avoid conflict with Go's `database/sql`.

### Formatting

- Use `go fmt` or your editor's format on save
- Maximum line length: 100 characters (soft guideline)
- Use tabs for indentation

### Types & Structs

- Use JSON tags on all public structs that are serialized
- Prefix interface names with `I` only if truly generic (e.g., `AlgoHandler` is fine as-is)
- Group related constants in typed blocks:

```go
type FeedAlgo string

const (
    AlgoValues     FeedAlgo = "values"
    AlgoEngagement FeedAlgo = "engagement"
)
```

### Naming Conventions

- **Variables**: camelCase (`userHandle`, `postCount`)
- **Constants**: PascalCase or camelCase for scoped constants (`MaxRetries`, `defaultLimit`)
- **Functions**: PascalCase (`CalculateScore`, `GetFeedPosts`)
- **Database columns**: snake_case in SQL, map to camelCase in Go
- **Packages**: single lowercase word (`db`, `schwartz`)

### Error Handling

- Use `fmt.Errorf` with `%w` for wrapped errors:

```go
if err := dbpkg.Init("../data.db"); err != nil {
    return fmt.Errorf("could not open database: %w", err)
}
```

- Return errors up the stack; don't log and ignore unless truly handling
- Use structured logging with `slog`:

```go
logger.Error("failed to initialize database", "err", err)
logger.Info("running server", "addr", addr)
```

### Logging

- Use `slog` for structured logging (standard library)
- Log levels: `Debug`, `Info`, `Warn`, `Error`
- Include relevant context as key-value pairs:

```go
logger.Info("processing posts", "count", len(posts), "model", model)
```

### Database

- All DB operations in the `db/` package
- Use prepared statements for queries
- Handle transactions when multiple writes are needed

### Templates (feed-service)

- Use [Templ](https://templ.golang.org/) for HTML components
- Generate with: `templ generate`
- Template files use `.templ` extension, generate to `*_templ.go`

### Configuration

- Environment variables for secrets (`.env` file)
- Config files (JSON/YAML) for non-secret settings
- Required env vars: `BSKY_HANDLE`, `BSKY_APP_PASSWORD`, `OPEN_ROUTER_KEY`

---

## 4. Important Patterns

### Local Package Replacements

The project uses Go module replacements for local packages:

```go
replace bsky-schwartz/db => ../db
replace bsky-schwartz/pkg/schwartz => ../pkg/schwartz
```

When importing local packages, use the module path:

```go
import (
    dbpkg "bsky-schwartz/db"
    "bsky-schwartz/pkg/schwartz"
)
```

### Engagement Score Formula

Posts are scored by engagement using:

```
score = log(engagement+1) / log(age_hours+2)
```

Where `engagement = likes + replies + reposts + quotes`.

### Value-Based Scoring

The 19 Schwartz values are scored 0-6, then weighted by user preferences.
The final score is the weighted sum: `score = Σ(value_i × weight_i)`.

---

## 5. Testing

- Test files: `*_test.go` in the same package
- Run all tests: `go test ./...`
- Run single test: `go test -v -run TestFunctionName ./package`
- No test framework configured; use standard `testing` package

---

## 6. Common Tasks

### Adding a new feed algorithm
1. Create handler function in `feed-service/algos.go`
2. Register in `Algos` map
3. Implement scoring logic

### Adding a new CLI command
1. Add flag definitions in `feed-generator/main.go`
2. Add case in the switch statement
3. Implement in appropriate file (collect.go, analyze.go, etc.)

### Database migrations
- Edit `db/db.go` `createTables()` function
- SQLite uses `CREATE TABLE IF NOT EXISTS` for idempotency
