-- positions: one row per distinct position in the graph.
CREATE TABLE IF NOT EXISTS positions (
    id              BIGSERIAL PRIMARY KEY,
    zobrist_hash    BIGINT NOT NULL,
    position_data   BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_positions_zobrist_hash ON positions (zobrist_hash);
CREATE UNIQUE INDEX IF NOT EXISTS idx_positions_data ON positions (position_data);

-- edges: one row per distinct move between two positions.
CREATE TABLE IF NOT EXISTS edges (
    id                  BIGSERIAL PRIMARY KEY,
    source_position_id  BIGINT NOT NULL REFERENCES positions (id),
    target_position_id  BIGINT NOT NULL REFERENCES positions (id),
    move                VARCHAR(5) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_edges_source ON edges (source_position_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_edges_source_move ON edges (source_position_id, move);

-- games: one row per ingested game.
CREATE TABLE IF NOT EXISTS games (
    id                      BIGSERIAL PRIMARY KEY,
    white                   TEXT NOT NULL,
    black                   TEXT NOT NULL,
    result                  SMALLINT NOT NULL,
    year                    SMALLINT,
    month                   SMALLINT,
    day                     SMALLINT,
    time_control            TEXT,
    time_control_category   SMALLINT,
    starting_fen            TEXT,
    dedup_hash              BYTEA NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_games_dedup ON games (dedup_hash);
CREATE INDEX IF NOT EXISTS idx_games_white ON games (white);
CREATE INDEX IF NOT EXISTS idx_games_black ON games (black);
CREATE INDEX IF NOT EXISTS idx_games_date ON games (year, month, day);
CREATE INDEX IF NOT EXISTS idx_games_tc_category ON games (time_control_category);

-- position_games: join table linking positions to games that reached them.
CREATE TABLE IF NOT EXISTS position_games (
    position_id BIGINT NOT NULL REFERENCES positions (id),
    game_id     BIGINT NOT NULL REFERENCES games (id),
    PRIMARY KEY (position_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_position_games_game ON position_games (game_id);

-- edge_games: join table linking edges to games that played them.
CREATE TABLE IF NOT EXISTS edge_games (
    edge_id BIGINT NOT NULL REFERENCES edges (id),
    game_id BIGINT NOT NULL REFERENCES games (id),
    PRIMARY KEY (edge_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_edge_games_game ON edge_games (game_id);
