-- Add sport_type to tournaments. Existing rows are chess (the original platform focus).
ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS sport_type VARCHAR(50) NOT NULL DEFAULT 'chess';
CREATE INDEX IF NOT EXISTS idx_tournaments_sport_type ON tournaments(sport_type);
