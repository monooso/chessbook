package pgn

import (
	"strings"
	"testing"
)

func TestParseSingleGame(t *testing.T) {
	input := `[Event "Test"]
[Site "Here"]
[Date "2024.01.01"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(games) != 1 {
		t.Fatalf("got %d games, want 1", len(games))
	}

	g := games[0]

	if g.Tags["Event"] != "Test" {
		t.Errorf("Event = %q, want %q", g.Tags["Event"], "Test")
	}
	if g.Tags["White"] != "Alice" {
		t.Errorf("White = %q, want %q", g.Tags["White"], "Alice")
	}

	wantMoves := []string{"e4", "e5", "Nf3", "Nc6", "Bb5"}
	if len(g.Moves) != len(wantMoves) {
		t.Fatalf("got %d moves, want %d: %v", len(g.Moves), len(wantMoves), g.Moves)
	}
	for i, want := range wantMoves {
		if g.Moves[i] != want {
			t.Errorf("move %d = %q, want %q", i, g.Moves[i], want)
		}
	}

	if g.Result != "1-0" {
		t.Errorf("Result = %q, want %q", g.Result, "1-0")
	}
}

func TestParseMultipleGames(t *testing.T) {
	input := `[Event "Game 1"]
[Result "1-0"]

1. e4 1-0

[Event "Game 2"]
[Result "0-1"]

1. d4 0-1
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(games) != 2 {
		t.Fatalf("got %d games, want 2", len(games))
	}

	if games[0].Tags["Event"] != "Game 1" {
		t.Errorf("game 1 Event = %q", games[0].Tags["Event"])
	}
	if games[1].Tags["Event"] != "Game 2" {
		t.Errorf("game 2 Event = %q", games[1].Tags["Event"])
	}
}

func TestParseCommentsStripped(t *testing.T) {
	input := `[Result "1-0"]

1. e4 {Best by test} e5 2. Nf3 {A good move} 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantMoves := []string{"e4", "e5", "Nf3"}
	if len(games[0].Moves) != len(wantMoves) {
		t.Fatalf("got %d moves, want %d: %v", len(games[0].Moves), len(wantMoves), games[0].Moves)
	}
	for i, want := range wantMoves {
		if games[0].Moves[i] != want {
			t.Errorf("move %d = %q, want %q", i, games[0].Moves[i], want)
		}
	}
}

func TestParseRAVsStripped(t *testing.T) {
	input := `[Result "1-0"]

1. e4 (1. d4 d5 2. c4) e5 (1... c5) 2. Nf3 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantMoves := []string{"e4", "e5", "Nf3"}
	if len(games[0].Moves) != len(wantMoves) {
		t.Fatalf("got %d moves, want %d: %v", len(games[0].Moves), len(wantMoves), games[0].Moves)
	}
}

func TestParseNAGsStripped(t *testing.T) {
	input := `[Result "1-0"]

1. e4 $1 e5 $2 2. Nf3 $10 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantMoves := []string{"e4", "e5", "Nf3"}
	if len(games[0].Moves) != len(wantMoves) {
		t.Fatalf("got %d moves, want %d: %v", len(games[0].Moves), len(wantMoves), games[0].Moves)
	}
}

func TestParseCastlingMoves(t *testing.T) {
	input := `[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bc4 Bc5 4. O-O Nf6 5. d3 O-O 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// O-O should appear twice in the moves list.
	castleCount := 0
	for _, m := range games[0].Moves {
		if m == "O-O" {
			castleCount++
		}
	}
	if castleCount != 2 {
		t.Errorf("got %d O-O moves, want 2: %v", castleCount, games[0].Moves)
	}
}

func TestParsePromotion(t *testing.T) {
	input := `[Result "1-0"]

1. e8=Q 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(games[0].Moves) != 1 || games[0].Moves[0] != "e8=Q" {
		t.Errorf("moves = %v, want [e8=Q]", games[0].Moves)
	}
}

func TestParseCheckAndCheckmate(t *testing.T) {
	input := `[Result "1-0"]

1. e4 e5 2. Qh5 Nc6 3. Bc4 Nf6 4. Qxf7# 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lastMove := games[0].Moves[len(games[0].Moves)-1]
	if lastMove != "Qxf7#" {
		t.Errorf("last move = %q, want %q", lastMove, "Qxf7#")
	}
}

func TestParseDrawResult(t *testing.T) {
	input := `[Result "1/2-1/2"]

1. e4 e5 1/2-1/2
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if games[0].Result != "1/2-1/2" {
		t.Errorf("Result = %q, want %q", games[0].Result, "1/2-1/2")
	}
}

func TestParseUnfinishedGame(t *testing.T) {
	input := `[Result "*"]

1. e4 e5 *
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if games[0].Result != "*" {
		t.Errorf("Result = %q, want %q", games[0].Result, "*")
	}
}

func TestParseBlackMoveNumbers(t *testing.T) {
	// Black move numbers after a comment or RAV.
	input := `[Result "1-0"]

1. e4 {comment} 1... e5 2. Nf3 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantMoves := []string{"e4", "e5", "Nf3"}
	if len(games[0].Moves) != len(wantMoves) {
		t.Fatalf("got %d moves, want %d: %v", len(games[0].Moves), len(wantMoves), games[0].Moves)
	}
}

func TestParseSemicolonComments(t *testing.T) {
	input := `[Result "1-0"]

1. e4 e5 ; this is a comment
2. Nf3 1-0
`

	games, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantMoves := []string{"e4", "e5", "Nf3"}
	if len(games[0].Moves) != len(wantMoves) {
		t.Fatalf("got %d moves, want %d: %v", len(games[0].Moves), len(wantMoves), games[0].Moves)
	}
}
