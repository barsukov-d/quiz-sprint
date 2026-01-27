package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// QuestionRepository is a PostgreSQL implementation of quiz.QuestionRepository
type QuestionRepository struct {
	db *sql.DB
}

// NewQuestionRepository creates a new PostgreSQL question repository
func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// FindByID retrieves a single question by ID
func (r *QuestionRepository) FindByID(id quiz.QuestionID) (*quiz.Question, error) {
	query := `
		SELECT q.id, q.text, q.points, q.position
		FROM questions q
		WHERE q.id = $1
	`

	var (
		questionID  string
		text        string
		points      int
		position    int
	)

	err := r.db.QueryRow(query, id.String()).Scan(
		&questionID, &text, &points, &position,
	)

	if err == sql.ErrNoRows {
		return nil, quiz.ErrQuestionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query question: %w", err)
	}

	// Load answers
	answers, err := r.loadAnswersForQuestion(id)
	if err != nil {
		return nil, err
	}

	// Reconstruct question
	return r.reconstructQuestion(questionID, text, points, position, answers)
}

// FindByIDs retrieves multiple questions by their IDs
func (r *QuestionRepository) FindByIDs(ids []quiz.QuestionID) ([]*quiz.Question, error) {
	if len(ids) == 0 {
		return []*quiz.Question{}, nil
	}

	// Convert IDs to strings
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	// Build placeholders ($1, $2, ...)
	placeholders := make([]string, len(stringIDs))
	args := make([]interface{}, len(stringIDs))
	for i, id := range stringIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT q.id, q.text, q.points, q.position
		FROM questions q
		WHERE q.id IN (%s)
		ORDER BY q.position ASC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions by IDs: %w", err)
	}
	defer rows.Close()

	var questions []*quiz.Question

	for rows.Next() {
		var (
			questionID  string
			text        string
			points      int
			position    int
		)

		err := rows.Scan(&questionID, &text, &points, &position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}

		// Load answers
		qid, _ := quiz.NewQuestionIDFromString(questionID)
		answers, err := r.loadAnswersForQuestion(qid)
		if err != nil {
			return nil, err
		}

		// Reconstruct question
		question, err := r.reconstructQuestion(questionID, text, points, position, answers)
		if err != nil {
			return nil, err
		}

		questions = append(questions, question)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating question rows: %w", err)
	}

	return questions, nil
}

