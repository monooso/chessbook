package db

import (
	"context"
	"os"
	"testing"
)

func databaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("CHESSBOOK_DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/chessbook_test?sslmode=disable"
	}
	return url
}

func TestMigrate(t *testing.T) {
	ctx := context.Background()
	url := databaseURL(t)

	pool, err := Connect(ctx, url)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()

	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Verify all five tables exist.
	tables := []string{"positions", "edges", "games", "position_games", "edge_games"}
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
			table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("checking table %q: %v", table, err)
		}
		if !exists {
			t.Errorf("table %q does not exist", table)
		}
	}
}

func TestMigrateSchemaDetails(t *testing.T) {
	ctx := context.Background()
	url := databaseURL(t)

	pool, err := Connect(ctx, url)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()

	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Verify unique constraints exist.
	uniqueIndexes := []struct {
		table string
		index string
	}{
		{"positions", "idx_positions_data"},
		{"edges", "idx_edges_source_move"},
		{"games", "idx_games_dedup"},
	}

	for _, ui := range uniqueIndexes {
		var isUnique bool
		err := pool.QueryRow(ctx,
			"SELECT indisunique FROM pg_index JOIN pg_class ON pg_class.oid = pg_index.indexrelid WHERE pg_class.relname = $1",
			ui.index,
		).Scan(&isUnique)
		if err != nil {
			t.Errorf("checking unique index %q on %q: %v", ui.index, ui.table, err)
			continue
		}
		if !isUnique {
			t.Errorf("index %q on %q is not unique", ui.index, ui.table)
		}
	}

	// Verify foreign keys on join tables.
	var fkCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.table_constraints
		WHERE constraint_type = 'FOREIGN KEY' AND table_name = 'position_games'
	`).Scan(&fkCount)
	if err != nil {
		t.Fatalf("checking position_games FKs: %v", err)
	}
	if fkCount != 2 {
		t.Errorf("position_games has %d foreign keys, want 2", fkCount)
	}

	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.table_constraints
		WHERE constraint_type = 'FOREIGN KEY' AND table_name = 'edge_games'
	`).Scan(&fkCount)
	if err != nil {
		t.Fatalf("checking edge_games FKs: %v", err)
	}
	if fkCount != 2 {
		t.Errorf("edge_games has %d foreign keys, want 2", fkCount)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	ctx := context.Background()
	url := databaseURL(t)

	pool, err := Connect(ctx, url)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()

	// Run migrate twice — should not error.
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}
