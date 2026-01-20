-- ==========================================
-- Seed Data: Multiple Quizzes with Categories
-- ==========================================

-- Quiz 1: JavaScript ES6+ (Programming Category)
DO $$
DECLARE
    quiz_id UUID := '01111111-1111-1111-1111-111111111111';
    q1_id UUID;
    q2_id UUID;
    q3_id UUID;
    q4_id UUID;
    q5_id UUID;
BEGIN
    -- Delete if exists
    DELETE FROM quizzes WHERE id = quiz_id;

    -- Insert quiz
    INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score)
    VALUES (
        quiz_id,
        'JavaScript ES6+ Fundamentals',
        'Master modern JavaScript features including arrow functions, promises, async/await, and destructuring',
        '11111111-1111-1111-1111-111111111111', -- Programming category
        300, -- 5 minutes
        70
    );

    -- Question 1
    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What does the spread operator (...) do in JavaScript?', 10, 1)
    RETURNING id INTO q1_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q1_id, 'Creates a new variable', FALSE, 1),
    (q1_id, 'Defines a function', FALSE, 2),
    (q1_id, 'Expands an iterable into individual elements', TRUE, 3),
    (q1_id, 'Creates a loop', FALSE, 4);

    -- Question 2
    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the purpose of async/await?', 10, 2)
    RETURNING id INTO q2_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q2_id, 'To handle asynchronous operations in a synchronous-looking way', TRUE, 1),
    (q2_id, 'To create classes', FALSE, 2),
    (q2_id, 'To define constants', FALSE, 3),
    (q2_id, 'To create loops', FALSE, 4);

    -- Question 3
    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is a closure in JavaScript?', 10, 3)
    RETURNING id INTO q3_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q3_id, 'An async function', FALSE, 1),
    (q3_id, 'A loop structure', FALSE, 2),
    (q3_id, 'A class method', FALSE, 3),
    (q3_id, 'A function with access to its outer scope even after the outer function returns', TRUE, 4);

    -- Question 4
    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the difference between let and const?', 10, 4)
    RETURNING id INTO q4_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q4_id, 'They are exactly the same', FALSE, 1),
    (q4_id, 'const cannot be reassigned, let can be reassigned', TRUE, 2),
    (q4_id, 'let is faster than const', FALSE, 3),
    (q4_id, 'const is deprecated', FALSE, 4);

    -- Question 5
    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What does array.map() return?', 10, 5)
    RETURNING id INTO q5_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q5_id, 'A filtered array', FALSE, 1),
    (q5_id, 'A sorted array', FALSE, 2),
    (q5_id, 'A new array with transformed elements', TRUE, 3),
    (q5_id, 'A reversed array', FALSE, 4);
END $$;

-- Quiz 2: Python Basics (Programming Category)
DO $$
DECLARE
    quiz_id UUID := '02222222-2222-2222-2222-222222222222';
    q1_id UUID;
    q2_id UUID;
    q3_id UUID;
    q4_id UUID;
    q5_id UUID;
BEGIN
    DELETE FROM quizzes WHERE id = quiz_id;

    INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score)
    VALUES (
        quiz_id,
        'Python Programming Essentials',
        'Test your knowledge of Python syntax, data structures, and core concepts',
        '11111111-1111-1111-1111-111111111111', -- Programming category
        240,
        65
    );

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is a list comprehension in Python?', 10, 1)
    RETURNING id INTO q1_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q1_id, 'A type of loop', FALSE, 1),
    (q1_id, 'A concise way to create lists', TRUE, 2),
    (q1_id, 'A function decorator', FALSE, 3),
    (q1_id, 'A class method', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What does the "self" keyword represent in Python classes?', 10, 2)
    RETURNING id INTO q2_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q2_id, 'The class itself', FALSE, 1),
    (q2_id, 'A global variable', FALSE, 2),
    (q2_id, 'A function parameter', FALSE, 3),
    (q2_id, 'The instance of the class', TRUE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the difference between a tuple and a list?', 10, 3)
    RETURNING id INTO q3_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q3_id, 'Tuples are immutable, lists are mutable', TRUE, 1),
    (q3_id, 'They are exactly the same', FALSE, 2),
    (q3_id, 'Lists are immutable, tuples are mutable', FALSE, 3),
    (q3_id, 'Tuples are faster than lists', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What does the "with" statement do?', 10, 4)
    RETURNING id INTO q4_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q4_id, 'Creates a new variable', FALSE, 1),
    (q4_id, 'Defines a function', FALSE, 2),
    (q4_id, 'Ensures proper resource cleanup using context managers', TRUE, 3),
    (q4_id, 'Creates a loop', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is a decorator in Python?', 10, 5)
    RETURNING id INTO q5_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q5_id, 'A type of variable', FALSE, 1),
    (q5_id, 'A function that modifies another function', TRUE, 2),
    (q5_id, 'A loop structure', FALSE, 3),
    (q5_id, 'A class attribute', FALSE, 4);