// FindByFilter retrieves questions matching filter criteria (NO randomization)
func (r *QuestionRepository) FindByFilter(filter quiz.QuestionFilter) ([]*quiz.Question, error) {
	query, args := r.buildFilterQuery(filter, false)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions by filter: %w", err)
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

// FindRandomQuestions retrieves random questions matching filter
func (r *QuestionRepository) FindRandomQuestions(filter quiz.QuestionFilter, limit int) ([]*quiz.Question, error) {
	query, args := r.buildFilterQuery(filter, true)

	// Add limit
	args = append(args, limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query random questions: %w", err)
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

// FindQuestionsBySeed retrieves questions using deterministic seed
// This ensures all players get the same questions for the same seed (e.g., daily challenge date)
func (r *QuestionRepository) FindQuestionsBySeed(filter quiz.QuestionFilter, limit int, seed int64) ([]*quiz.Question, error) {
	// PostgreSQL's setseed() requires a value between -1.0 and 1.0
	// We normalize the seed: seed % 1000000 / 1000000.0 gives us 0.0 to 0.999999
	normalizedSeed := float64(seed%1000000) / 1000000.0

	// Set the random seed for this session
	_, err := r.db.Exec("SELECT setseed($1)", normalizedSeed)
	if err != nil {
		return nil, fmt.Errorf("failed to set seed: %w", err)
	}

	// Build query with ORDER BY random() (but now it's deterministic due to setseed)
	query, args := r.buildFilterQuery(filter, true)

	// Add limit
	args = append(args, limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions by seed: %w", err)
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

// CountByFilter returns count of questions matching filter
func (r *QuestionRepository) CountByFilter(filter quiz.QuestionFilter) (int, error) {
	baseQuery, args := r.buildFilterQueryBase(filter)

	query := "SELECT COUNT(*) FROM (" + baseQuery + ") AS filtered"

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count questions: %w", err)
	}

	return count, nil
}

// Save persists a question (create or update)
func (r *QuestionRepository) Save(question *quiz.Question) error {
	// TODO: Implement Save
	return fmt.Errorf("not implemented")
}

// SaveAll persists multiple questions at once
func (r *QuestionRepository) SaveAll(questions []*quiz.Question) error {
	// TODO: Implement SaveAll (for import functionality)
	return fmt.Errorf("not implemented")
}

// Delete removes a question by ID
func (r *QuestionRepository) Delete(id quiz.QuestionID) error {
	// TODO: Implement Delete
	return fmt.Errorf("not implemented")
}

// ========================================
// Helper Methods
// ========================================

// buildFilterQuery builds SELECT query with filters
func (r *QuestionRepository) buildFilterQuery(filter quiz.QuestionFilter, random bool) (string, []interface{}) {
	baseQuery, args := r.buildFilterQueryBase(filter)

	// Add ordering
	if random {
		baseQuery += " ORDER BY RANDOM()"
	} else {
		baseQuery += " ORDER BY q.position ASC"
	}

	return baseQuery, args
}

// buildFilterQueryBase builds base query with WHERE clauses
func (r *QuestionRepository) buildFilterQueryBase(filter quiz.QuestionFilter) (string, []interface{}) {
	query := `
		SELECT q.id, q.text, q.points, q.position
		FROM questions q
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	// Filter by category - NOTE: questions table doesn't have category_id
	// Category is on the quiz level, not question level
	// For now, ignore category filter
	// TODO: Join with quizzes table if category filtering is needed

	// Filter by difficulty - NOTE: questions table doesn't have difficulty column
	// Difficulty would need to be on quiz level or added as migration
	// For now, ignore difficulty filter

	// Exclude specific IDs
	if filter.HasExcludeFilter() {
		stringIDs := make([]string, len(filter.ExcludeIDs))
		placeholders := make([]string, len(filter.ExcludeIDs))

		for i, id := range filter.ExcludeIDs {
			argCount++
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			stringIDs[i] = id.String()
			args = append(args, stringIDs[i])
		}

		query += fmt.Sprintf(" AND q.id NOT IN (%s)", strings.Join(placeholders, ","))
	}

	return query, args
}

// scanQuestions scans rows into Question objects
func (r *QuestionRepository) scanQuestions(rows *sql.Rows) ([]*quiz.Question, error) {
	var questions []*quiz.Question

	for rows.Next() {
		var (
			questionID  string
			text        string
			points      int
			position    int
		)

		err := rows.Scan(&questionID, &text, &points, &position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}

		// Load answers
		qid, _ := quiz.NewQuestionIDFromString(questionID)
		answers, err := r.loadAnswersForQuestion(qid)
		if err != nil {
			return nil, err
		}

		// Reconstruct question
		question, err := r.reconstructQuestion(questionID, text, points, position, answers)
		if err != nil {
			return nil, err
		}

		questions = append(questions, question)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating question rows: %w", err)
	}

	return questions, nil
}

// loadAnswersForQuestion loads all answers for a question
func (r *QuestionRepository) loadAnswersForQuestion(questionID quiz.QuestionID) ([]answerRow, error) {
	query := `
		SELECT id, text, is_correct, position
		FROM answers
		WHERE question_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(query, questionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query answers: %w", err)
	}
	defer rows.Close()

	var answers []answerRow

	for rows.Next() {
		var answer answerRow
		err := rows.Scan(&answer.ID, &answer.Text, &answer.IsCorrect, &answer.Position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan answer row: %w", err)
		}
		answers = append(answers, answer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating answer rows: %w", err)
	}

	return answers, nil
}

// answerRow represents an answer row from database
type answerRow struct {
	ID        string
	Text      string
	IsCorrect bool
	Position  int
}

// reconstructQuestion reconstructs a Question aggregate from database data
func (r *QuestionRepository) reconstructQuestion(
	id string,
	text string,
	points int,
	position int,
	answers []answerRow,
) (*quiz.Question, error) {
	// Parse question ID
	questionID, err := quiz.NewQuestionIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid question_id: %w", err)
	}

	// Create question text
	questionText, err := quiz.NewQuestionText(text)
	if err != nil {
		return nil, fmt.Errorf("invalid question text: %w", err)
	}

	// Create points
	questionPoints, err := quiz.NewPoints(points)
	if err != nil {
		return nil, fmt.Errorf("invalid points: %w", err)
	}

	// Create question entity
	question, err := quiz.NewQuestion(questionID, questionText, questionPoints, position)
	if err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	// Add answers to question
	for _, ans := range answers {
		answerID, err := quiz.NewAnswerIDFromString(ans.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid answer_id: %w", err)
		}

		answerText, err := quiz.NewAnswerText(ans.Text)
		if err != nil {
			return nil, fmt.Errorf("invalid answer text: %w", err)
		}

		answer, err := quiz.NewAnswer(answerID, answerText, ans.IsCorrect, ans.Position)
		if err != nil {
			return nil, fmt.Errorf("failed to create answer: %w", err)
		}

		if err := question.AddAnswer(*answer); err != nil {
			return nil, fmt.Errorf("failed to add answer: %w", err)
		}
	}

	return question, nil
}
