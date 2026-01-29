-- Migration 004: Add Quiz Tags Support
-- Purpose: Add tag system for flexible quiz classification and filtering
-- Date: 2026-01-20
--
-- Summary:
-- - Creates tags table for tag definitions
-- - Creates quiz_tags junction table for many-to-many relationship
-- - Adds tag-related columns to quizzes table for quick access
-- - IMPORTANT: Keeps category_id column (hybrid approach: categories + tags)

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quiz-tag junction table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS quiz_tags (
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    tag_id VARCHAR(50) NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (quiz_id, tag_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_quiz_tags_tag_id ON quiz_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_quiz_tags_quiz_id ON quiz_tags(quiz_id);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- Add import metadata columns to quizzes table
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS import_batch_id VARCHAR(100);
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS generated_at TIMESTAMP;

-- Add GIN index for array search performance (PostgreSQL-specific)
CREATE INDEX IF NOT EXISTS idx_quizzes_tags_gin ON quizzes USING GIN(tags);

-- IMPORTANT: category_id column is preserved!
-- Hybrid approach:
--   - category_id (one per quiz) = main navigation in UI (CategoriesView)
--   - tags (many per quiz) = filtering and metadata
--   - Both coexist: category for structure, tags for flexibility

-- Comments for documentation
COMMENT ON TABLE tags IS 'Tag definitions for quiz classification. Format: {category}:{value} (e.g., language:go, difficulty:easy)';
COMMENT ON TABLE quiz_tags IS 'Many-to-many relationship between quizzes and tags';
COMMENT ON COLUMN tags.name IS 'Tag name in format {category}:{value}. Must be lowercase, pattern: ^[a-z0-9-:]+$';
COMMENT ON COLUMN quizzes.tags IS 'Denormalized array of tag names for quick access. Synced from quiz_tags table.';
COMMENT ON COLUMN quizzes.import_batch_id IS 'Batch identifier for quizzes imported together (e.g., from LLM generation)';
COMMENT ON COLUMN quizzes.generated_at IS 'Timestamp when quiz was generated (for LLM-generated quizzes)';
