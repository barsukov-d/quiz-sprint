package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/lib/pq"
)

// TagRepository is a PostgreSQL implementation of quiz.TagRepository
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new PostgreSQL tag repository
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Save persists a tag (create or update)
func (r *TagRepository) Save(tag *quiz.Tag) error {
	ctx := context.Background()

	query := `
		INSERT INTO tags (id, name, created_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (name) DO NOTHING
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tag.ID().String(),
		tag.Name().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to save tag: %w", err)
	}

	return nil
}

// SaveAll persists multiple tags at once
func (r *TagRepository) SaveAll(tags []*quiz.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	ctx := context.Background()

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

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	return nil
}

// FindByName retrieves a tag by its name
func (r *TagRepository) FindByName(name string) (*quiz.Tag, error) {
	ctx := context.Background()

	query := `
		SELECT id, name
		FROM tags
		WHERE name = $1
	`

	var (
		id      string
		tagName string
	)

	err := r.db.QueryRowContext(ctx, query, name).Scan(&id, &tagName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find tag: %w", err)
	}

	return quiz.ReconstructTag(id, tagName), nil
}

// FindByNames retrieves multiple tags by their names
func (r *TagRepository) FindByNames(names []string) ([]*quiz.Tag, error) {
	if len(names) == 0 {
		return []*quiz.Tag{}, nil
	}

	ctx := context.Background()

	query := `
		SELECT id, name
		FROM tags
		WHERE name = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(names))
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []*quiz.Tag

	for rows.Next() {
		var (
			id      string
			tagName string
		)

		if err := rows.Scan(&id, &tagName); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		tags = append(tags, quiz.ReconstructTag(id, tagName))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// FindAll retrieves all tags
func (r *TagRepository) FindAll() ([]*quiz.Tag, error) {
	ctx := context.Background()

	query := `
		SELECT id, name
		FROM tags
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tags: %w", err)
	}
	defer rows.Close()

	var tags []*quiz.Tag

	for rows.Next() {
		var (
			id      string
			tagName string
		)

		if err := rows.Scan(&id, &tagName); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		tags = append(tags, quiz.ReconstructTag(id, tagName))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// FindByQuizID retrieves all tags assigned to a quiz
func (r *TagRepository) FindByQuizID(quizID quiz.QuizID) ([]*quiz.Tag, error) {
	ctx := context.Background()

	query := `
		SELECT t.id, t.name
		FROM tags t
		INNER JOIN quiz_tags qt ON t.id = qt.tag_id
		WHERE qt.quiz_id = $1
		ORDER BY t.name
	`

	rows, err := r.db.QueryContext(ctx, query, quizID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query tags by quiz ID: %w", err)
	}
	defer rows.Close()

	var tags []*quiz.Tag

	for rows.Next() {
		var (
			id      string
			tagName string
		)

		if err := rows.Scan(&id, &tagName); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		tags = append(tags, quiz.ReconstructTag(id, tagName))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// AssignTagsToQuiz creates relationships between quiz and tags
func (r *TagRepository) AssignTagsToQuiz(quizID quiz.QuizID, tags []*quiz.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	ctx := context.Background()

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

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to assign tags to quiz: %w", err)
	}

	return nil
}

// RemoveTagsFromQuiz removes all tag relationships for a quiz
func (r *TagRepository) RemoveTagsFromQuiz(quizID quiz.QuizID) error {
	ctx := context.Background()

	query := `DELETE FROM quiz_tags WHERE quiz_id = $1`

	_, err := r.db.ExecContext(ctx, query, quizID.String())
	if err != nil {
		return fmt.Errorf("failed to remove tags from quiz: %w", err)
	}

	return nil
}

// FindQuizzesByTag retrieves quizzes that have a specific tag
func (r *TagRepository) FindQuizzesByTag(tagName string, limit, offset int) ([]*quiz.QuizSummary, error) {
	ctx := context.Background()

	query := `
		SELECT
			q.id, q.title, q.description, q.category_id, q.time_limit, q.passing_score, q.created_at,
			COUNT(qu.id) as question_count
		FROM quizzes q
		INNER JOIN quiz_tags qt ON q.id = qt.quiz_id
		INNER JOIN tags t ON qt.tag_id = t.id
		LEFT JOIN questions qu ON q.id = qu.quiz_id
		WHERE t.name = $1
		GROUP BY q.id
		ORDER BY q.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tagName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query quizzes by tag: %w", err)
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

		// Reconstruct value objects
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
