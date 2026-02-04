-- ==========================================
-- Categories Table Migration
-- ==========================================

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon_url VARCHAR(255),
    color VARCHAR(7), -- HEX color code (e.g., #3B82F6)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active);
CREATE INDEX IF NOT EXISTS idx_categories_display_order ON categories(display_order);

-- Add category_id to quizzes table
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES categories(id) ON DELETE SET NULL;

-- Create index for category_id
CREATE INDEX IF NOT EXISTS idx_quizzes_category_id ON quizzes(category_id);

-- ==========================================
-- Seed Data: Categories
-- ==========================================
INSERT INTO categories (id, name, slug, description, icon_url, color, display_order) VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        'Programming',
        'programming',
        'Test your coding skills across multiple programming languages',
        '/icons/code.svg',
        '#3B82F6',
        1
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'Science',
        'science',
        'Explore physics, chemistry, biology, and mathematics',
        '/icons/flask.svg',
        '#10B981',
        2
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        'History',
        'history',
        'Journey through historical events and civilizations',
        '/icons/book.svg',
        '#F59E0B',
        3
    ),
    (
        '44444444-4444-4444-4444-444444444444',
        'Geography',
        'geography',
        'Discover countries, capitals, and landmarks around the world',
        '/icons/globe.svg',
        '#EF4444',
        4
    ),
    (
        '55555555-5555-5555-5555-555555555555',
        'General Knowledge',
        'general-knowledge',
        'Test your overall knowledge on various topics',
        '/icons/lightbulb.svg',
        '#8B5CF6',
        5
    ),
    (
        'dc000000-0000-0000-0000-000000000001',
        'Daily Challenge',
        'daily-challenge',
        'Daily quiz questions refreshed every day',
        '/icons/calendar.svg',
        '#F97316',
        6
    )
ON CONFLICT (id) DO NOTHING;

-- Update existing quiz to have a category
UPDATE quizzes
SET category_id = '11111111-1111-1111-1111-111111111111'
WHERE title = 'Go Programming Basics';
