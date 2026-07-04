# Riot Games Database Schema

This directory contains the database schema and queries for tracking Riot games.

## Schema

The `riot_games` table tracks all played games with the following fields:

- `id`: Primary key (bigserial)
- `match_id`: Unique Riot match ID (text)
- `played_at`: When the game was played (timestamptz)
- `won`: Whether the game was won (boolean)
- `analysis`: Game analysis text for losses (nullable text)
- `created_at`: Record creation time (timestamptz)
- `updated_at`: Record last update time (timestamptz)

## Queries

- `GetRiotGamesByMatchIDs`: Fetch games by a list of match IDs
- `GetRiotGamesInRange`: Fetch games within a time range

## Generated Code

Run `sqlc generate` to regenerate the Go bindings in `internal/db/`.