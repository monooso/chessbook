package query

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/monooso/chessbook/chess/board"
	"github.com/monooso/chessbook/chess/fen"
	"github.com/monooso/chessbook/chess/pgn"
	"github.com/monooso/chessbook/chess/zobrist"
	"github.com/monooso/chessbook/db"
	"github.com/monooso/chessbook/ingest"
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

	_, err = pool.Exec(ctx, `
		TRUNCATE edge_games, position_games, edges, games, positions RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool, ctx
}

func ingestPGN(t *testing.T, pool *pgxpool.Pool, ctx context.Context, input string) {
	t.Helper()
	games, err := pgn.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("pgn.Parse: %v", err)
	}
	validated, err := ingest.ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}
	_, err = ingest.IngestFile(ctx, pool, validated, games)
	if err != nil {
		t.Fatalf("IngestFile: %v", err)
	}
}

func posFromFEN(t *testing.T, f string) board.Position {
	t.Helper()
	p, err := fen.Parse(f)
	if err != nil {
		t.Fatalf("fen.Parse: %v", err)
	}
	return p
}

func TestLookupPosition(t *testing.T) {
	pool, ctx := setupDB(t)

	ingestPGN(t, pool, ctx, `[Event "T"]
[Site "S"]
[Date "2024.01.01"]
[Round "1"]
[White "A"]
[Black "B"]
[Result "1-0"]

1. e4 e5 1-0
`)

	// Look up the starting position.
	pos := board.StartingPosition()
	hash := zobrist.Hash(&pos)
	data := board.Encode(&pos)

	result, err := LookupPosition(ctx, pool, hash, data)
	if err != nil {
		t.Fatalf("LookupPosition: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result for starting position")
	}

	if result.PositionID == 0 {
		t.Error("expected non-zero position ID")
	}
}

func TestLookupPositionNotFound(t *testing.T) {
	pool, ctx := setupDB(t)

	pos := board.StartingPosition()
	hash := zobrist.Hash(&pos)
	data := board.Encode(&pos)

	result, err := LookupPosition(ctx, pool, hash, data)
	if err != nil {
		t.Fatalf("LookupPosition: %v", err)
	}
	if result != nil {
		t.Error("expected nil for position not in database")
	}
}

func TestPositionView(t *testing.T) {
	pool, ctx := setupDB(t)

	// Three games from the starting position:
	// Game 1: 1. e4 e5 1-0 (white win)
	// Game 2: 1. e4 d5 0-1 (black win)
	// Game 3: 1. d4 d5 1/2-1/2 (draw)
	ingestPGN(t, pool, ctx, `[Event "T"]
[Site "S"]
[Date "2024.01.01"]
[Round "1"]
[White "A"]
[Black "B"]
[Result "1-0"]

1. e4 e5 1-0

[Event "T"]
[Site "S"]
[Date "2024.01.01"]
[Round "2"]
[White "A"]
[Black "B"]
[Result "0-1"]

1. e4 d5 0-1

[Event "T"]
[Site "S"]
[Date "2024.01.01"]
[Round "3"]
[White "A"]
[Black "B"]
[Result "1/2-1/2"]

1. d4 d5 1/2-1/2
`)

	pos := board.StartingPosition()
	hash := zobrist.Hash(&pos)
	data := board.Encode(&pos)

	view, err := PositionView(ctx, pool, hash, data, nil)
	if err != nil {
		t.Fatalf("PositionView: %v", err)
	}

	// Position stats: 3 games total (1W, 1L, 1D).
	if view.Stats.Total != 3 {
		t.Errorf("position total = %d, want 3", view.Stats.Total)
	}
	if view.Stats.WhiteWins != 1 {
		t.Errorf("position white wins = %d, want 1", view.Stats.WhiteWins)
	}
	if view.Stats.BlackWins != 1 {
		t.Errorf("position black wins = %d, want 1", view.Stats.BlackWins)
	}
	if view.Stats.Draws != 1 {
		t.Errorf("position draws = %d, want 1", view.Stats.Draws)
	}

	// Two outgoing moves: e4 (2 games) and d4 (1 game).
	if len(view.Moves) != 2 {
		t.Fatalf("move count = %d, want 2", len(view.Moves))
	}

	moveStats := make(map[string]Stats)
	for _, mv := range view.Moves {
		moveStats[mv.UCI] = mv.Stats
	}

	e4 := moveStats["e2e4"]
	if e4.Total != 2 {
		t.Errorf("e4 total = %d, want 2", e4.Total)
	}
	if e4.WhiteWins != 1 {
		t.Errorf("e4 white wins = %d, want 1", e4.WhiteWins)
	}

	d4 := moveStats["d2d4"]
	if d4.Total != 1 {
		t.Errorf("d4 total = %d, want 1", d4.Total)
	}
	if d4.Draws != 1 {
		t.Errorf("d4 draws = %d, want 1", d4.Draws)
	}
}

func TestPositionViewWithFilter(t *testing.T) {
	pool, ctx := setupDB(t)

	// Two games: one blitz (300+0), one rapid (600+0).
	ingestPGN(t, pool, ctx, `[Event "T"]
[Site "S"]
[Date "2024.01.01"]
[Round "1"]
[White "A"]
[Black "B"]
[Result "1-0"]
[TimeControl "300+0"]

1. e4 e5 1-0

[Event "T"]
[Site "S"]
[Date "2024.01.01"]
[Round "2"]
[White "A"]
[Black "B"]
[Result "0-1"]
[TimeControl "600+0"]

1. e4 d5 0-1
`)

	pos := board.StartingPosition()
	hash := zobrist.Hash(&pos)
	data := board.Encode(&pos)

	// Filter to blitz only.
	filter := &Filter{TimeControlCategory: new(2)}
	view, err := PositionView(ctx, pool, hash, data, filter)
	if err != nil {
		t.Fatalf("PositionView: %v", err)
	}

	if view.Stats.Total != 1 {
		t.Errorf("filtered total = %d, want 1", view.Stats.Total)
	}
	if view.Stats.WhiteWins != 1 {
		t.Errorf("filtered white wins = %d, want 1", view.Stats.WhiteWins)
	}
}
