package postgres

import (
	"database/sql"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// PersonalBestRepository is a PostgreSQL implementation of solo_marathon.PersonalBestRepository
type PersonalBestRepository struct {
	db *sql.DB
}

// NewPersonalBestRepository creates a new PostgreSQL personal best repository
func NewPersonalBestRepository(db *sql.DB) *PersonalBestRepository {
	return &PersonalBestRepository{db: db}
}

// Save persists a personal best record
func (r *PersonalBestRepository) Save(pb *solo_marathon.PersonalBest) error {
	query := `
		INSERT INTO marathon_personal_bests (
			id, player_id, category_id, best_streak, best_score, achieved_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (player_id, category_id) DO UPDATE SET
			best_streak = EXCLUDED.best_streak,
			best_score = EXCLUDED.best_score,
			achieved_at = EXCLUDED.achieved_at,
			updated_at = EXCLUDED.updated_at
	`

	// Get category ID (nullable for "all categories")
	var categoryID *string
	if !pb.Category().IsAllCategories() {
		cid := pb.Category().CategoryID().String()
		categoryID = &cid
	}

	_, err := r.db.Exec(query,
		pb.ID().String(),
		pb.PlayerID().String(),
		categoryID,
		pb.BestStreak(),
		pb.BestScore(),
		pb.AchievedAt(),
		pb.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save personal best: %w", err)
	}

	return nil
}

// FindByPlayerAndCategory retrieves personal best for a player in specific category
func (r *PersonalBestRepository) FindByPlayerAndCategory(
	playerID solo_marathon.UserID,
	category solo_marathon.MarathonCategory,
) (*solo_marathon.PersonalBest, error) {
	query := `
		SELECT id, player_id, category_id, best_streak, best_score, achieved_at, updated_at
		FROM marathon_personal_bests
		WHERE player_id = $1 AND (
			($2::uuid IS NULL AND category_id IS NULL) OR
			(category_id = $2)
		)
	`

	// Get category ID for query
	var categoryIDParam *string
	if !category.IsAllCategories() {
		cid := category.CategoryID().String()
		categoryIDParam = &cid
	}

	var (
		id         string
		dbPlayerID string
		categoryID sql.NullString
		bestStreak int
		bestScore  int
		achievedAt int64
		updatedAt  int64
	)

	err := r.db.QueryRow(query, playerID.String(), categoryIDParam).Scan(
		&id, &dbPlayerID, &categoryID, &bestStreak, &bestScore, &achievedAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, solo_marathon.ErrPersonalBestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query personal best: %w", err)
	}

	return r.reconstructPersonalBest(
		id, dbPlayerID, categoryID, bestStreak, bestScore, achievedAt, updatedAt,
	)
}

// FindTopByCategory retrieves top N players in a category
func (r *PersonalBestRepository) FindTopByCategory(
	category solo_marathon.MarathonCategory,
	limit int,
) ([]*solo_marathon.PersonalBest, error) {
	query := `
		SELECT id, player_id, category_id, best_streak, best_score, achieved_at, updated_at
		FROM marathon_personal_bests
		WHERE (
			($1::uuid IS NULL AND category_id IS NULL) OR
			(category_id = $1)
		)
		ORDER BY best_streak DESC, best_score DESC, achieved_at ASC
		LIMIT $2
	`

	// Get category ID for query
	var categoryIDParam *string
	if !category.IsAllCategories() {
		cid := category.CategoryID().String()
		categoryIDParam = &cid
	}

	rows, err := r.db.Query(query, categoryIDParam, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top personal bests: %w", err)
	}
	defer rows.Close()

	var results []*solo_marathon.PersonalBest

	for rows.Next() {
		var (
			id         string
			playerID   string
			categoryID sql.NullString
			bestStreak int
			bestScore  int
			achievedAt int64
			updatedAt  int64
		)

		err := rows.Scan(&id, &playerID, &categoryID, &bestStreak, &bestScore, &achievedAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan personal best row: %w", err)
		}

		pb, err := r.reconstructPersonalBest(
			id, playerID, categoryID, bestStreak, bestScore, achievedAt, updatedAt,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, pb)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating personal best rows: %w", err)
	}

	if len(results) == 0 {
		return nil, solo_marathon.ErrPersonalBestNotFound
	}

	return results, nil
}

// FindAllByPlayer retrieves all personal bests for a player (across all categories)
func (r *PersonalBestRepository) FindAllByPlayer(
	playerID solo_marathon.UserID,
) ([]*solo_marathon.PersonalBest, error) {
	query := `
		SELECT id, player_id, category_id, best_streak, best_score, achieved_at, updated_at
		FROM marathon_personal_bests
		WHERE player_id = $1
		ORDER BY best_streak DESC
	`

	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query player personal bests: %w", err)
	}
	defer rows.Close()

	var results []*solo_marathon.PersonalBest

	for rows.Next() {
		var (
			id         string
			dbPlayerID string
			categoryID sql.NullString
			bestStreak int
			bestScore  int
			achievedAt int64
			updatedAt  int64
		)

		err := rows.Scan(&id, &dbPlayerID, &categoryID, &bestStreak, &bestScore, &achievedAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan personal best row: %w", err)
		}

		pb, err := r.reconstructPersonalBest(
			id, dbPlayerID, categoryID, bestStreak, bestScore, achievedAt, updatedAt,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, pb)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating personal best rows: %w", err)
	}

	if len(results) == 0 {
		return nil, solo_marathon.ErrPersonalBestNotFound
	}

	return results, nil
}

// ========================================
// Helper Methods
// ========================================

// reconstructPersonalBest reconstructs a PersonalBest from database row
func (r *PersonalBestRepository) reconstructPersonalBest(
	id string,
	playerID string,
	categoryID sql.NullString,
	bestStreak int,
	bestScore int,
	achievedAt int64,
	updatedAt int64,
) (*solo_marathon.PersonalBest, error) {
	// Parse IDs
	pbID := solo_marathon.NewPersonalBestIDFromString(id)
	userID, err := shared.NewUserID(playerID)
	if err != nil {
		return nil, fmt.Errorf("invalid player_id: %w", err)
	}

	// Reconstruct category
	var category solo_marathon.MarathonCategory
	if categoryID.Valid {
		catID, err := quiz.NewCategoryIDFromString(categoryID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id: %w", err)
		}
		// TODO: Load category name from categories table
		category = solo_marathon.NewMarathonCategory(catID, "")
	} else {
		category = solo_marathon.NewMarathonCategoryAll()
	}

	// Reconstruct aggregate
	pb := solo_marathon.ReconstructPersonalBest(
		pbID,
		userID,
		category,
		bestStreak,
		bestScore,
		achievedAt,
		updatedAt,
	)

	return pb, nil
}
