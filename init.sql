-- Cosine2API Database Schema

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    auth TEXT NOT NULL,
    team_id VARCHAR(50) NOT NULL,
    linuxdo_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for faster queries on active accounts
CREATE INDEX IF NOT EXISTS idx_accounts_is_active ON accounts(is_active);

-- Example insert (replace with your actual values)
-- INSERT INTO accounts (auth, team_id) VALUES ('your_firebase_session_token', 'your_team_id');
