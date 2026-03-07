package ingest

import (
	"fmt"

	"github.com/monooso/chessbook/chess/board"
	"github.com/monooso/chessbook/chess/fen"
	"github.com/monooso/chessbook/chess/move"
	"github.com/monooso/chessbook/chess/notation"
	"github.com/monooso/chessbook/chess/pgn"
)

// SevenTagRoster lists the PGN Seven Tag Roster fields in canonical order.
// Used for both validation (all must be present) and dedup hashing.
var SevenTagRoster = []string{"Event", "Site", "Date", "Round", "White", "Black", "Result"}

// ValidatedGame is a game that has passed validation. It carries the replayed
// positions and moves so that ingestion does not need to replay them.
type ValidatedGame struct {
	Tags        map[string]string
	Positions   []board.Position
	Moves       []move.Move
	Result      string
	StartingFEN string // empty for standard starting position
}

// ValidateFile validates all games in a parsed PGN file.
// It rejects the entire file if any game is malformed.
func ValidateFile(games []pgn.Game) ([]ValidatedGame, error) {
	validated := make([]ValidatedGame, 0, len(games))

	for i, g := range games {
		vg, err := validateGame(g)
		if err != nil {
			return nil, fmt.Errorf("ingest: game %d: %w", i+1, err)
		}
		validated = append(validated, vg)
	}

	return validated, nil
}

func validateGame(g pgn.Game) (ValidatedGame, error) {
	// Check required tags.
	for _, tag := range SevenTagRoster {
		if _, ok := g.Tags[tag]; !ok {
			return ValidatedGame{}, fmt.Errorf("missing required tag %q", tag)
		}
	}

	// Reject games with unknown result.
	if g.Result == "*" {
		return ValidatedGame{}, fmt.Errorf("incomplete game (result is *)")
	}

	// Determine starting position.
	var pos board.Position
	var startingFEN string

	if fenStr, ok := g.Tags["FEN"]; ok && g.Tags["SetUp"] == "1" {
		var err error
		pos, err = fen.Parse(fenStr)
		if err != nil {
			return ValidatedGame{}, fmt.Errorf("invalid FEN %q: %w", fenStr, err)
		}
		startingFEN = fenStr
	} else {
		pos = board.StartingPosition()
	}

	// Replay moves, collecting positions and validated moves.
	positions := make([]board.Position, 0, len(g.Moves)+1)
	moves := make([]move.Move, 0, len(g.Moves))
	positions = append(positions, pos)

	for j, san := range g.Moves {
		m, err := notation.SANToMove(&pos, san)
		if err != nil {
			return ValidatedGame{}, fmt.Errorf("move %d (%q): %w", j+1, san, err)
		}
		moves = append(moves, m)
		pos = move.Apply(&pos, m)
		positions = append(positions, pos)
	}

	return ValidatedGame{
		Tags:        g.Tags,
		Positions:   positions,
		Moves:       moves,
		Result:      g.Result,
		StartingFEN: startingFEN,
	}, nil
}