END $$;

-- Quiz 3: Physics Fundamentals (Science Category)
DO $$
DECLARE
    quiz_id UUID := '03333333-3333-3333-3333-333333333333';
    q1_id UUID;
    q2_id UUID;
    q3_id UUID;
    q4_id UUID;
    q5_id UUID;
BEGIN
    DELETE FROM quizzes WHERE id = quiz_id;

    INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score)
    VALUES (
        quiz_id,
        'Physics Fundamentals',
        'Explore classical mechanics, energy, and forces in our physical world',
        '22222222-2222-2222-2222-222222222222', -- Science category
        360,
        70
    );

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is Newton''s First Law of Motion?', 10, 1)
    RETURNING id INTO q1_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q1_id, 'Force equals mass times acceleration', FALSE, 1),
    (q1_id, 'Every action has an equal and opposite reaction', FALSE, 2),
    (q1_id, 'Energy cannot be created or destroyed', FALSE, 3),
    (q1_id, 'An object at rest stays at rest unless acted upon by a force', TRUE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the speed of light in vacuum?', 10, 2)
    RETURNING id INTO q2_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q2_id, '150,000,000 meters per second', FALSE, 1),
    (q2_id, '500,000,000 meters per second', FALSE, 2),
    (q2_id, '299,792,458 meters per second', TRUE, 3),
    (q2_id, '1,000,000 meters per second', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is kinetic energy?', 10, 3)
    RETURNING id INTO q3_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q3_id, 'Energy stored in position', FALSE, 1),
    (q3_id, 'Energy of motion', TRUE, 2),
    (q3_id, 'Heat energy', FALSE, 3),
    (q3_id, 'Chemical energy', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is gravity?', 10, 4)
    RETURNING id INTO q4_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q4_id, 'A force that attracts objects with mass toward each other', TRUE, 1),
    (q4_id, 'A type of energy', FALSE, 2),
    (q4_id, 'A chemical reaction', FALSE, 3),
    (q4_id, 'A magnetic force', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the formula for force?', 10, 5)
    RETURNING id INTO q5_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q5_id, 'E = mc²', FALSE, 1),
    (q5_id, 'P = mv', FALSE, 2),
    (q5_id, 'W = Fd', FALSE, 3),
    (q5_id, 'F = ma (Force equals mass times acceleration)', TRUE, 4);
END $$;

-- Quiz 4: World History (History Category)
DO $$
DECLARE
    quiz_id UUID := '04444444-4444-4444-4444-444444444444';
    q1_id UUID;
    q2_id UUID;
    q3_id UUID;
    q4_id UUID;
    q5_id UUID;
BEGIN
    DELETE FROM quizzes WHERE id = quiz_id;

    INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score)
    VALUES (
        quiz_id,
        'World History Milestones',
        'Journey through major historical events that shaped our modern world',
        '33333333-3333-3333-3333-333333333333', -- History category
        420,
        65
    );

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'In what year did World War II end?', 10, 1)
    RETURNING id INTO q1_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q1_id, '1945', TRUE, 1),
    (q1_id, '1944', FALSE, 2),
    (q1_id, '1946', FALSE, 3),
    (q1_id, '1943', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'Who was the first president of the United States?', 10, 2)
    RETURNING id INTO q2_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q2_id, 'Thomas Jefferson', FALSE, 1),
    (q2_id, 'George Washington', TRUE, 2),
    (q2_id, 'Abraham Lincoln', FALSE, 3),
    (q2_id, 'John Adams', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'When did the Berlin Wall fall?', 10, 3)
    RETURNING id INTO q3_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q3_id, '1991', FALSE, 1),
    (q3_id, '1987', FALSE, 2),
    (q3_id, '1990', FALSE, 3),
    (q3_id, '1989', TRUE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What ancient civilization built the pyramids?', 10, 4)
    RETURNING id INTO q4_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q4_id, 'Ancient Greeks', FALSE, 1),
    (q4_id, 'Ancient Romans', FALSE, 2),
    (q4_id, 'Ancient Egyptians', TRUE, 3),
    (q4_id, 'Maya', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'Who wrote the Communist Manifesto?', 10, 5)
    RETURNING id INTO q5_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q5_id, 'Karl Marx and Friedrich Engels', TRUE, 1),
    (q5_id, 'Vladimir Lenin', FALSE, 2),
    (q5_id, 'Joseph Stalin', FALSE, 3),
    (q5_id, 'Mao Zedong', FALSE, 4);
