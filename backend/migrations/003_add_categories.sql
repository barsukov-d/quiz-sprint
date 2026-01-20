-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Add category_id to quizzes table
ALTER TABLE quizzes
ADD COLUMN category_id UUID,
ADD CONSTRAINT fk_quizzes_category
    FOREIGN KEY (category_id)
    REFERENCES categories(id)
    ON DELETE SET NULL;

CREATE INDEX idx_quizzes_category_id ON quizzes(category_id);
