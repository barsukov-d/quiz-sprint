package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// QuizRepository is an in-memory implementation of quiz.QuizRepository
type QuizRepository struct {
	quizzes map[string]*quiz.Quiz
	mu      sync.RWMutex
}

// NewQuizRepository creates a new in-memory quiz repository
func NewQuizRepository() *QuizRepository {
	repo := &QuizRepository{
		quizzes: make(map[string]*quiz.Quiz),
	}
	repo.seedData()
	return repo
}

func (r *QuizRepository) seedData() {
	// Create sample quiz
	now := time.Now().Unix()

	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("Go Programming Basics")
	timeLimit, _ := quiz.NewTimeLimit(30)
	passingScore, _ := quiz.NewPassingScore(70)

	sampleQuiz, _ := quiz.NewQuiz(quizID, title, "Test your knowledge of Go", timeLimit, passingScore, now)

	// Add questions
	q1Text, _ := quiz.NewQuestionText("What is a goroutine?")
	q1Points, _ := quiz.NewPoints(10)
	q1, _ := quiz.NewQuestion(quiz.NewQuestionID(), q1Text, q1Points, 1)

	a1Text, _ := quiz.NewAnswerText("A lightweight thread")
	a1, _ := quiz.NewAnswer(quiz.NewAnswerID(), a1Text, true, 1)
	q1.AddAnswer(*a1)

	a2Text, _ := quiz.NewAnswerText("A function")
	a2, _ := quiz.NewAnswer(quiz.NewAnswerID(), a2Text, false, 2)
	q1.AddAnswer(*a2)

	a3Text, _ := quiz.NewAnswerText("A variable")
	a3, _ := quiz.NewAnswer(quiz.NewAnswerID(), a3Text, false, 3)
	q1.AddAnswer(*a3)

	sampleQuiz.AddQuestion(*q1)

	// Second question
	q2Text, _ := quiz.NewQuestionText("Which keyword is used for error handling?")
	q2Points, _ := quiz.NewPoints(10)
	q2, _ := quiz.NewQuestion(quiz.NewQuestionID(), q2Text, q2Points, 2)

	b1Text, _ := quiz.NewAnswerText("try")
	b1, _ := quiz.NewAnswer(quiz.NewAnswerID(), b1Text, false, 1)
	q2.AddAnswer(*b1)

	b2Text, _ := quiz.NewAnswerText("catch")
	b2, _ := quiz.NewAnswer(quiz.NewAnswerID(), b2Text, false, 2)
	q2.AddAnswer(*b2)

	b3Text, _ := quiz.NewAnswerText("defer")
	b3, _ := quiz.NewAnswer(quiz.NewAnswerID(), b3Text, true, 3)
	q2.AddAnswer(*b3)

	sampleQuiz.AddQuestion(*q2)

	r.quizzes[quizID.String()] = sampleQuiz
}

// FindByID retrieves a quiz by ID
func (r *QuizRepository) FindByID(id quiz.QuizID) (*quiz.Quiz, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q, exists := r.quizzes[id.String()]
	if !exists {
		return nil, quiz.ErrQuizNotFound
	}

	return q, nil
}

// FindAll retrieves all quizzes
func (r *QuizRepository) FindAll() ([]quiz.Quiz, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quizzes := make([]quiz.Quiz, 0, len(r.quizzes))
	for _, q := range r.quizzes {
		quizzes = append(quizzes, *q)
	}

	return quizzes, nil
}

// Save stores a quiz
func (r *QuizRepository) Save(q *quiz.Quiz) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.quizzes[q.ID().String()] = q
	return nil
}

// Delete removes a quiz
func (r *QuizRepository) Delete(id quiz.QuizID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.quizzes[id.String()]; !exists {
		return quiz.ErrQuizNotFound
	}

	delete(r.quizzes, id.String())
	return nil
}

// SessionRepository is an in-memory implementation of quiz.SessionRepository
type SessionRepository struct {
	sessions map[string]*quiz.QuizSession
	mu       sync.RWMutex
}

// NewSessionRepository creates a new in-memory session repository
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]*quiz.QuizSession),
	}
}

// FindByID retrieves a session by ID
func (r *SessionRepository) FindByID(id quiz.SessionID) (*quiz.QuizSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id.String()]
	if !exists {
		return nil, quiz.ErrSessionNotFound
	}

	return session, nil
}

// FindActiveByUserAndQuiz finds an active session for a user and quiz
func (r *SessionRepository) FindActiveByUserAndQuiz(userID shared.UserID, quizID quiz.QuizID) (*quiz.QuizSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.UserID().Equals(userID) &&
			session.QuizID().Equals(quizID) &&
			session.IsActive() {
			return session, nil
		}
	}

	return nil, quiz.ErrSessionNotFound
}

// Save stores a session
func (r *SessionRepository) Save(session *quiz.QuizSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.ID().String()] = session
	return nil
}

// LeaderboardRepository is an in-memory implementation of quiz.LeaderboardRepository
type LeaderboardRepository struct {
	sessionRepo *SessionRepository
}

// NewLeaderboardRepository creates a new in-memory leaderboard repository
func NewLeaderboardRepository(sessionRepo *SessionRepository) *LeaderboardRepository {
	return &LeaderboardRepository{
		sessionRepo: sessionRepo,
	}
}

// GetLeaderboard retrieves top scores for a quiz
func (r *LeaderboardRepository) GetLeaderboard(quizID quiz.QuizID, limit int) ([]quiz.LeaderboardEntry, error) {
	r.sessionRepo.mu.RLock()
	defer r.sessionRepo.mu.RUnlock()

	// Collect completed sessions for this quiz
	type sessionScore struct {
		userID      shared.UserID
		score       quiz.Points
		completedAt int64
	}

	scores := make([]sessionScore, 0)
	for _, session := range r.sessionRepo.sessions {
		if session.QuizID().Equals(quizID) && session.IsCompleted() {
			scores = append(scores, sessionScore{
				userID:      session.UserID(),
				score:       session.Score(),
				completedAt: session.CompletedAt(),
			})
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score.Value() > scores[j].score.Value()
	})

	// Limit results
	if len(scores) > limit {
		scores = scores[:limit]
	}

	// Convert to LeaderboardEntry
	entries := make([]quiz.LeaderboardEntry, 0, len(scores))
	for i, s := range scores {
		entry := quiz.NewLeaderboardEntry(
			s.userID,
			"User "+s.userID.String()[:8], // Simplified username
			s.score,
			i+1, // Rank
			quizID,
			s.completedAt,
		)
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetUserRank retrieves a user's rank in a quiz leaderboard
func (r *LeaderboardRepository) GetUserRank(quizID quiz.QuizID, userID shared.UserID) (int, error) {
	entries, err := r.GetLeaderboard(quizID, 1000) // Get all
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.UserID().Equals(userID) {
			return entry.Rank(), nil
		}
	}

	return 0, quiz.ErrSessionNotFound
}
