# Setup Instructions for Riot Games Integration

## Database Setup

1. Create a PostgreSQL database for the application
2. Run the schema to create the `riot_games` table:

```sql
-- You can find this in db/schema.sql
-- Copy and run it in your PostgreSQL client
```

## Configuration

1. Copy `.env.example` to `.env`
2. Update the `DATABASE_URL` with your PostgreSQL connection string

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (required for database integration)

## Behavior

- Daily limits: 2 games on weekdays, 3 games on weekends
- Last game of the day is only allowed if all games in the last 7 days are accounted for
- "Accounted for" means: win OR (loss with non-empty analysis)
- If database connection fails or no database is configured, the last game of the day is blocked