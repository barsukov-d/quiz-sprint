package postgres

import (
	"database/sql"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// CategoryRepository is a PostgreSQL implementation of quiz.CategoryRepository
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new PostgreSQL category repository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindByID retrieves a category by its ID.
func (r *CategoryRepository) FindByID(id quiz.CategoryID) (*quiz.Category, error) {
	var (
		idStr string
		name  string
	)
	query := `SELECT id, name FROM categories WHERE id = $1`
	err := r.db.QueryRow(query, id.String()).Scan(&idStr, &name)
	if err == sql.ErrNoRows {
		return nil, quiz.ErrCategoryNotFound // I need to add this error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan category: %w", err)
	}

	catID, err := quiz.NewCategoryIDFromString(idStr)
	if err != nil {
		return nil, err
	}

	catName, err := quiz.NewCategoryName(name)
	if err != nil {
		return nil, err
	}

	return quiz.ReconstructCategory(catID, catName), nil
}

// FindAll retrieves all categories from the database.
func (r *CategoryRepository) FindAll() ([]*quiz.Category, error) {
	query := `SELECT id, name FROM categories ORDER BY name ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*quiz.Category
	for rows.Next() {
		var (
			idStr string
			name  string
		)
		if err := rows.Scan(&idStr, &name); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}

		catID, err := quiz.NewCategoryIDFromString(idStr)
		if err != nil {
			return nil, err
		}
		catName, err := quiz.NewCategoryName(name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, quiz.ReconstructCategory(catID, catName))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category rows: %w", err)
	}

	return categories, nil
}

// Delete removes a category from the database.
func (r *CategoryRepository) Delete(id quiz.CategoryID) error {
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return quiz.ErrCategoryNotFound
	}

	return nil
}

// Save inserts or updates a category in the database.
func (r *CategoryRepository) Save(category *quiz.Category) error {
	query := `
		INSERT INTO categories (id, name)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
	`
	_, err := r.db.Exec(query, category.ID().String(), category.Name().String())
	if err != nil {
		return fmt.Errorf("failed to save category: %w", err)
	}
	return nil
}
