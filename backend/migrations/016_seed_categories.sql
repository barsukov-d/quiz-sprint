-- ==========================================
-- Seed Categories (idempotent)
-- ==========================================
-- This migration ensures categories exist regardless of
-- whether the seed data in 003_create_categories_table.sql
-- was applied (it may have been missing in earlier deploys).

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
