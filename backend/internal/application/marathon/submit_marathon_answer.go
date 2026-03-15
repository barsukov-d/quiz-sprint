package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// SubmitMarathonAnswerUseCase handles submitting an answer in marathon mode
type SubmitMarathonAnswerUseCase struct {
	marathonRepo     solo_marathon.Repository
	personalBestRepo solo_marathon.PersonalBestRepository
	questionRepo     quiz.QuestionRepository
	eventBus         EventBus
	inventoryService InventoryService
}

// NewSubmitMarathonAnswerUseCase creates a new SubmitMarathonAnswerUseCase
func NewSubmitMarathonAnswerUseCase(
	marathonRepo solo_marathon.Repository,
	personalBestRepo solo_marathon.PersonalBestRepository,
	questionRepo quiz.QuestionRepository,
	eventBus EventBus,
	inventoryService InventoryService,
) *SubmitMarathonAnswerUseCase {
	return &SubmitMarathonAnswerUseCase{
		marathonRepo:     marathonRepo,
		personalBestRepo: personalBestRepo,
		questionRepo:     questionRepo,
		eventBus:         eventBus,
		inventoryService: inventoryService,
	}
}

// milestoneRewards defines coins and bonuses awarded at each milestone
var milestoneRewards = map[int]map[string]int{
	10:  {"coins": 50},
	25:  {"coins": 100, "shield": 1},
	50:  {"coins": 250, "fifty_fifty": 1},
	75:  {"coins": 400, "freeze": 1},
	100: {"coins": 500, "shield": 1, "fifty_fifty": 1},
}

const personalBestBonusCoins = 500

// Execute submits an answer for a marathon game question
func (uc *SubmitMarathonAnswerUseCase) Execute(input SubmitMarathonAnswerInput) (SubmitMarathonAnswerOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return SubmitMarathonAnswerOutput{}, solo_marathon.ErrInvalidGameID
	}

	questionID, err := quiz.NewQuestionIDFromString(input.QuestionID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	answerID, err := quiz.NewAnswerIDFromString(input.AnswerID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// Validate timeTaken
	if input.TimeTaken < 0 {
		return SubmitMarathonAnswerOutput{}, quiz.ErrInvalidTimeTaken
	}
	if input.TimeTaken > 3600000 { // Max 1 hour in milliseconds
		return SubmitMarathonAnswerOutput{}, quiz.ErrTimeTakenTooLong
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return SubmitMarathonAnswerOutput{}, quiz.ErrUnauthorized
	}

	// 4. Capture current question before answering (to extract correct answer text)
	answeredQuestion := game.CurrentQuestion()

	// 5. Submit answer (domain business logic)
	now := time.Now().Unix()
	result, err := game.AnswerQuestion(questionID, answerID, input.TimeTaken, now)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 6. If game continues (not game_over), load next question
	if !result.IsGameOver {
		questionSelector := solo_marathon.NewQuestionSelector(uc.questionRepo)
		if err := game.LoadNextQuestion(questionSelector); err != nil {
			// Log error but don't fail - game can continue
			_ = err
		}
	}

	// Find correct answer text from the captured question
	correctAnswerText := ""
	if answeredQuestion != nil {
		for _, a := range answeredQuestion.Answers() {
			if a.IsCorrect() {
				correctAnswerText = a.Text().String()
				break
			}
		}
	}

	// 7. Build output
	output := SubmitMarathonAnswerOutput{
		IsCorrect:          result.IsCorrect,
		CorrectAnswerID:    result.CorrectAnswerID.String(),
		CorrectAnswerText:  correctAnswerText,
		TimeTaken:          result.TimeTaken,
		Score:           result.Score,
		TotalQuestions:  result.TotalQuestions,
		DifficultyLevel: string(result.DifficultyLevel),
		LifeLost:        result.LifeLost,
		ShieldConsumed:  result.ShieldConsumed,
		Lives:           ToLivesDTO(game.Lives(), now),
		BonusInventory:  ToBonusInventoryDTO(game.BonusInventory()),
		IsGameOver:      result.IsGameOver,
		Milestone:       ToMilestoneDTO(result.Score),
		StreakCount:     result.StreakCount,   // NEW
		LifeRestored:    result.LifeRestored, // NEW
	}

	// 7. Handle game over scenario (intermediate — continue offered)
	if result.IsGameOver && result.GameOverData != nil {
		output.GameOverResult = &GameOverResultDTO{
			FinalScore:        result.GameOverData.FinalScore,
			TotalQuestions:    result.GameOverData.TotalQuestions,
			IsNewPersonalBest: result.GameOverData.IsNewRecord,
			PreviousRecord:    game.PersonalBestScore(),
			ContinueOffer:     ToContinueOfferDTO(result.GameOverData.ContinueOffer),
		}
	} else if !result.IsGameOver {
		// Game continues - get next question
		nextQuestion, err := game.GetCurrentQuestion()
		if err == nil {
			nextQuestionDTO := ToQuestionDTO(nextQuestion)
			output.NextQuestion = &nextQuestionDTO

			// Calculate time limit for next question
			nextTimeLimit := GetTimeLimit(game.Difficulty(), game.QuestionNumber())
			output.NextTimeLimit = &nextTimeLimit
		}
	}

	// 8. Credit milestone rewards and personal best bonus
	if uc.inventoryService != nil && result.IsGameOver && result.GameOverData != nil {
		playerIDStr := input.PlayerID
		// Personal best bonus
		if result.GameOverData.IsNewRecord {
			_ = uc.inventoryService.Credit(playerIDStr, "personal_best", map[string]int{"coins": personalBestBonusCoins})
		}
		// Milestone rewards for all milestones reached in this game
		for milestone, rewards := range milestoneRewards {
			if result.GameOverData.FinalScore >= milestone {
				_ = uc.inventoryService.Credit(playerIDStr, "marathon_milestone", rewards)
			}
		}
	}

	// 9. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 10. Update PersonalBest if game is over and it's a new record
	// Note: PersonalBest is updated when game transitions to completed (not game_over)
	// This happens in CompleteGame or Abandon use case

	// 10. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	return output, nil
}
