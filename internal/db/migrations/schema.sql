CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    daily_new_limit INTEGER NOT NULL DEFAULT 20,
    retention_goal FLOAT NOT NULL DEFAULT 0.90,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS words (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    korean VARCHAR(255) NOT NULL,
    english VARCHAR(255) NOT NULL,
    part_of_speech VARCHAR(50),
    topik_level INTEGER CHECK (topik_level >= 1 AND topik_level <= 6),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(korean, english)
);

CREATE TABLE IF NOT EXISTS user_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word_id UUID NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    state INTEGER NOT NULL DEFAULT 0, -- 0: New, 1: Learning, 2: Review, 3: Relearning
    due TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    stability FLOAT NOT NULL DEFAULT 0,
    difficulty FLOAT NOT NULL DEFAULT 0,
    elapsed_days INTEGER NOT NULL DEFAULT 0,
    scheduled_days INTEGER NOT NULL DEFAULT 0,
    reps INTEGER NOT NULL DEFAULT 0,
    lapses INTEGER NOT NULL DEFAULT 0,
    last_review TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, word_id)
);

CREATE INDEX IF NOT EXISTS idx_user_reviews_due ON user_reviews(user_id, due);
