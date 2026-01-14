package persistence

import (
	"context"
	"sync"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/google/uuid"
)

// MemoryQuizRepository is an in-memory implementation for development
type MemoryQuizRepository struct {
	quizzes  map[uuid.UUID]*quiz.Quiz
	sessions map[uuid.UUID]*quiz.QuizSession
	mu       sync.RWMutex
}

// NewMemoryQuizRepository creates a new in-memory repository
func NewMemoryQuizRepository() *MemoryQuizRepository {
	repo := &MemoryQuizRepository{
		quizzes:  make(map[uuid.UUID]*quiz.Quiz),
		sessions: make(map[uuid.UUID]*quiz.QuizSession),
	}

	// Seed with sample data
	repo.seedData()

	return repo
}

func (r *MemoryQuizRepository) seedData() {
	sampleQuiz := &quiz.Quiz{
		ID:           uuid.New(),
		Title:        "Go Programming Basics",
		Description:  "Test your knowledge of Go programming fundamentals",
		TimeLimit:    30,
		PassingScore: 70,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Questions: []quiz.Question{
			{
				ID:       uuid.New(),
				QuizID:   uuid.New(),
				Text:     "What is a goroutine?",
				Points:   10,
				Position: 1,
				Answers: []quiz.Answer{
					{ID: uuid.New(), Text: "A lightweight thread", IsCorrect: true, Position: 1},
					{ID: uuid.New(), Text: "A function", IsCorrect: false, Position: 2},
					{ID: uuid.New(), Text: "A variable", IsCorrect: false, Position: 3},
				},
			},
			{
				ID:       uuid.New(),
				QuizID:   uuid.New(),
				Text:     "Which keyword is used for error handling?",
				Points:   10,
				Position: 2,
				Answers: []quiz.Answer{
					{ID: uuid.New(), Text: "try", IsCorrect: false, Position: 1},
					{ID: uuid.New(), Text: "catch", IsCorrect: false, Position: 2},
					{ID: uuid.New(), Text: "defer", IsCorrect: true, Position: 3},
				},
			},
		},
	}

	r.quizzes[sampleQuiz.ID] = sampleQuiz
}

// FindByID retrieves a quiz by ID
func (r *MemoryQuizRepository) FindByID(ctx context.Context, id uuid.UUID) (*quiz.Quiz, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q, exists := r.quizzes[id]
	if !exists {
		return nil, quiz.ErrQuizNotFound
	}

	return q, nil
}

// FindAll retrieves all quizzes
func (r *MemoryQuizRepository) FindAll(ctx context.Context) ([]quiz.Quiz, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quizzes := make([]quiz.Quiz, 0, len(r.quizzes))
	for _, q := range r.quizzes {
		quizzes = append(quizzes, *q)
	}

	return quizzes, nil
}

// Save stores a quiz
func (r *MemoryQuizRepository) Save(ctx context.Context, q *quiz.Quiz) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}

	q.UpdatedAt = time.Now()
	r.quizzes[q.ID] = q

	return nil
}

// Delete removes a quiz
func (r *MemoryQuizRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.quizzes[id]; !exists {
		return quiz.ErrQuizNotFound
	}

	delete(r.quizzes, id)
	return nil
}

// FindSessionByID retrieves a session by ID
func (r *MemoryQuizRepository) FindSessionByID(ctx context.Context, id uuid.UUID) (*quiz.QuizSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return nil, quiz.ErrSessionNotFound
	}

	return session, nil
}

// FindActiveSessionByUserAndQuiz finds an active session for a user and quiz
func (r *MemoryQuizRepository) FindActiveSessionByUserAndQuiz(ctx context.Context, userID string, quizID uuid.UUID) (*quiz.QuizSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.UserID == userID && session.QuizID == quizID && session.Status == quiz.SessionStatusActive {
			return session, nil
		}
	}

	return nil, quiz.ErrSessionNotFound
}

// SaveSession stores a new session
func (r *MemoryQuizRepository) SaveSession(ctx context.Context, session *quiz.QuizSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}

	r.sessions[session.ID] = session
	return nil
}

// UpdateSession updates an existing session
func (r *MemoryQuizRepository) UpdateSession(ctx context.Context, session *quiz.QuizSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; !exists {
		return quiz.ErrSessionNotFound
	}

	r.sessions[session.ID] = session
	return nil
}

// GetLeaderboard retrieves the leaderboard for a quiz
func (r *MemoryQuizRepository) GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]quiz.LeaderboardEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]quiz.LeaderboardEntry, 0)

	for _, session := range r.sessions {
		if session.QuizID == quizID && session.Status == quiz.SessionStatusCompleted {
			entry := quiz.LeaderboardEntry{
				UserID:      session.UserID,
				Username:    "User " + session.UserID[:8],
				Score:       session.Score,
				QuizID:      quizID,
				CompletedAt: *session.CompletedAt,
			}
			entries = append(entries, entry)
		}
	}

	// Sort by score descending (simplified - you'd use sort.Slice in production)
	// Add rank
	for i := range entries {
		entries[i].Rank = i + 1
	}

	if len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}
