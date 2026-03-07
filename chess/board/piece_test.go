package board

import "testing"

func TestNewPiece(t *testing.T) {
	tests := []struct {
		name      string
		color     Color
		pieceType PieceType
	}{
		{"white pawn", White, Pawn},
		{"white knight", White, Knight},
		{"white bishop", White, Bishop},
		{"white rook", White, Rook},
		{"white queen", White, Queen},
		{"white king", White, King},
		{"black pawn", Black, Pawn},
		{"black knight", Black, Knight},
		{"black bishop", Black, Bishop},
		{"black rook", Black, Rook},
		{"black queen", Black, Queen},
		{"black king", Black, King},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPiece(tt.color, tt.pieceType)

			if got := p.Color(); got != tt.color {
				t.Errorf("Color() = %d, want %d", got, tt.color)
			}
			if got := p.Type(); got != tt.pieceType {
				t.Errorf("Type() = %d, want %d", got, tt.pieceType)
			}
		})
	}
}

func TestNoPieceIsEmpty(t *testing.T) {
	if !NoPiece.IsEmpty() {
		t.Error("NoPiece.IsEmpty() = false, want true")
	}

	p := NewPiece(White, Pawn)
	if p.IsEmpty() {
		t.Error("white pawn IsEmpty() = true, want false")
	}
}

func TestColorFlip(t *testing.T) {
	if White.Flip() != Black {
		t.Errorf("White.Flip() = %d, want %d", White.Flip(), Black)
	}
	if Black.Flip() != White {
		t.Errorf("Black.Flip() = %d, want %d", Black.Flip(), White)
	}
}

func TestPieceString(t *testing.T) {
	tests := []struct {
		piece Piece
		want  string
	}{
		{NoPiece, "."},
		{NewPiece(White, Pawn), "P"},
		{NewPiece(White, Knight), "N"},
		{NewPiece(White, Bishop), "B"},
		{NewPiece(White, Rook), "R"},
		{NewPiece(White, Queen), "Q"},
		{NewPiece(White, King), "K"},
		{NewPiece(Black, Pawn), "p"},
		{NewPiece(Black, Knight), "n"},
		{NewPiece(Black, Bishop), "b"},
		{NewPiece(Black, Rook), "r"},
		{NewPiece(Black, Queen), "q"},
		{NewPiece(Black, King), "k"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.piece.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
