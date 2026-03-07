# notation package: decisions

## Match-against-legal-moves approach

Both `SANToMove` and `MoveToSAN` generate all legal moves for the position and match against them, rather than trying to directly compute the move from SAN or vice versa. This is simple and correct — it reuses the move generation logic without duplicating any chess rules.

The trade-off is performance: move generation is called multiple times per conversion (once in the main function, once in `disambiguation` or `matchCastling`, and once in `addCheckSuffix`). This is acceptable for our use case (ingestion, not real-time play).

## Check and checkmate suffixes

`MoveToSAN` applies the move to the position and checks whether the opponent is in check. If so, it generates moves for the opponent to determine whether it's checkmate (no legal moves) or just check. This is the only reliable way to determine the suffix.

`SANToMove` strips `+` and `#` markers before parsing, since they're redundant — the move is fully determined by the piece, disambiguation, target square, and promotion.

## Duplicate `findKing`

Both `notation` and `move` packages contain an unexported `findKing` function. The `move` package's version is unexported, so `notation` needs its own copy. This is a small amount of acceptable duplication.

## Castling direction detection

Castling is classified as kingside or queenside by comparing the rook's file (the `To` square in our king-captures-rook representation) to the king's file. This works for both standard and Chess960 positions.
