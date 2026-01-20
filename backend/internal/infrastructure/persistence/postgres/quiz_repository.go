package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// QuizRepository is a PostgreSQL implementation of quiz.QuizRepository
type QuizRepository struct {
	db *sql.DB
}

// NewQuizRepository creates a new PostgreSQL quiz repository
func NewQuizRepository(db *sql.DB) *QuizRepository {
	return &QuizRepository{db: db}
}

// FindByID retrieves a quiz by ID with all questions and answers
func (r *QuizRepository) FindByID(id quiz.QuizID) (*quiz.Quiz, error) {
	// Load quiz
	quizData, err := r.loadQuiz(id)
	if err != nil {
		return nil, err
	}

	// Load tags
	tags, err := r.loadTags(id)
	if err != nil {
		return nil, err
	}

	// Load questions with answers
	questions, err := r.loadQuestions(id)
	if err != nil {
		return nil, err
	}

	// Reconstruct aggregate
	q := quiz.ReconstructQuiz(
		quizData.id,
		quizData.title,
		quizData.description,
		quizData.categoryID,
		quizData.timeLimit,
		quizData.passingScore,
		quizData.createdAt,
		quizData.updatedAt,
		tags,
		quizData.importBatchID,
		quizData.generatedAt,
	)

	// Add questions
	for _, question := range questions {
		if err := q.AddQuestion(question); err != nil {
			return nil, fmt.Errorf("failed to add question: %w", err)
		}
	}

	return q, nil
}

// FindAll retrieves all quizzes (without questions for performance)
func (r *QuizRepository) FindAll() ([]quiz.Quiz, error) {
	query := `
		SELECT id, title, description, category_id, time_limit, passing_score, created_at, updated_at, import_batch_id, generated_at
		FROM quizzes
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query quizzes: %w", err)
	}
	defer rows.Close()

	var quizzes []quiz.Quiz

	for rows.Next() {
		quizData, err := r.scanQuizRow(rows)
		if err != nil {
			return nil, err
		}

		// Reconstruct quiz without questions and tags (performance)
		q := quiz.ReconstructQuiz(
			quizData.id,
			quizData.title,
			quizData.description,
			quizData.categoryID,
			quizData.timeLimit,
			quizData.passingScore,
			quizData.createdAt,
			quizData.updatedAt,
			[]quiz.Tag{},
			quizData.importBatchID,
			quizData.generatedAt,
		)

		quizzes = append(quizzes, *q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quiz rows: %w", err)
	}

	return quizzes, nil
}

// FindAllSummaries retrieves all quizzes and their question counts (without loading full questions for performance)
func (r *QuizRepository) FindAllSummaries() ([]*quiz.QuizSummary, error) {
	query := `
		SELECT
			q.id, q.title, q.description, q.category_id, q.time_limit, q.passing_score, q.created_at,
			COUNT(qu.id) as question_count
		FROM quizzes q
		LEFT JOIN questions qu ON q.id = qu.quiz_id
		GROUP BY q.id
		ORDER BY q.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query quiz summaries: %w", err)
	}
	defer rows.Close()

	var summaries []*quiz.QuizSummary

	for rows.Next() {
		var (
			idStr         string
			title         string
			description   sql.NullString
			categoryIDStr sql.NullString
			timeLimit     int
			passingScore  int
			createdAt     int64
			questionCount int
		)

		err := rows.Scan(
			&idStr, &title, &description, &categoryIDStr, &timeLimit,
			&passingScore, &createdAt, &questionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quiz summary: %w", err)
		}

		// Reconstruct value objects from scanned data
		quizID, err := quiz.NewQuizIDFromString(idStr)
		if err != nil {
			return nil, err
		}
		quizTitle, err := quiz.NewQuizTitle(title)
		if err != nil {
			return nil, err
		}
		var categoryID quiz.CategoryID
		if categoryIDStr.Valid && categoryIDStr.String != "" {
			categoryID, err = quiz.NewCategoryIDFromString(categoryIDStr.String)
			if err != nil {
				return nil, err
			}
		}
		quizTimeLimit, err := quiz.NewTimeLimit(timeLimit)
		if err != nil {
			return nil, err
		}
		quizPassingScore, err := quiz.NewPassingScore(passingScore)
		if err != nil {
			return nil, err
		}
		desc := ""
		if description.Valid {
			desc = description.String
		}

		// Create summary object
		summary := quiz.NewQuizSummary(
			quizID,
			quizTitle,
			desc,
			categoryID,
			quizTimeLimit,
			quizPassingScore,
			createdAt,
			questionCount,
		)

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quiz summary rows: %w", err)
	}

	return summaries, nil
}

