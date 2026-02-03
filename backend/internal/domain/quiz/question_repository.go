package quiz

// QuestionRepository defines the interface for question persistence and querying
// This is the SINGLE SOURCE of questions for all game modes
type QuestionRepository interface {
	// FindByID retrieves a single question by ID
	FindByID(id QuestionID) (*Question, error)

	// FindByIDs retrieves multiple questions by their IDs
	// Used by: Daily Challenge (load 10 questions), Duel (load 7 questions), Party (load N questions)
	FindByIDs(ids []QuestionID) ([]*Question, error)

	// FindByFilter retrieves questions matching filter criteria
	// Returns ALL matching questions (no randomization, no limit)
	FindByFilter(filter QuestionFilter) ([]*Question, error)

	// FindRandomQuestions retrieves random questions matching filter
	// Used by: Marathon (1 question at a time), Duel (select 7), Party (select N)
	FindRandomQuestions(filter QuestionFilter, limit int) ([]*Question, error)

	// FindQuestionsBySeed retrieves questions using deterministic seed
	// Used by: Daily Challenge (ensures all players get same questions for a given date)
	// seed should be derived from date (e.g., hash("2026-01-25") -> int64)
	FindQuestionsBySeed(filter QuestionFilter, limit int, seed int64) ([]*Question, error)

	// FindQuestionsByQuizSeed selects a whole quiz deterministically by seed
	// Picks one quiz with exactly questionsPerQuiz questions, returns all its questions
	// Used by: Daily Challenge (all questions share a common theme)
	FindQuestionsByQuizSeed(questionsPerQuiz int, seed int64) ([]*Question, error)

	// CountByFilter returns count of questions matching filter
	// Used for: validation, statistics
	CountByFilter(filter QuestionFilter) (int, error)

	// Save persists a question (create or update)
	Save(question *Question) error

	// SaveAll persists multiple questions at once
	SaveAll(questions []*Question) error

	// Delete removes a question by ID
	Delete(id QuestionID) error
}

// QuestionFilter represents criteria for filtering questions
type QuestionFilter struct {
	// CategoryID filters by category (nil = all categories)
	CategoryID *CategoryID

	// Difficulty filters by difficulty level ("easy", "medium", "hard")
	// nil = all difficulties
	Difficulty *string

	// ExcludeIDs excludes specific question IDs
	// Used by Marathon to avoid showing recent questions
	ExcludeIDs []QuestionID

	// MinPoints filters questions with points >= this value
	MinPoints *int

	// MaxPoints filters questions with points <= this value
	MaxPoints *int
}

// NewQuestionFilter creates a new empty filter
func NewQuestionFilter() QuestionFilter {
	return QuestionFilter{
		ExcludeIDs: make([]QuestionID, 0),
	}
}

// WithCategory adds category filter
func (f QuestionFilter) WithCategory(categoryID CategoryID) QuestionFilter {
	f.CategoryID = &categoryID
	return f
}

// WithDifficulty adds difficulty filter
func (f QuestionFilter) WithDifficulty(difficulty string) QuestionFilter {
	f.Difficulty = &difficulty
	return f
}

// WithExcludeIDs adds IDs to exclude
func (f QuestionFilter) WithExcludeIDs(ids []QuestionID) QuestionFilter {
	f.ExcludeIDs = append(f.ExcludeIDs, ids...)
	return f
}

// HasCategoryFilter checks if category filter is set
func (f QuestionFilter) HasCategoryFilter() bool {
	return f.CategoryID != nil
}

// HasDifficultyFilter checks if difficulty filter is set
func (f QuestionFilter) HasDifficultyFilter() bool {
	return f.Difficulty != nil
}

// HasExcludeFilter checks if exclude IDs filter is set
func (f QuestionFilter) HasExcludeFilter() bool {
	return len(f.ExcludeIDs) > 0
}
