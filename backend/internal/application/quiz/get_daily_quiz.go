package quiz

import (
	"crypto/md5"
	"encoding/binary"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// GetDailyQuizUseCase handles the business logic for getting the daily quiz
type GetDailyQuizUseCase struct {
	quizRepo        quiz.QuizRepository
	sessionRepo     quiz.SessionRepository
	leaderboardRepo quiz.LeaderboardRepository
}

// NewGetDailyQuizUseCase creates a new GetDailyQuizUseCase
func NewGetDailyQuizUseCase(
	quizRepo quiz.QuizRepository,
	sessionRepo quiz.SessionRepository,
	leaderboardRepo quiz.LeaderboardRepository,
) *GetDailyQuizUseCase {
	return &GetDailyQuizUseCase{
		quizRepo:        quizRepo,
		sessionRepo:     sessionRepo,
		leaderboardRepo: leaderboardRepo,
	}
}

// Execute retrieves the daily quiz for today with completion status
// Uses deterministic selection: hash(date) % totalQuizzes
func (uc *GetDailyQuizUseCase) Execute(input GetDailyQuizInput) (GetDailyQuizOutput, error) {
	// 1. Parse userID
	userID, err := shared.NewUserID(input.UserID)
	if err != nil {
		return GetDailyQuizOutput{}, err
	}

	// 2. Get today's date (UTC)
	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	// Start and end of today (UTC)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// 3. Get all quiz summaries
	summaries, err := uc.quizRepo.FindAllSummaries()
	if err != nil {
		return GetDailyQuizOutput{}, err
	}

	if len(summaries) == 0 {
		return GetDailyQuizOutput{}, quiz.ErrQuizNotFound
	}

	// 4. Deterministic selection: hash(date) % count
	selectedIndex := hashDateToIndex(today, len(summaries))
	selectedQuizID := summaries[selectedIndex].ID()

	// 5. Load full quiz details
	quizAggregate, err := uc.quizRepo.FindByID(selectedQuizID)
	if err != nil {
		return GetDailyQuizOutput{}, err
	}

	// 6. Check if user completed this quiz today
	completionStatus := "not_attempted"
	var userResult *DailyQuizUserResultDTO

	completedSession, err := uc.sessionRepo.FindCompletedByUserQuizAndDate(
		userID,
		selectedQuizID,
		startOfDay.Unix(),
		endOfDay.Unix(),
	)

	if err == nil && completedSession != nil {
		// User completed the quiz today
		completionStatus = "completed"

		// Get user's rank
		rank, err := uc.leaderboardRepo.GetUserRank(selectedQuizID, userID)
		if err != nil {
			// If rank not found, default to 0
			rank = 0
		}

		userResult = &DailyQuizUserResultDTO{
			Score:       completedSession.Score().Value(),
			Rank:        rank,
			CompletedAt: completedSession.CompletedAt(),
		}
	}

	// 7. Get top scores
	leaderboard, err := uc.leaderboardRepo.GetLeaderboard(selectedQuizID, 10)
	if err != nil {
		// If leaderboard fails, continue with empty list
		leaderboard = []quiz.LeaderboardEntry{}
	}

	topScores := make([]LeaderboardEntryDTO, len(leaderboard))
	for i, entry := range leaderboard {
		topScores[i] = LeaderboardEntryDTO{
			UserID:      entry.UserID().String(),
			Username:    entry.Username(),
			Score:       entry.Score().Value(),
			Rank:        entry.Rank(),
			CompletedAt: entry.CompletedAt(),
		}
	}

	// 8. Return DTO
	return GetDailyQuizOutput{
		Quiz:             ToQuizDetailDTO(quizAggregate),
		CompletionStatus: completionStatus,
		UserResult:       userResult,
		TopScores:        topScores,
	}, nil
}

// hashDateToIndex converts date string to deterministic index
// Algorithm: MD5 hash of date string → convert to uint64 → modulo count
func hashDateToIndex(date string, count int) int {
	hash := md5.Sum([]byte(date))
	// Take first 8 bytes and convert to uint64
	num := binary.BigEndian.Uint64(hash[:8])
	return int(num % uint64(count))
}
