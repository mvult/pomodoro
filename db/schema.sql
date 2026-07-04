-- Riot games tracking schema
CREATE TABLE IF NOT EXISTS riot_games (
    id BIGSERIAL PRIMARY KEY,
    match_id TEXT UNIQUE NOT NULL,
    played_at TIMESTAMPTZ NOT NULL,
    won BOOLEAN NOT NULL,
    analysis TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for fast lookup by match_id
CREATE INDEX IF NOT EXISTS idx_riot_games_match_id ON riot_games(match_id);

-- Index for time range queries
CREATE INDEX IF NOT EXISTS idx_riot_games_played_at ON riot_games(played_at);