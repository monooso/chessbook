# move package: decisions

## Pseudo-legal generation + legality filter

Moves are generated in two stages: first generate all pseudo-legal moves (candidate moves ignoring whether they leave the king in check), then filter by applying each move and checking if the king is attacked. This is simpler to implement and reason about than detecting pins during generation. The perft results confirm correctness.

The filtering approach copies the full position for each candidate move. This is not the fastest possible approach (make/unmake with incremental updates would be faster), but it's correct and simple. Performance is adequate for the ingestion use case — we're replaying games sequentially, not running a search engine. If profiling later shows this is a bottleneck, make/unmake can be retrofitted without changing the public API.

## Castling as king-captures-rook

Castling moves use the Chess960 convention: `From` is the king's square, `To` is the rook's square, and the `Castle` flag is set. The `Apply` function determines the actual destination squares (g/c for king, f/d for rook) based on which side of the king the rook is on. This works for both standard chess and Chess960 without special-casing.

## Attack detection by piece type

`IsSquareAttacked` checks each piece type separately (pawn, knight, bishop, rook, queen, king) rather than scanning outward from the target square. This is slightly redundant (bishop and queen share diagonal directions) but keeps each function simple and self-contained. The queen check uses the combined bishop + rook directions.

## findKing linear scan

`findKing` scans all 64 squares to locate the king. This is called once per candidate move during legality filtering. An alternative would be to track king positions explicitly in the Position type, but that adds state management complexity for a marginal performance gain. The linear scan is O(64) and typically finds the king early.

## Shared direction tables

`kingOffsets` and `queenDirs` reference the same slice. This is safe because neither is mutated after package initialization.
