package query

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/monooso/chessbook/chess/board"
)

// Stats holds win/loss/draw counts.
type Stats struct {
	WhiteWins int
	BlackWins int
	Draws     int
	Total     int
}

// MoveView holds stats for a single outgoing move from a position.
type MoveView struct {
	UCI   string
	Stats Stats
}

// PositionViewResult holds the complete view of a position: its overall stats
// and per-move stats.
type PositionViewResult struct {
	PositionID int64
	Stats      Stats
	Moves      []MoveView
}

// Filter specifies optional filtering criteria for game statistics.
type Filter struct {
	TimeControlCategory *int
	White               *string
	Black               *string
	YearFrom            *int
	YearTo              *int
}

// LookupResult holds the database ID for a found position.
type LookupResult struct {
	PositionID int64
}

// LookupPosition finds a position in the database by Zobrist hash and
// position data. Returns nil if the position is not found.
func LookupPosition(ctx context.Context, pool *pgxpool.Pool, hash uint64, data board.EncodedPosition) (*LookupResult, error) {
	var id int64
	err := pool.QueryRow(ctx,
		"SELECT id FROM positions WHERE zobrist_hash = $1 AND position_data = $2",
		int64(hash), data[:],
	).Scan(&id)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query: lookup position: %w", err)
	}

	return &LookupResult{PositionID: id}, nil
}

// PositionView returns the full view of a position: overall stats and
// per-move stats, optionally filtered by the given criteria.
func PositionView(ctx context.Context, pool *pgxpool.Pool, hash uint64, data board.EncodedPosition, filter *Filter) (*PositionViewResult, error) {
	lookup, err := LookupPosition(ctx, pool, hash, data)
	if err != nil {
		return nil, err
	}
	if lookup == nil {
		return nil, fmt.Errorf("query: position not found")
	}

	posID := lookup.PositionID

	// Fetch outgoing edges.
	type edge struct {
		id  int64
		uci string
	}
	rows, err := pool.Query(ctx,
		"SELECT id, move FROM edges WHERE source_position_id = $1",
		posID,
	)
	if err != nil {
		return nil, fmt.Errorf("query: fetch edges: %w", err)
	}

	var edges []edge
	var edgeIDs []int64
	for rows.Next() {
		var e edge
		if err := rows.Scan(&e.id, &e.uci); err != nil {
			rows.Close()
			return nil, fmt.Errorf("query: scan edge: %w", err)
		}
		edges = append(edges, e)
		edgeIDs = append(edgeIDs, e.id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: edges iteration: %w", err)
	}

	// Build the combined stats query.
	posStats, moveStats, err := fetchStats(ctx, pool, posID, edgeIDs, filter)
	if err != nil {
		return nil, err
	}

	result := &PositionViewResult{
		PositionID: posID,
		Stats:      posStats,
	}

	for _, e := range edges {
		ms := moveStats[e.id]
		result.Moves = append(result.Moves, MoveView{
			UCI:   e.uci,
			Stats: ms,
		})
	}

	return result, nil
}

func fetchStats(ctx context.Context, pool *pgxpool.Pool, posID int64, edgeIDs []int64, filter *Filter) (Stats, map[int64]Stats, error) {
	// Build filter clause for games.
	var conditions []string
	var args []any
	argIdx := 1

	// Position ID placeholder.
	posArg := "$" + strconv.Itoa(argIdx)
	args = append(args, posID)
	argIdx++

	if filter != nil {
		if filter.TimeControlCategory != nil {
			conditions = append(conditions, "g.time_control_category = $"+strconv.Itoa(argIdx))
			args = append(args, *filter.TimeControlCategory)
			argIdx++
		}
		if filter.White != nil {
			conditions = append(conditions, "g.white = $"+strconv.Itoa(argIdx))
			args = append(args, *filter.White)
			argIdx++
		}
		if filter.Black != nil {
			conditions = append(conditions, "g.black = $"+strconv.Itoa(argIdx))
			args = append(args, *filter.Black)
			argIdx++
		}
		if filter.YearFrom != nil {
			conditions = append(conditions, "g.year >= $"+strconv.Itoa(argIdx))
			args = append(args, *filter.YearFrom)
			argIdx++
		}
		if filter.YearTo != nil {
			conditions = append(conditions, "g.year <= $"+strconv.Itoa(argIdx))
			args = append(args, *filter.YearTo)
			argIdx++
		}
	}

	filterClause := ""
	if len(conditions) > 0 {
		filterClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Edge IDs placeholder.
	edgeArg := "$" + strconv.Itoa(argIdx)
	args = append(args, edgeIDs)

	// Single combined query per the design doc.
	sql := fmt.Sprintf(`
		SELECT 'position' AS source, 0::BIGINT AS edge_id, g.result
		FROM position_games pg
		JOIN games g ON g.id = pg.game_id
		WHERE pg.position_id = %s%s

		UNION ALL

		SELECT 'edge' AS source, eg.edge_id, g.result
		FROM edge_games eg
		JOIN games g ON g.id = eg.game_id
		WHERE eg.edge_id = ANY(%s)%s
	`, posArg, filterClause, edgeArg, filterClause)

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return Stats{}, nil, fmt.Errorf("query: fetch stats: %w", err)
	}
	defer rows.Close()

	var posStats Stats
	moveStatsMap := make(map[int64]Stats)

	for rows.Next() {
		var source string
		var edgeID int64
		var result int16

		if err := rows.Scan(&source, &edgeID, &result); err != nil {
			return Stats{}, nil, fmt.Errorf("query: scan stats: %w", err)
		}

		if source == "position" {
			addResult(&posStats, result)
		} else {
			s := moveStatsMap[edgeID]
			addResult(&s, result)
			moveStatsMap[edgeID] = s
		}
	}

	if err := rows.Err(); err != nil {
		return Stats{}, nil, fmt.Errorf("query: stats iteration: %w", err)
	}

	return posStats, moveStatsMap, nil
}

func addResult(s *Stats, result int16) {
	s.Total++
	switch result {
	case 1:
		s.WhiteWins++
	case -1:
		s.BlackWins++
	default:
		s.Draws++
	}
}
