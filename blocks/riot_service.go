package blocks

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mvult/pomodoro/internal/db"
)

type RiotService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewRiotService(ctx context.Context, databaseURL string) (*RiotService, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &RiotService{
		pool:    pool,
		queries: db.New(pool),
	}, nil
}

func (rs *RiotService) Close() {
	if rs.pool != nil {
		rs.pool.Close()
	}
}

// IsLastWeekAccountedFor checks if all games in the last 7 days are accounted for
// (won OR lost with non-empty analysis)
func (rs *RiotService) IsLastWeekAccountedFor(ctx context.Context, matchIDs []string) (bool, error) {
	if len(matchIDs) == 0 {
		return true, nil // No games to check
	}

	dbGames, err := rs.queries.GetRiotGamesByMatchIDs(ctx, matchIDs)
	if err != nil {
		return false, fmt.Errorf("failed to fetch riot games: %w", err)
	}

	// Create a map for quick lookup
	gameMap := make(map[string]db.RiotGame)
	for _, game := range dbGames {
		gameMap[game.MatchID] = game
	}

	// Check each match ID
	for _, matchID := range matchIDs {
		game, exists := gameMap[matchID]
		if !exists {
			// Game not in DB - not accounted for
			log.Printf("Riot: match %s not found in database", matchID)
			return false, nil
		}

		// Check if accounted for: won OR (lost with non-empty analysis)
		if game.Won {
			continue // Won games are accounted for
		}

		// Lost game - check if analysis is filled
		if !game.Analysis.Valid || game.Analysis.String == "" {
			log.Printf("Riot: match %s lost but analysis is empty", matchID)
			return false, nil
		}
	}

	log.Printf("Riot: all %d matches in last week are accounted for", len(matchIDs))
	return true, nil
}
