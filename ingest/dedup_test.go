package ingest

import (
	"testing"

	"github.com/monooso/chessbook/chess/pgn"
)

func TestDedupHash(t *testing.T) {
	g := pgn.Game{
		Tags: map[string]string{
			"Event":  "Test",
			"Site":   "Here",
			"Date":   "2024.01.15",
			"Round":  "1",
			"White":  "Alice",
			"Black":  "Bob",
			"Result": "1-0",
		},
		Moves:  []string{"e4", "e5", "Nf3"},
		Result: "1-0",
	}

	hash1 := dedupHash(g)
	hash2 := dedupHash(g)

	// Same game produces same hash.
	if hash1 != hash2 {
		t.Error("same game produced different hashes")
	}

	// Different movetext produces different hash.
	g2 := g
	g2.Moves = []string{"d4", "d5"}
	hash3 := dedupHash(g2)
	if hash1 == hash3 {
		t.Error("different movetext produced same hash")
	}

	// Different player produces different hash.
	g3 := g
	g3.Tags = make(map[string]string)
	for k, v := range g.Tags {
		g3.Tags[k] = v
	}
	g3.Tags["White"] = "Charlie"
	hash4 := dedupHash(g3)
	if hash1 == hash4 {
		t.Error("different player produced same hash")
	}
}
