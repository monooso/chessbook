package ingest

import (
	"strings"
	"testing"

	"github.com/monooso/chessbook/chess/pgn"
)

func parseGames(t *testing.T, input string) []pgn.Game {
	t.Helper()
	games, err := pgn.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("pgn.Parse: %v", err)
	}
	return games
}

func TestValidateFileValidGame(t *testing.T) {
	games := parseGames(t, `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 1-0
`)

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	if len(validated) != 1 {
		t.Fatalf("got %d validated games, want 1", len(validated))
	}

	vg := validated[0]

	// 6 moves played, so 7 positions (starting + after each move).
	if len(vg.Positions) != 7 {
		t.Errorf("got %d positions, want 7", len(vg.Positions))
	}

	// 6 moves.
	if len(vg.Moves) != 6 {
		t.Errorf("got %d moves, want 6", len(vg.Moves))
	}

	// Tags preserved.
	if vg.Tags["White"] != "Alice" {
		t.Errorf("White = %q, want %q", vg.Tags["White"], "Alice")
	}

	if vg.Result != "1-0" {
		t.Errorf("Result = %q, want %q", vg.Result, "1-0")
	}
}

func TestValidateFileMissingRequiredTag(t *testing.T) {
	// Missing White tag.
	games := parseGames(t, `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[Black "Bob"]
[Result "1-0"]

1. e4 1-0
`)

	_, err := ValidateFile(games)
	if err == nil {
		t.Fatal("expected error for missing required tag")
	}
}

func TestValidateFileIllegalMove(t *testing.T) {
	games := parseGames(t, `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nxe5 1-0
`)

	_, err := ValidateFile(games)
	if err == nil {
		t.Fatal("expected error for illegal move")
	}
}

func TestValidateFileMultipleGames(t *testing.T) {
	games := parseGames(t, `[Event "Test"]
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

1. d4 d5 0-1
`)

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	if len(validated) != 2 {
		t.Fatalf("got %d validated games, want 2", len(validated))
	}
}

func TestValidateFileRejectsUnknownResult(t *testing.T) {
	games := parseGames(t, `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "*"]

1. e4 *
`)

	_, err := ValidateFile(games)
	if err == nil {
		t.Fatal("expected error for * result")
	}
}

func TestValidateFileCustomStartingPosition(t *testing.T) {
	games := parseGames(t, `[Event "Test"]
[Site "Here"]
[Date "2024.01.15"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]
[FEN "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3"]
[SetUp "1"]

1... e5 1-0
`)

	validated, err := ValidateFile(games)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	if len(validated) != 1 {
		t.Fatalf("got %d games, want 1", len(validated))
	}

	// 1 move, so 2 positions.
	if len(validated[0].Positions) != 2 {
		t.Errorf("got %d positions, want 2", len(validated[0].Positions))
	}

	if validated[0].StartingFEN == "" {
		t.Error("StartingFEN should be set for non-standard starting position")
	}
}
