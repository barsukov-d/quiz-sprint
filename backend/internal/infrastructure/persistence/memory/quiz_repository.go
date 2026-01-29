package memory

import (
	"sort"
	"sync"

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

// seedData is implemented in seed_data.go

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

// FindAllSummaries retrieves all quiz summaries
func (r *QuizRepository) FindAllSummaries() ([]*quiz.QuizSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summaries := make([]*quiz.QuizSummary, 0, len(r.quizzes))
	for _, q := range r.quizzes {
		summary := quiz.NewQuizSummary(
			q.ID(),
			q.Title(),
			q.Description(),
			q.CategoryID(),
			q.TimeLimit(),
			q.PassingScore(),
			q.CreatedAt(),
			q.QuestionsCount(),
		)
		summaries = append(summaries, summary)
	}

	return summaries, nil
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

// FindAllActiveByUser retrieves all active sessions for a user
func (r *SessionRepository) FindAllActiveByUser(userID shared.UserID) ([]*quiz.QuizSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var activeSessions []*quiz.QuizSession

	for _, session := range r.sessions {
		if session.UserID().Equals(userID) && session.IsActive() {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// FindCompletedByUserQuizAndDate finds a completed session for a user, quiz, and date range
func (r *SessionRepository) FindCompletedByUserQuizAndDate(userID shared.UserID, quizID quiz.QuizID, startTime, endTime int64) (*quiz.QuizSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.UserID().Equals(userID) &&
			session.QuizID().Equals(quizID) &&
			session.Status() == quiz.SessionStatusCompleted &&
			session.CompletedAt() >= startTime &&
			session.CompletedAt() < endTime {
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

// Delete removes a session by ID
func (r *SessionRepository) Delete(id quiz.SessionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[id.String()]; !exists {
		return quiz.ErrSessionNotFound
	}

	delete(r.sessions, id.String())
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
// Shows only the BEST score per user (if user completed quiz multiple times)
func (r *LeaderboardRepository) GetLeaderboard(quizID quiz.QuizID, limit int) ([]quiz.LeaderboardEntry, error) {
	r.sessionRepo.mu.RLock()
	defer r.sessionRepo.mu.RUnlock()

	// Collect best scores per user for this quiz
	type sessionScore struct {
		userID      shared.UserID
		score       quiz.Points
		completedAt int64
	}

	// Map: userID -> best session score
	userBestScores := make(map[string]sessionScore)

	for _, session := range r.sessionRepo.sessions {
		if session.QuizID().Equals(quizID) && session.IsCompleted() {
			userIDStr := session.UserID().String()

			// Check if this is the user's first score or a better score
			if existing, found := userBestScores[userIDStr]; !found || session.Score().Value() > existing.score.Value() {
				userBestScores[userIDStr] = sessionScore{
					userID:      session.UserID(),
					score:       session.Score(),
					completedAt: session.CompletedAt(),
				}
			} else if session.Score().Value() == existing.score.Value() {
				// If scores are equal, keep the earlier completion time
				if session.CompletedAt() < existing.completedAt {
					userBestScores[userIDStr] = sessionScore{
						userID:      session.UserID(),
						score:       session.Score(),
						completedAt: session.CompletedAt(),
					}
				}
			}
		}
	}

	// Convert map to slice
	scores := make([]sessionScore, 0, len(userBestScores))
	for _, s := range userBestScores {
		scores = append(scores, s)
	}

	// Sort by score descending, then by completed_at ascending (earlier is better)
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score.Value() == scores[j].score.Value() {
			return scores[i].completedAt < scores[j].completedAt
		}
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

// GetGlobalLeaderboard retrieves top scores across all quizzes
// Shows sum of best scores per quiz for each user
func (r *LeaderboardRepository) GetGlobalLeaderboard(limit int) ([]quiz.GlobalLeaderboardEntry, error) {
	r.sessionRepo.mu.RLock()
	defer r.sessionRepo.mu.RUnlock()

	// Map: userID -> map[quizID -> best score]
	userQuizScores := make(map[string]map[string]quiz.Points)

	// Collect best scores per user per quiz
	for _, session := range r.sessionRepo.sessions {
		if !session.IsCompleted() {
			continue
		}

		userIDStr := session.UserID().String()
		quizIDStr := session.QuizID().String()

		if _, exists := userQuizScores[userIDStr]; !exists {
			userQuizScores[userIDStr] = make(map[string]quiz.Points)
		}

		// Keep best score for this quiz
		if existingScore, found := userQuizScores[userIDStr][quizIDStr]; !found || session.Score().Value() > existingScore.Value() {
			userQuizScores[userIDStr][quizIDStr] = session.Score()
		}
	}

	// Calculate total scores
	type userTotal struct {
		userID           shared.UserID
		totalScore       quiz.Points
		quizzesCompleted int
		lastActivityAt   int64
	}

	userTotals := make([]userTotal, 0)
	for userIDStr, quizScores := range userQuizScores {
		totalPoints := 0
		var lastActivity int64

		for _, score := range quizScores {
			totalPoints += score.Value()
		}

		// Find last activity time
		for _, session := range r.sessionRepo.sessions {
			if session.UserID().String() == userIDStr && session.IsCompleted() {
				if session.CompletedAt() > lastActivity {
					lastActivity = session.CompletedAt()
				}
			}
		}

		totalScoreVO, _ := quiz.NewPoints(totalPoints)
		userID, _ := shared.NewUserID(userIDStr)

		userTotals = append(userTotals, userTotal{
			userID:           userID,
			totalScore:       totalScoreVO,
			quizzesCompleted: len(quizScores),
			lastActivityAt:   lastActivity,
		})
	}

	// Sort by total score descending, then by last activity ascending
	sort.Slice(userTotals, func(i, j int) bool {
		if userTotals[i].totalScore.Value() == userTotals[j].totalScore.Value() {
			return userTotals[i].lastActivityAt < userTotals[j].lastActivityAt
		}
		return userTotals[i].totalScore.Value() > userTotals[j].totalScore.Value()
	})

	// Limit results
	if len(userTotals) > limit {
		userTotals = userTotals[:limit]
	}

	// Convert to GlobalLeaderboardEntry
	entries := make([]quiz.GlobalLeaderboardEntry, 0, len(userTotals))
	for i, ut := range userTotals {
		entry := quiz.NewGlobalLeaderboardEntry(
			ut.userID,
			"User "+ut.userID.String()[:8], // Simplified username
			ut.totalScore,
			ut.quizzesCompleted,
			i+1, // Rank
			ut.lastActivityAt,
		)
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetUserGlobalRank retrieves a user's rank in the global leaderboard
func (r *LeaderboardRepository) GetUserGlobalRank(userID shared.UserID) (int, error) {
	entries, err := r.GetGlobalLeaderboard(1000) // Get all
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.UserID().Equals(userID) {
			return entry.Rank(), nil
		}
	}

	return 0, nil // User not found = rank 0
}
