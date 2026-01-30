-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Add category_id to quizzes table
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS category_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_quizzes_category'
    ) THEN
        ALTER TABLE quizzes ADD CONSTRAINT fk_quizzes_category
            FOREIGN KEY (category_id)
            REFERENCES categories(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_quizzes_category_id ON quizzes(category_id);
