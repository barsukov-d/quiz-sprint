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
	var questionTimeRemaining *int
	if game.Status() == daily_challenge.GameStatusInProgress {
		if q, err := session.GetCurrentQuestion(); err == nil {
			questionDTO := ToQuestionDTO(q)
			currentQuestion = &questionDTO
		}
		questionIndex = session.CurrentQuestionIndex()

		// Calculate time remaining for current question
		// timeRemaining = timeLimit - (now - questionStartedAt)
		timeLimit := 15
		elapsed := now - game.QuestionStartedAt()
		remaining := int64(timeLimit) - elapsed

		// Clamp to [0, timeLimit]
		if remaining < 0 {
			remaining = 0
		}
		if remaining > int64(timeLimit) {
			remaining = int64(timeLimit)
		}

		tr := int(remaining)
		questionTimeRemaining = &tr
	} else {
		questionIndex = session.Quiz().QuestionsCount() // All answered
	}

	// Time remaining (Note: ExpiresAt is on DailyQuiz, not Quiz)
	// For now, we don't have direct access to DailyQuiz here
	timeRemaining := int64(0)

	gameIDString := game.ID().String()
	return DailyGameDTO{
		ID:                    gameIDString, // Deprecated, use GameID
		GameID:                gameIDString, // Matches Swagger spec
		PlayerID:              game.PlayerID().String(),
		DailyQuizID:           game.DailyQuizID().String(),
		Date:                  game.Date().String(),
		Status:                string(game.Status()),
		CurrentQuestion:       currentQuestion,
		QuestionIndex:         questionIndex,
		TotalQuestions:        session.Quiz().QuestionsCount(),
		BaseScore:             session.BaseScore().Value(),
		FinalScore:            game.GetFinalScore(),
		CorrectAnswers:        game.GetCorrectAnswersCount(),
		Streak:                ToStreakDTO(streak, game.Date()),
		Rank:                  game.Rank(),
		TimeRemaining:         timeRemaining,
		QuestionTimeRemaining: questionTimeRemaining,
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
// quizAggregate is needed to calculate time bonus (same logic as kernel.AnswerQuestion)
func ToAnsweredQuestionDTO(
	question *quiz.Question,
	answerData kernel.AnswerData,
	quizAggregate *quiz.Quiz,
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

	// Calculate points earned (base + time bonus, same logic as quiz_gameplay_session.go)
	pointsEarned := 0
	if answerData.IsCorrect() {
		// Base points: use question-level, fallback to quiz-level
		basePoints := question.Points().Value()
		if basePoints == 0 {
			basePoints = quizAggregate.BasePoints().Value()
		}

		// Time bonus: max(0, (timeLimit - timeTaken) * maxTimeBonus / timeLimit)
		timeBonus := 0
		timeTaken := answerData.TimeTaken()
		timeLimitMs := int64(quizAggregate.TimeLimitPerQuestion()) * 1000
		if timeTaken > 0 && timeTaken < timeLimitMs && quizAggregate.MaxTimeBonus().Value() > 0 {
			remaining := timeLimitMs - timeTaken
			timeBonus = int(float64(quizAggregate.MaxTimeBonus().Value()) * float64(remaining) / float64(timeLimitMs))
		}

		pointsEarned = basePoints + timeBonus
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

// ToChestRewardDTO converts domain ChestReward to DTO
func ToChestRewardDTO(reward daily_challenge.ChestReward) ChestRewardDTO {
	bonuses := make([]string, len(reward.MarathonBonuses()))
	for i, bonus := range reward.MarathonBonuses() {
		bonuses[i] = bonus.String()
	}

	return ChestRewardDTO{
		ChestType:       reward.ChestType().String(),
		Coins:           reward.Coins(),
		PvpTickets:      reward.PvpTickets(),
		MarathonBonuses: bonuses,
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
				answeredQuestions = append(answeredQuestions, ToAnsweredQuestionDTO(question, answerData, session.Quiz()))
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

	// Get chest reward (should be set by use case)
	chestReward := ChestRewardDTO{
		ChestType:       "wooden",
		Coins:           0,
		PvpTickets:      0,
		MarathonBonuses: []string{},
	}
	if game.ChestReward() != nil {
		chestReward = ToChestRewardDTO(*game.ChestReward())
	}

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
		ChestReward:       chestReward,
		AnsweredQuestions: answeredQuestions,
		Leaderboard:       leaderboard,
	}
}
