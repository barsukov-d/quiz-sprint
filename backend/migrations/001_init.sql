-- ==========================================
-- Quiz Sprint Database Schema
-- ==========================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==========================================
-- Quizzes Table
-- ==========================================
CREATE TABLE IF NOT EXISTS quizzes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    time_limit INTEGER NOT NULL DEFAULT 30,
    passing_score INTEGER NOT NULL DEFAULT 70,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX idx_quizzes_created_at ON quizzes(created_at DESC);

-- ==========================================
-- Questions Table
-- ==========================================
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    text VARCHAR(500) NOT NULL,
    points INTEGER NOT NULL DEFAULT 10,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_questions_quiz_id ON questions(quiz_id);
CREATE INDEX idx_questions_position ON questions(quiz_id, position);

-- ==========================================
-- Answers Table
-- ==========================================
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text VARCHAR(200) NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_answers_question_id ON answers(question_id);

-- ==========================================
-- Quiz Sessions Table
-- ==========================================
CREATE TABLE IF NOT EXISTS quiz_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id VARCHAR(100) NOT NULL,
    current_question INTEGER NOT NULL DEFAULT 0,
    score INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    started_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    completed_at BIGINT
);

CREATE INDEX idx_sessions_quiz_id ON quiz_sessions(quiz_id);
CREATE INDEX idx_sessions_user_id ON quiz_sessions(user_id);
CREATE INDEX idx_sessions_status ON quiz_sessions(status);
CREATE INDEX idx_sessions_user_quiz ON quiz_sessions(user_id, quiz_id, status);

-- ==========================================
-- User Answers Table
-- ==========================================
CREATE TABLE IF NOT EXISTS user_answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES quiz_sessions(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID NOT NULL REFERENCES answers(id) ON DELETE CASCADE,
    is_correct BOOLEAN NOT NULL,
    points INTEGER NOT NULL DEFAULT 0,
    answered_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX idx_user_answers_session_id ON user_answers(session_id);
CREATE UNIQUE INDEX idx_user_answers_unique ON user_answers(session_id, question_id);

-- ==========================================
-- Leaderboard View (for fast queries)
-- ==========================================
CREATE OR REPLACE VIEW leaderboard AS
SELECT
    qs.quiz_id,
    qs.user_id,
    qs.score,
    qs.completed_at,
    ROW_NUMBER() OVER (PARTITION BY qs.quiz_id ORDER BY qs.score DESC, qs.completed_at ASC) as rank
FROM quiz_sessions qs
WHERE qs.status = 'completed'
ORDER BY qs.quiz_id, rank;

-- ==========================================
-- Seed Data (Sample Quiz)
-- ==========================================
DO $$
DECLARE
    quiz_id UUID;
    q1_id UUID;
    q2_id UUID;
BEGIN
    -- Check if data already exists
    IF NOT EXISTS (SELECT 1 FROM quizzes LIMIT 1) THEN
        -- Create sample quiz
        INSERT INTO quizzes (id, title, description, time_limit, passing_score)
        VALUES (uuid_generate_v4(), 'Go Programming Basics', 'Test your knowledge of Go programming fundamentals', 30, 70)
        RETURNING id INTO quiz_id;

        -- Question 1
        INSERT INTO questions (id, quiz_id, text, points, position)
        VALUES (uuid_generate_v4(), quiz_id, 'What is a goroutine?', 10, 1)
        RETURNING id INTO q1_id;

        INSERT INTO answers (question_id, text, is_correct, position) VALUES
        (q1_id, 'A lightweight thread', TRUE, 1),
        (q1_id, 'A function', FALSE, 2),
        (q1_id, 'A variable', FALSE, 3);

        -- Question 2
        INSERT INTO questions (id, quiz_id, text, points, position)
        VALUES (uuid_generate_v4(), quiz_id, 'Which keyword is used for error handling?', 10, 2)
        RETURNING id INTO q2_id;

        INSERT INTO answers (question_id, text, is_correct, position) VALUES
        (q2_id, 'try', FALSE, 1),
        (q2_id, 'catch', FALSE, 2),
        (q2_id, 'defer', TRUE, 3);

        RAISE NOTICE 'Sample data created successfully';
    ELSE
        RAISE NOTICE 'Data already exists, skipping seed';
    END IF;
END $$;
