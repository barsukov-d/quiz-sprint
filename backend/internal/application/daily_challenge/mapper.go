package daily_challenge

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// ========================================
// Domain → DTO Mappers
// ========================================

// ToDailyQuizDTO converts domain DailyQuiz to DTO
func ToDailyQuizDTO(dailyQuiz *daily_challenge.DailyQuiz) DailyQuizDTO {
	questionIDs := make([]string, len(dailyQuiz.QuestionIDs()))
	for i, qid := range dailyQuiz.QuestionIDs() {
		questionIDs[i] = qid.String()
	}

	return DailyQuizDTO{
		ID:          dailyQuiz.ID().String(),
		Date:        dailyQuiz.Date().String(),
		QuestionIDs: questionIDs,
		ExpiresAt:   dailyQuiz.ExpiresAt(),
		CreatedAt:   dailyQuiz.CreatedAt(),
	}
}

// ToDailyGameDTO converts domain DailyGame to DTO
func ToDailyGameDTO(game *daily_challenge.DailyGame, now int64) DailyGameDTO {
	session := game.Session()
	streak := game.Streak()

	// Current question (only if in progress)
	var currentQuestion *QuestionDTO
	var questionIndex int
	if game.Status() == daily_challenge.GameStatusInProgress {
		if q, err := session.GetCurrentQuestion(); err == nil {
			questionDTO := ToQuestionDTO(q)
			currentQuestion = &questionDTO
		}
		questionIndex = session.CurrentQuestionIndex()
	} else {
		questionIndex = session.Quiz().QuestionsCount() // All answered
	}

	// Time remaining (Note: ExpiresAt is on DailyQuiz, not Quiz)
	// For now, we don't have direct access to DailyQuiz here
	timeRemaining := int64(0)

	gameIDString := game.ID().String()
	return DailyGameDTO{
		ID:              gameIDString, // Deprecated, use GameID
		GameID:          gameIDString, // Matches Swagger spec
		PlayerID:        game.PlayerID().String(),
		DailyQuizID:     game.DailyQuizID().String(),
		Date:            game.Date().String(),
		Status:          string(game.Status()),
		CurrentQuestion: currentQuestion,
		QuestionIndex:   questionIndex,
		TotalQuestions:  session.Quiz().QuestionsCount(),
		BaseScore:       session.BaseScore().Value(),
		FinalScore:      game.GetFinalScore(),
		CorrectAnswers:  game.GetCorrectAnswersCount(),
		Streak:          ToStreakDTO(streak, game.Date()),
		Rank:            game.Rank(),
		TimeRemaining:   timeRemaining,
	}
}

// ToStreakDTO converts domain StreakSystem to DTO
func ToStreakDTO(streak daily_challenge.StreakSystem, today daily_challenge.Date) StreakDTO {
	bonusMultiplier := streak.GetBonus()
	bonusPercent := int((bonusMultiplier - 1.0) * 100)

	return StreakDTO{
		CurrentStreak:  streak.CurrentStreak(),
		BestStreak:     streak.BestStreak(),
		LastPlayedDate: streak.LastPlayedDate().String(),
		BonusPercent:   bonusPercent,
		IsActive:       streak.IsActive(today),
	}
}

// ToQuestionDTO converts domain Question to DTO (WITHOUT IsCorrect on answers)
func ToQuestionDTO(question *quiz.Question) QuestionDTO {
	answers := make([]AnswerDTO, len(question.Answers()))
	for i, answer := range question.Answers() {
		answers[i] = AnswerDTO{
			ID:       answer.ID().String(),
			Text:     answer.Text().String(),
			Position: answer.Position(),
		}
	}

	return QuestionDTO{
		ID:       question.ID().String(),
		Text:     question.Text().String(),
		Answers:  answers,
		Points:   question.Points().Value(),
		Position: question.Position(),
	}
}

// ToAnsweredQuestionDTO converts session answer history to DTO (WITH correctness feedback)
func ToAnsweredQuestionDTO(
	question *quiz.Question,
	answerData kernel.AnswerData,
) AnsweredQuestionDTO {
	// Find correct answer and player answer text
	var correctAnswerID, correctAnswerText, playerAnswerText string
	for _, answer := range question.Answers() {
		if answer.IsCorrect() {
			correctAnswerID = answer.ID().String()
			correctAnswerText = answer.Text().String()
		}
		if answer.ID().String() == answerData.AnswerID().String() {
			playerAnswerText = answer.Text().String()
		}
	}

	// Calculate points earned
	pointsEarned := 0
	if answerData.IsCorrect() {
		pointsEarned = question.Points().Value()
	}

	return AnsweredQuestionDTO{
		QuestionID:        question.ID().String(),
		QuestionText:      question.Text().String(),
		PlayerAnswerID:    answerData.AnswerID().String(),
		PlayerAnswerText:  playerAnswerText,
		CorrectAnswerID:   correctAnswerID,
		CorrectAnswerText: correctAnswerText,
		IsCorrect:         answerData.IsCorrect(),
		TimeTaken:         answerData.TimeTaken(),
		PointsEarned:      pointsEarned,
	}
}

// ToLeaderboardEntryDTO converts domain DailyGame to leaderboard entry
func ToLeaderboardEntryDTO(game *daily_challenge.DailyGame, username string, rank int) LeaderboardEntryDTO {
	return LeaderboardEntryDTO{
		PlayerID:       game.PlayerID().String(),
		Username:       username,
		Score:          game.GetFinalScore(),
		CorrectAnswers: game.GetCorrectAnswersCount(),
		Rank:           rank,
		StreakDays:     game.Streak().CurrentStreak(),
		CompletedAt:    game.Session().FinishedAt(),
	}
}

// BuildGameResultsDTO creates game results with full answer breakdown
func BuildGameResultsDTO(
	game *daily_challenge.DailyGame,
	rank int,
	totalPlayers int,
	leaderboard []LeaderboardEntryDTO,
) GameResultsDTO {
	session := game.Session()
	streak := game.Streak()

	// Get all answered questions with correctness
	answeredQuestions := make([]AnsweredQuestionDTO, 0)
	for i := 0; i < session.Quiz().QuestionsCount(); i++ {
		if question, err := session.Quiz().GetQuestionByIndex(i); err == nil {
			if answerData, exists := session.GetAnswer(question.ID()); exists {
				answeredQuestions = append(answeredQuestions, ToAnsweredQuestionDTO(question, answerData))
			}
		}
	}

	// Calculate percentile
	percentile := 100
	if totalPlayers > 0 {
		percentile = int((float64(totalPlayers - rank + 1) / float64(totalPlayers)) * 100)
	}

	bonusMultiplier := streak.GetBonus()
	bonusPercent := int((bonusMultiplier - 1.0) * 100)

	return GameResultsDTO{
		BaseScore:         session.BaseScore().Value(),
		FinalScore:        game.GetFinalScore(),
		CorrectAnswers:    game.GetCorrectAnswersCount(),
		TotalQuestions:    session.Quiz().QuestionsCount(),
		StreakBonus:       bonusPercent,
		CurrentStreak:     streak.CurrentStreak(),
		Rank:              rank,
		TotalPlayers:      totalPlayers,
		Percentile:        percentile,
		AnsweredQuestions: answeredQuestions,
		Leaderboard:       leaderboard,
	}
}