END $$;

-- Quiz 5: Countries & Capitals (Geography Category)
DO $$
DECLARE
    quiz_id UUID := '05555555-5555-5555-5555-555555555555';
    q1_id UUID;
    q2_id UUID;
    q3_id UUID;
    q4_id UUID;
    q5_id UUID;
BEGIN
    DELETE FROM quizzes WHERE id = quiz_id;

    INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score)
    VALUES (
        quiz_id,
        'Countries & Capitals Quiz',
        'Test your knowledge of world capitals, countries, and their locations',
        '44444444-4444-4444-4444-444444444444', -- Geography category
        180,
        60
    );

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the capital of France?', 10, 1)
    RETURNING id INTO q1_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q1_id, 'London', FALSE, 1),
    (q1_id, 'Berlin', FALSE, 2),
    (q1_id, 'Paris', TRUE, 3),
    (q1_id, 'Madrid', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the capital of Japan?', 10, 2)
    RETURNING id INTO q2_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q2_id, 'Seoul', FALSE, 1),
    (q2_id, 'Beijing', FALSE, 2),
    (q2_id, 'Bangkok', FALSE, 3),
    (q2_id, 'Tokyo', TRUE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the capital of Australia?', 10, 3)
    RETURNING id INTO q3_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q3_id, 'Canberra', TRUE, 1),
    (q3_id, 'Sydney', FALSE, 2),
    (q3_id, 'Melbourne', FALSE, 3),
    (q3_id, 'Brisbane', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'Which country has the largest population?', 10, 4)
    RETURNING id INTO q4_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q4_id, 'China', FALSE, 1),
    (q4_id, 'India', TRUE, 2),
    (q4_id, 'United States', FALSE, 3),
    (q4_id, 'Indonesia', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the smallest country in the world?', 10, 5)
    RETURNING id INTO q5_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q5_id, 'Monaco', FALSE, 1),
    (q5_id, 'Vatican City', TRUE, 2),
    (q5_id, 'San Marino', FALSE, 3),
    (q5_id, 'Liechtenstein', FALSE, 4);
END $$;

-- Quiz 6: Arts & Culture (General Knowledge Category)
DO $$
DECLARE
    quiz_id UUID := '06666666-6666-6666-6666-666666666666';
    q1_id UUID;
    q2_id UUID;
    q3_id UUID;
    q4_id UUID;
    q5_id UUID;
BEGIN
    DELETE FROM quizzes WHERE id = quiz_id;

    INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score)
    VALUES (
        quiz_id,
        'Arts & Culture Trivia',
        'Explore art, music, literature, and cultural achievements throughout history',
        '55555555-5555-5555-5555-555555555555', -- General Knowledge category
        300,
        65
    );

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'Who painted the Mona Lisa?', 10, 1)
    RETURNING id INTO q1_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q1_id, 'Michelangelo', FALSE, 1),
    (q1_id, 'Leonardo da Vinci', TRUE, 2),
    (q1_id, 'Raphael', FALSE, 3),
    (q1_id, 'Donatello', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'Who wrote "Romeo and Juliet"?', 10, 2)
    RETURNING id INTO q2_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q2_id, 'Charles Dickens', FALSE, 1),
    (q2_id, 'Jane Austen', FALSE, 2),
    (q2_id, 'William Shakespeare', TRUE, 3),
    (q2_id, 'Mark Twain', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'What is the largest museum in the world?', 10, 3)
    RETURNING id INTO q3_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q3_id, 'The British Museum', FALSE, 1),
    (q3_id, 'The Metropolitan Museum of Art', FALSE, 2),
    (q3_id, 'The Hermitage', FALSE, 3),
    (q3_id, 'The Louvre', TRUE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'Who composed the "Four Seasons"?', 10, 4)
    RETURNING id INTO q4_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q4_id, 'Antonio Vivaldi', TRUE, 1),
    (q4_id, 'Wolfgang Amadeus Mozart', FALSE, 2),
    (q4_id, 'Ludwig van Beethoven', FALSE, 3),
    (q4_id, 'Johann Sebastian Bach', FALSE, 4);

    INSERT INTO questions (id, quiz_id, text, points, position)
    VALUES (uuid_generate_v4(), quiz_id, 'In which city is the Sistine Chapel located?', 10, 5)
    RETURNING id INTO q5_id;

    INSERT INTO answers (question_id, text, is_correct, position) VALUES
    (q5_id, 'Rome', FALSE, 1),
    (q5_id, 'Florence', FALSE, 2),
    (q5_id, 'Vatican City', TRUE, 3),
    (q5_id, 'Venice', FALSE, 4);
END $$;