// FindSummariesByCategory retrieves all quizzes for a given category and their question counts
func (r *QuizRepository) FindSummariesByCategory(categoryID quiz.CategoryID) ([]*quiz.QuizSummary, error) {
	query := `
		SELECT
			q.id, q.title, q.description, q.category_id, q.time_limit, q.passing_score, q.created_at,
			COUNT(qu.id) as question_count
		FROM quizzes q
		LEFT JOIN questions qu ON q.id = qu.quiz_id
		WHERE q.category_id = $1
		GROUP BY q.id
		ORDER BY q.created_at DESC
	`

	rows, err := r.db.Query(query, categoryID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query quiz summaries by category: %w", err)
	}
	defer rows.Close()

	var summaries []*quiz.QuizSummary

	for rows.Next() {
		var (
			idStr         string
			title         string
			description   sql.NullString
			categoryIDStr sql.NullString
			timeLimit     int
			passingScore  int
			createdAt     int64
			questionCount int
		)

		err := rows.Scan(
			&idStr, &title, &description, &categoryIDStr, &timeLimit,
			&passingScore, &createdAt, &questionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quiz summary by category: %w", err)
		}

		// Reconstruct value objects from scanned data
		quizID, err := quiz.NewQuizIDFromString(idStr)
		if err != nil {
			return nil, err
		}
		quizTitle, err := quiz.NewQuizTitle(title)
		if err != nil {
			return nil, err
		}
		var catID quiz.CategoryID
		if categoryIDStr.Valid && categoryIDStr.String != "" {
			catID, err = quiz.NewCategoryIDFromString(categoryIDStr.String)
			if err != nil {
				return nil, err
			}
		}
		quizTimeLimit, err := quiz.NewTimeLimit(timeLimit)
		if err != nil {
			return nil, err
		}
		quizPassingScore, err := quiz.NewPassingScore(passingScore)
		if err != nil {
			return nil, err
		}
		desc := ""
		if description.Valid {
			desc = description.String
		}

		// Create summary object
		summary := quiz.NewQuizSummary(
			quizID,
			quizTitle,
			desc,
			catID,
			quizTimeLimit,
			quizPassingScore,
			createdAt,
			questionCount,
		)

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quiz summary rows by category: %w", err)
	}

	return summaries, nil
}

// Save stores a quiz with all its questions and answers in a transaction
func (r *QuizRepository) Save(q *quiz.Quiz) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Save quiz
	err = r.saveQuiz(tx, q)
	if err != nil {
		return err
	}

	// Save tags to tags table
	tags := q.Tags()
	if len(tags) > 0 {
		err = r.saveTags(tx, tags)
		if err != nil {
			return err
		}

		// Delete existing quiz-tag relationships
		_, err = tx.Exec("DELETE FROM quiz_tags WHERE quiz_id = $1", q.ID().String())
		if err != nil {
			return fmt.Errorf("failed to delete existing quiz tags: %w", err)
		}

		// Create new quiz-tag relationships
		err = r.assignTagsToQuiz(tx, q.ID(), tags)
		if err != nil {
			return err
		}
	}

	// Delete existing questions and answers (will be re-inserted)
	_, err = tx.Exec("DELETE FROM questions WHERE quiz_id = $1", q.ID().String())
	if err != nil {
		return fmt.Errorf("failed to delete existing questions: %w", err)
	}

	// Save questions and answers
	for _, question := range q.Questions() {
		err = r.saveQuestion(tx, q.ID(), question)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete removes a quiz (cascade will delete questions and answers)
func (r *QuizRepository) Delete(id quiz.QuizID) error {
	query := `DELETE FROM quizzes WHERE id = $1`

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete quiz: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return quiz.ErrQuizNotFound
	}

	return nil
}

// ========================================
// Private helper methods
// ========================================

type quizData struct {
	id            quiz.QuizID
	title         quiz.QuizTitle
	description   string
	categoryID    quiz.CategoryID
	timeLimit     quiz.TimeLimit
	passingScore  quiz.PassingScore
	createdAt     int64
	updatedAt     int64
	tags          []quiz.Tag
	importBatchID *string
	generatedAt   *int64
}

