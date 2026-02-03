package daily_challenge

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"time"
)

// GetOrCreateDailyQuizUseCase ensures daily quiz exists for a date
type GetOrCreateDailyQuizUseCase struct {
	dailyQuizRepo daily_challenge.DailyQuizRepository
	dailyGameRepo daily_challenge.DailyGameRepository
	questionRepo  quiz.QuestionRepository
	eventBus      EventBus
}

func NewGetOrCreateDailyQuizUseCase(
	dailyQuizRepo daily_challenge.DailyQuizRepository,
	dailyGameRepo daily_challenge.DailyGameRepository,
	questionRepo quiz.QuestionRepository,
	eventBus EventBus,
) *GetOrCreateDailyQuizUseCase {
	return &GetOrCreateDailyQuizUseCase{
		dailyQuizRepo: dailyQuizRepo,
		dailyGameRepo: dailyGameRepo,
		questionRepo:  questionRepo,
		eventBus:      eventBus,
	}
}

func (uc *GetOrCreateDailyQuizUseCase) Execute(input GetOrCreateDailyQuizInput) (GetOrCreateDailyQuizOutput, error) {
	// 1. Determine date
	var date daily_challenge.Date
	if input.Date != "" {
		date = daily_challenge.NewDateFromString(input.Date)
	} else {
		date = daily_challenge.TodayUTC()
	}

	now := time.Now().UTC().Unix()
	println("🔍 [GetOrCreateDailyQuiz] Starting for date:", date.String())

	// 2. Try to find existing daily quiz
	existingQuiz, err := uc.dailyQuizRepo.FindByDate(date)
	if err == nil && existingQuiz != nil {
		// Quiz exists
		println("✅ [GetOrCreateDailyQuiz] Found existing quiz:", existingQuiz.ID().String())
		totalPlayers, _ := uc.dailyGameRepo.GetTotalPlayersByDate(date)

		return GetOrCreateDailyQuizOutput{
			DailyQuiz:    ToDailyQuizDTO(existingQuiz),
			TotalPlayers: totalPlayers,
			IsNew:        false,
		}, nil
	}

	println("⚠️  [GetOrCreateDailyQuiz] No existing quiz found, creating new one...")

	// 3. Daily quiz doesn't exist - generate it
	dailyQuiz, err := uc.generateDailyQuiz(date, now)
	if err != nil {
		println("❌ [GetOrCreateDailyQuiz] Failed to generate quiz:", err.Error())
		return GetOrCreateDailyQuizOutput{}, err
	}

	println("✅ [GetOrCreateDailyQuiz] Generated quiz with ID:", dailyQuiz.ID().String())

	// 4. Save daily quiz
	if err := uc.dailyQuizRepo.Save(dailyQuiz); err != nil {
		println("❌ [GetOrCreateDailyQuiz] Failed to save quiz:", err.Error())
		return GetOrCreateDailyQuizOutput{}, err
	}

	println("✅ [GetOrCreateDailyQuiz] Successfully saved quiz to database")

	// 5. Publish events
	for _, event := range dailyQuiz.Events() {
		uc.eventBus.Publish(event)
	}

	return GetOrCreateDailyQuizOutput{
		DailyQuiz:    ToDailyQuizDTO(dailyQuiz),
		TotalPlayers: 0, // New quiz, no players yet
		IsNew:        true,
	}, nil
}

// generateDailyQuiz generates a new daily quiz with 10 questions
// Uses deterministic seed based on date for consistency
func (uc *GetOrCreateDailyQuizUseCase) generateDailyQuiz(
	date daily_challenge.Date,
	now int64,
) (*daily_challenge.DailyQuiz, error) {
	// Calculate expiry (next day 00:00 UTC)
	dateTime, _ := time.Parse("2006-01-02", date.String())
	nextDay := dateTime.AddDate(0, 0, 1)
	expiresAt := nextDay.Unix()

	println("📊 [generateDailyQuiz] Selecting questions...")

	// Generate deterministic seed from date
	// This ensures ALL players worldwide get the SAME quiz for a given date
	seed := date.ToSeed()
	println("🎲 [generateDailyQuiz] Using seed:", seed, "for date:", date.String())

	// Select a whole quiz with exactly 10 questions (all questions share a common theme)
	questions, err := uc.questionRepo.FindQuestionsByQuizSeed(daily_challenge.QuestionsPerDay, seed)
	if err != nil {
		println("⚠️  [generateDailyQuiz] No quiz with exactly", daily_challenge.QuestionsPerDay, "questions, falling back to random selection")
		// Fallback: random questions from global pool
		filter := quiz.NewQuestionFilter()
		questions, err = uc.questionRepo.FindQuestionsBySeed(filter, daily_challenge.QuestionsPerDay, seed)
		if err != nil {
			println("❌ [generateDailyQuiz] Failed to find questions:", err.Error())
			return nil, err
		}
	}

	println("📊 [generateDailyQuiz] Found", len(questions), "questions (need", daily_challenge.QuestionsPerDay, ")")

	if len(questions) < daily_challenge.QuestionsPerDay {
		println("❌ [generateDailyQuiz] Not enough questions")
		return nil, daily_challenge.ErrInvalidDate // Not enough questions
	}

	// Extract question IDs
	questionIDs := make([]daily_challenge.QuestionID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID()
	}

	println("✅ [generateDailyQuiz] Creating DailyQuiz aggregate...")

	// Create daily quiz
	return daily_challenge.NewDailyQuiz(date, questionIDs, expiresAt, now)
}
