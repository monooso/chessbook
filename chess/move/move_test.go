package move

import (
	"testing"

	"github.com/monooso/chessbook/chess/board"
)

func TestMoveString(t *testing.T) {
	tests := []struct {
		name string
		move Move
		want string
	}{
		{
			"normal move",
			Move{From: board.NewSquare(board.FileE, board.Rank2), To: board.NewSquare(board.FileE, board.Rank4)},
			"e2e4",
		},
		{
			"knight move",
			Move{From: board.NewSquare(board.FileG, board.Rank1), To: board.NewSquare(board.FileF, board.Rank3)},
			"g1f3",
		},
		{
			"promotion to queen",
			Move{
				From:      board.NewSquare(board.FileE, board.Rank7),
				To:        board.NewSquare(board.FileE, board.Rank8),
				Promotion: board.Queen,
			},
			"e7e8q",
		},
		{
			"promotion to knight",
			Move{
				From:      board.NewSquare(board.FileA, board.Rank7),
				To:        board.NewSquare(board.FileA, board.Rank8),
				Promotion: board.Knight,
			},
			"a7a8n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.move.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