// loadQuiz loads a quiz from database (without questions)
func (r *QuizRepository) loadQuiz(id quiz.QuizID) (*quizData, error) {
	query := `
		SELECT id, title, description, category_id, time_limit, passing_score, created_at, updated_at, import_batch_id, generated_at
		FROM quizzes
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id.String())
	return r.scanQuizRow(row)
}

// scanQuizRow scans a quiz row into quizData
func (r *QuizRepository) scanQuizRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*quizData, error) {
	var (
		idStr          string
		title          string
		description    sql.NullString
		categoryIDStr  sql.NullString
		timeLimit      int
		passingScore   int
		createdAt      int64
		updatedAt      int64
		importBatchID  sql.NullString
		generatedAtSQL sql.NullTime
	)

	err := scanner.Scan(&idStr, &title, &description, &categoryIDStr, &timeLimit, &passingScore, &createdAt, &updatedAt, &importBatchID, &generatedAtSQL)
	if err == sql.ErrNoRows {
		return nil, quiz.ErrQuizNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan quiz: %w", err)
	}

	// Reconstruct value objects
	quizID, err := quiz.NewQuizIDFromString(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid quiz ID: %w", err)
	}

	quizTitle, err := quiz.NewQuizTitle(title)
	if err != nil {
		return nil, fmt.Errorf("invalid quiz title: %w", err)
	}

	// Category ID is nullable
	var categoryID quiz.CategoryID
	if categoryIDStr.Valid && categoryIDStr.String != "" {
		categoryID, err = quiz.NewCategoryIDFromString(categoryIDStr.String)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID: %w", err)
		}
	}

	quizTimeLimit, err := quiz.NewTimeLimit(timeLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid time limit: %w", err)
	}

	quizPassingScore, err := quiz.NewPassingScore(passingScore)
	if err != nil {
		return nil, fmt.Errorf("invalid passing score: %w", err)
	}

	desc := ""
	if description.Valid {
		desc = description.String
	}

	var batchID *string
	if importBatchID.Valid {
		batchID = &importBatchID.String
	}

	var generatedAt *int64
	if generatedAtSQL.Valid {
		ts := generatedAtSQL.Time.Unix()
		generatedAt = &ts
	}

	return &quizData{
		id:            quizID,
		title:         quizTitle,
		description:   desc,
		categoryID:    categoryID,
		timeLimit:     quizTimeLimit,
		passingScore:  quizPassingScore,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		tags:          []quiz.Tag{}, // Will be loaded separately
		importBatchID: batchID,
		generatedAt:   generatedAt,
	}, nil
}

// loadQuestions loads all questions with their answers for a quiz
func (r *QuizRepository) loadQuestions(quizID quiz.QuizID) ([]quiz.Question, error) {
	query := `
		SELECT id, text, points, position
		FROM questions
		WHERE quiz_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(query, quizID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []quiz.Question

	for rows.Next() {
		var (
			idStr    string
			text     string
			points   int
			position int
		)

		err := rows.Scan(&idStr, &text, &points, &position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		// Reconstruct value objects
		questionID, err := quiz.NewQuestionIDFromString(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid question ID: %w", err)
		}

		questionText, err := quiz.NewQuestionText(text)
		if err != nil {
			return nil, fmt.Errorf("invalid question text: %w", err)
		}

		questionPoints, err := quiz.NewPoints(points)
		if err != nil {
			return nil, fmt.Errorf("invalid points: %w", err)
		}

		// Create question
		question, err := quiz.NewQuestion(questionID, questionText, questionPoints, position)
		if err != nil {
			return nil, fmt.Errorf("failed to create question: %w", err)
		}

		// Load answers for this question
		answers, err := r.loadAnswers(questionID)
		if err != nil {
			return nil, err
		}

		for _, answer := range answers {
			if err := question.AddAnswer(answer); err != nil {
				return nil, fmt.Errorf("failed to add answer: %w", err)
			}
		}

		questions = append(questions, *question)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating question rows: %w", err)
	}

	return questions, nil
}

// loadAnswers loads all answers for a question
func (r *QuizRepository) loadAnswers(questionID quiz.QuestionID) ([]quiz.Answer, error) {
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

	var answers []quiz.Answer

	for rows.Next() {
		var (
			idStr     string
			text      string
			isCorrect bool
			position  int
		)

		err := rows.Scan(&idStr, &text, &isCorrect, &position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan answer: %w", err)
		}

		// Reconstruct value objects
		answerID, err := quiz.NewAnswerIDFromString(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid answer ID: %w", err)
		}

		answerText, err := quiz.NewAnswerText(text)
		if err != nil {
			return nil, fmt.Errorf("invalid answer text: %w", err)
		}

		// Create answer
		answer, err := quiz.NewAnswer(answerID, answerText, isCorrect, position)
		if err != nil {
			return nil, fmt.Errorf("failed to create answer: %w", err)
		}

		answers = append(answers, *answer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating answer rows: %w", err)
	}

	return answers, nil
}

// saveQuiz saves or updates a quiz
func (r *QuizRepository) saveQuiz(tx *sql.Tx, q *quiz.Quiz) error {
	query := `
		INSERT INTO quizzes (id, title, description, category_id, time_limit, passing_score, created_at, updated_at, tags, import_batch_id, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			category_id = EXCLUDED.category_id,
			time_limit = EXCLUDED.time_limit,
			passing_score = EXCLUDED.passing_score,
			updated_at = EXCLUDED.updated_at,
			tags = EXCLUDED.tags,
			import_batch_id = EXCLUDED.import_batch_id,
			generated_at = EXCLUDED.generated_at
	`

	var categoryIDStr interface{}
	if !q.CategoryID().IsZero() {
		categoryIDStr = q.CategoryID().String()
	}

	// Convert tags to string array for denormalized storage
	tagNames := q.TagNames()
	var tagsArray interface{}
	if len(tagNames) > 0 {
		tagsArray = "{" + strings.Join(tagNames, ",") + "}"
	}

	// Convert generated_at from Unix timestamp to SQL timestamp
	var generatedAtSQL interface{}
	if q.GeneratedAt() != nil {
		generatedAtSQL = time.Unix(*q.GeneratedAt(), 0)
	}

	_, err := tx.Exec(
		query,
		q.ID().String(),
		q.Title().String(),
		q.Description(),
		categoryIDStr,
		q.TimeLimit().Seconds(),
		q.PassingScore().Percentage(),
		q.CreatedAt(),
		q.UpdatedAt(),
		tagsArray,
		q.ImportBatchID(),
		generatedAtSQL,
	)

	if err != nil {
		return fmt.Errorf("failed to save quiz: %w", err)
	}

	return nil
}

// saveQuestion saves a question with its answers
func (r *QuizRepository) saveQuestion(tx *sql.Tx, quizID quiz.QuizID, q quiz.Question) error {
	// Save question
	query := `
		INSERT INTO questions (id, quiz_id, text, points, position)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.Exec(
		query,
		q.ID().String(),
		quizID.String(),
		q.Text().String(),
		q.Points().Value(),
		q.Position(),
	)

	if err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}

	// Save answers
	for _, answer := range q.Answers() {
		err = r.saveAnswer(tx, q.ID(), answer)
		if err != nil {
			return err
		}
	}

	return nil
}

// saveAnswer saves an answer
func (r *QuizRepository) saveAnswer(tx *sql.Tx, questionID quiz.QuestionID, a quiz.Answer) error {
	query := `
		INSERT INTO answers (id, question_id, text, is_correct, position)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.Exec(
		query,
		a.ID().String(),
		questionID.String(),
		a.Text().String(),
		a.IsCorrect(),
		a.Position(),
	)

	if err != nil {
		return fmt.Errorf("failed to save answer: %w", err)
	}

	return nil
}

// saveTags saves multiple tags to the tags table
func (r *QuizRepository) saveTags(tx *sql.Tx, tags []quiz.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	// Build batch insert
	valueStrings := make([]string, 0, len(tags))
	valueArgs := make([]interface{}, 0, len(tags)*2)

	for i, tag := range tags {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, CURRENT_TIMESTAMP)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, tag.ID().String(), tag.Name().String())
	}

	query := fmt.Sprintf(`
		INSERT INTO tags (id, name, created_at)
		VALUES %s
		ON CONFLICT (name) DO NOTHING
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	return nil
}

// assignTagsToQuiz creates relationships between quiz and tags in quiz_tags table
func (r *QuizRepository) assignTagsToQuiz(tx *sql.Tx, quizID quiz.QuizID, tags []quiz.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	// Build batch insert
	valueStrings := make([]string, 0, len(tags))
	valueArgs := make([]interface{}, 0, len(tags)*2)

	for i, tag := range tags {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, CURRENT_TIMESTAMP)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, quizID.String(), tag.ID().String())
	}

	query := fmt.Sprintf(`
		INSERT INTO quiz_tags (quiz_id, tag_id, created_at)
		VALUES %s
		ON CONFLICT (quiz_id, tag_id) DO NOTHING
	`, strings.Join(valueStrings, ","))

	_, err := tx.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to assign tags to quiz: %w", err)
	}

	return nil
}

// loadTags loads tags for a quiz from quiz_tags and tags tables
func (r *QuizRepository) loadTags(quizID quiz.QuizID) ([]quiz.Tag, error) {
	query := `
		SELECT t.id, t.name
		FROM tags t
		INNER JOIN quiz_tags qt ON t.id = qt.tag_id
		WHERE qt.quiz_id = $1
		ORDER BY t.name
	`

	rows, err := r.db.Query(query, quizID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	defer rows.Close()

	var tags []quiz.Tag

	for rows.Next() {
		var (
			id      string
			tagName string
		)

		if err := rows.Scan(&id, &tagName); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		tag := quiz.ReconstructTag(id, tagName)
		tags = append(tags, *tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}
