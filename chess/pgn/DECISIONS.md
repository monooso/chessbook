# pgn package: decisions

## Parse-only, no validation

The PGN parser extracts structure (tags, SAN strings, result) but does not validate that the moves are legal or even syntactically valid SAN. Move validation is the ingestion layer's responsibility — it replays each game using the `move` and `notation` packages. This keeps the parser simple and avoids duplicating chess logic.

## Eager loading, not streaming

`Parse` reads the entire PGN stream into a slice of games. For very large files (multi-GB Lichess exports), a streaming API that yields one game at a time would be more memory-efficient. The current approach is simpler and sufficient for the initial implementation. A streaming parser can be added alongside `Parse` without breaking existing callers.

## Comments in braces support nesting

Brace comments `{...}` are parsed with a depth counter, allowing nested braces. The PGN specification doesn't formally define nesting for brace comments, but some PGN files in the wild contain them. Being lenient here is harmless — the content is discarded regardless.

## Game boundary detection

A new game is detected when a tag pair line appears after movetext has been collected. This matches the PGN specification, which requires tag pairs for every game. Games separated only by blank lines (with no tags) would be incorrectly concatenated, but this is a malformed PGN that would be rejected at the ingestion validation stage.
