package ingest

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/monooso/chessbook/chess/pgn"
	"github.com/monooso/chessbook/db"
)

func setupDB(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	ctx := context.Background()
	url := os.Getenv("CHESSBOOK_DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/chessbook_test?sslmode=disable"
	}

	pool, err := db.Connect(ctx, url)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Clean tables for test isolation.
	_, err = pool.Exec(ctx, `
		TRUNCATE edge_games, position_games, edges, games, positions RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool, ctx
}

func TestIngestFile(t *testing.T) {
	pool, ctx := setupDB(t)

	input := `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 1-0
`

	games, err := pgn.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("pgn.Parse: %v", err)
	}

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	result, err := IngestFile(ctx, pool, validated, games)
	if err != nil {
		t.Fatalf("IngestFile: %v", err)
	}

	if result.Ingested != 1 {
		t.Errorf("Ingested = %d, want 1", result.Ingested)
	}
	if result.Duplicates != 0 {
		t.Errorf("Duplicates = %d, want 0", result.Duplicates)
	}

	// Verify game was inserted.
	var gameCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM games").Scan(&gameCount)
	if gameCount != 1 {
		t.Errorf("games count = %d, want 1", gameCount)
	}

	// Verify positions: starting + after e4 + after e5 + after Nf3 = 4.
	var posCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM positions").Scan(&posCount)
	if posCount != 4 {
		t.Errorf("positions count = %d, want 4", posCount)
	}

	// Verify edges: e4, e5, Nf3 = 3.
	var edgeCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM edges").Scan(&edgeCount)
	if edgeCount != 3 {
		t.Errorf("edges count = %d, want 3", edgeCount)
	}

	// Verify join tables.
	var pgCount, egCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM position_games").Scan(&pgCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM edge_games").Scan(&egCount)
	if pgCount != 4 {
		t.Errorf("position_games count = %d, want 4", pgCount)
	}
	if egCount != 3 {
		t.Errorf("edge_games count = %d, want 3", egCount)
	}
}

func TestIngestFileDuplicate(t *testing.T) {
	pool, ctx := setupDB(t)

	input := `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 1-0
`

	games, err := pgn.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("pgn.Parse: %v", err)
	}

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	// Ingest once.
	_, err = IngestFile(ctx, pool, validated, games)
	if err != nil {
		t.Fatalf("first IngestFile: %v", err)
	}

	// Ingest again — should detect duplicate.
	result, err := IngestFile(ctx, pool, validated, games)
	if err != nil {
		t.Fatalf("second IngestFile: %v", err)
	}

	if result.Ingested != 0 {
		t.Errorf("Ingested = %d, want 0", result.Ingested)
	}
	if result.Duplicates != 1 {
		t.Errorf("Duplicates = %d, want 1", result.Duplicates)
	}
}

func TestIngestFileSharedPositions(t *testing.T) {
	pool, ctx := setupDB(t)

	// Two games that share the starting position and 1.e4.
	input := `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 1-0

[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "2"]
[White "Alice"]
[Black "Bob"]
[Result "0-1"]

1. e4 d5 0-1
`

	games, err := pgn.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("pgn.Parse: %v", err)
	}

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	result, err := IngestFile(ctx, pool, validated, games)
	if err != nil {
		t.Fatalf("IngestFile: %v", err)
	}

	if result.Ingested != 2 {
		t.Errorf("Ingested = %d, want 2", result.Ingested)
	}

	// Positions: starting (shared), after e4 (shared), after e5, after d5 = 4.
	var posCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM positions").Scan(&posCount)
	if posCount != 4 {
		t.Errorf("positions count = %d, want 4", posCount)
	}

	// The starting position and after-e4 position should each link to 2 games.
	var sharedCount int
	pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT position_id FROM position_games GROUP BY position_id HAVING COUNT(*) = 2
		) sub
	`).Scan(&sharedCount)
	if sharedCount != 2 {
		t.Errorf("positions shared by 2 games = %d, want 2", sharedCount)
	}
}

func TestIngestFileGameMetadata(t *testing.T) {
	pool, ctx := setupDB(t)

	input := `[Event "Test"]
[Site "Here"]
[Date "2024.01.??"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "0-1"]
[TimeControl "300+0"]

1. e4 0-1
`

	games, err := pgn.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("pgn.Parse: %v", err)
	}

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	_, err = IngestFile(ctx, pool, validated, games)
	if err != nil {
		t.Fatalf("IngestFile: %v", err)
	}

	var white, black string
	var result, year, tcCat int16
	var month, day *int16
	pool.QueryRow(ctx,
		"SELECT white, black, result, year, month, day, time_control_category FROM games LIMIT 1",
	).Scan(&white, &black, &result, &year, &month, &day, &tcCat)

	if white != "Alice" {
		t.Errorf("white = %q, want %q", white, "Alice")
	}
	if black != "Bob" {
		t.Errorf("black = %q, want %q", black, "Bob")
	}
	if result != -1 {
		t.Errorf("result = %d, want -1", result)
	}
	if year != 2024 {
		t.Errorf("year = %d, want 2024", year)
	}
	if month == nil || *month != 1 {
		t.Errorf("month = %v, want 1", month)
	}
	if day != nil {
		t.Errorf("day = %v, want nil", day)
	}
	if tcCat != 2 {
		t.Errorf("time_control_category = %d, want 2 (blitz)", tcCat)
	}
}
