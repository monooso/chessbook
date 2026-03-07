package ingest

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/monooso/chessbook/chess/board"
	"github.com/monooso/chessbook/chess/move"
	"github.com/monooso/chessbook/chess/pgn"
	"github.com/monooso/chessbook/chess/zobrist"
)

// IngestResult summarises the outcome of an ingestion run.
type IngestResult struct {
	Ingested   int
	Duplicates int
	Errors     []error
}

// IngestFile ingests all validated games into the database.
// Each game is processed in its own transaction; failures are recorded
// but do not prevent other games from being ingested.
func IngestFile(ctx context.Context, pool *pgxpool.Pool, validated []ValidatedGame, raw []pgn.Game) (IngestResult, error) {
	if len(validated) != len(raw) {
		return IngestResult{}, fmt.Errorf("ingest: validated (%d) and raw (%d) game counts differ", len(validated), len(raw))
	}

	var result IngestResult

	for i, vg := range validated {
		err := ingestGame(ctx, pool, vg, raw[i])
		if err != nil {
			if isDuplicate(err) {
				result.Duplicates++
				continue
			}
			result.Errors = append(result.Errors, fmt.Errorf("game %d: %w", i+1, err))
			continue
		}
		result.Ingested++
	}

	return result, nil
}

func ingestGame(ctx context.Context, pool *pgxpool.Pool, vg ValidatedGame, raw pgn.Game) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Insert game record.
	gameID, err := insertGame(ctx, tx, vg, raw)
	if err != nil {
		return err
	}

	// Upsert positions and insert join table rows.
	positionIDs := make([]int64, len(vg.Positions))
	for j, pos := range vg.Positions {
		posID, err := upsertPosition(ctx, tx, &pos)
		if err != nil {
			return fmt.Errorf("position %d: %w", j, err)
		}
		positionIDs[j] = posID

		if err := insertPositionGame(ctx, tx, posID, gameID); err != nil {
			return fmt.Errorf("position_game %d: %w", j, err)
		}
	}

	// Upsert edges and insert join table rows.
	for j, m := range vg.Moves {
		sourceID := positionIDs[j]
		targetID := positionIDs[j+1]

		edgeID, err := upsertEdge(ctx, tx, sourceID, targetID, m)
		if err != nil {
			return fmt.Errorf("edge %d: %w", j, err)
		}

		if err := insertEdgeGame(ctx, tx, edgeID, gameID); err != nil {
			return fmt.Errorf("edge_game %d: %w", j, err)
		}
	}

	return tx.Commit(ctx)
}

func insertGame(ctx context.Context, tx pgx.Tx, vg ValidatedGame, raw pgn.Game) (int64, error) {
	hash := dedupHash(raw)
	result := parseResult(vg.Result)
	year, yearNull, month, monthNull, day, dayNull := parseDate(vg.Tags["Date"])
	tcCat, tcNull := parseTimeControlCategory(vg.Tags["TimeControl"])

	var yearVal, monthVal, dayVal, tcCatVal any
	if !yearNull {
		yearVal = year
	}
	if !monthNull {
		monthVal = month
	}
	if !dayNull {
		dayVal = day
	}
	if !tcNull {
		tcCatVal = tcCat
	}

	var startingFEN *string
	if vg.StartingFEN != "" {
		startingFEN = &vg.StartingFEN
	}

	var gameID int64
	err := tx.QueryRow(ctx, `
		INSERT INTO games (white, black, result, year, month, day, time_control, time_control_category, starting_fen, dedup_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		vg.Tags["White"],
		vg.Tags["Black"],
		result,
		yearVal,
		monthVal,
		dayVal,
		nilIfEmpty(vg.Tags["TimeControl"]),
		tcCatVal,
		startingFEN,
		hash[:],
	).Scan(&gameID)

	if err != nil {
		return 0, fmt.Errorf("insert game: %w", err)
	}
	return gameID, nil
}

func upsertPosition(ctx context.Context, tx pgx.Tx, pos *board.Position) (int64, error) {
	hash := zobrist.Hash(pos)
	data := board.Encode(pos)

	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO positions (zobrist_hash, position_data)
		VALUES ($1, $2)
		ON CONFLICT (position_data) DO UPDATE SET position_data = EXCLUDED.position_data
		RETURNING id
	`, int64(hash), data[:]).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert position: %w", err)
	}
	return id, nil
}

func upsertEdge(ctx context.Context, tx pgx.Tx, sourceID, targetID int64, m move.Move) (int64, error) {
	uci := m.String()

	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO edges (source_position_id, target_position_id, move)
		VALUES ($1, $2, $3)
		ON CONFLICT (source_position_id, move) DO UPDATE SET source_position_id = EXCLUDED.source_position_id
		RETURNING id
	`, sourceID, targetID, uci).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert edge: %w", err)
	}
	return id, nil
}

func insertPositionGame(ctx context.Context, tx pgx.Tx, positionID, gameID int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO position_games (position_id, game_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, positionID, gameID)
	return err
}

func insertEdgeGame(ctx context.Context, tx pgx.Tx, edgeID, gameID int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO edge_games (edge_id, game_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, edgeID, gameID)
	return err
}

func isDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
