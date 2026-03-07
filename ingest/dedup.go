package ingest

import (
	"crypto/sha256"
	"strings"

	"github.com/monooso/chessbook/chess/pgn"
)

// sevenTagRoster lists the PGN Seven Tag Roster fields in canonical order.
var sevenTagRoster = []string{"Event", "Site", "Date", "Round", "White", "Black", "Result"}

// dedupHash computes a SHA-256 hash of the Seven Tag Roster plus movetext.
// This is used to detect duplicate games during ingestion.
func dedupHash(g pgn.Game) [32]byte {
	h := sha256.New()
	for _, tag := range sevenTagRoster {
		h.Write([]byte(g.Tags[tag]))
		h.Write([]byte{0}) // null separator
	}
	h.Write([]byte(strings.Join(g.Moves, " ")))

	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}
