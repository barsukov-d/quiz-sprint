package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// ========================================
// Domain → DTO Mappers
// ========================================

// ToMarathonGameDTO converts a MarathonGame V1 aggregate to DTO
func ToMarathonGameDTO(game *solo_marathon.MarathonGame, now int64) MarathonGameDTO {
	// Get current question if game is in progress
	var currentQuestion *QuestionDTO
	if game.Status() == solo_marathon.GameStatusInProgress {
		if q, err := game.GetCurrentQuestion(); err == nil {
			dto := quiz.ToQuestionDTO(q)
			questionDTO := QuestionDTO{
				ID:       dto.ID,
				Text:     dto.Text,
				Answers:  toAnswerDTOs(dto.Answers),
				Points:   dto.Points,
				Position: dto.Position,
			}
			currentQuestion = &questionDTO
		}
	}

	return MarathonGameDTO{
		ID:              game.ID().String(),
		PlayerID:        game.PlayerID().String(),
		Category:        ToCategoryDTO(game.Category()),
		Status:          string(game.Status()),
		Score:           game.CurrentStreak(),
		TotalQuestions:  0, // V1 doesn't track total questions
		Lives:           ToLivesDTO(game.Lives(), now),
		BonusInventory:  ToBonusInventoryDTO(game.Hints()),
		ShieldActive:    false, // V1 doesn't support shield
		DifficultyLevel: string(game.Difficulty().Level()),
		ContinueCount:   0, // V1 doesn't support continue
		PersonalBest:    game.PersonalBestStreak(),
		CurrentQuestion: currentQuestion,
		QuestionNumber:  game.CurrentStreak() + 1,
		TimeLimit:       game.Difficulty().GetTimeLimit(game.CurrentStreak() + 1),
	}
}

// ToMarathonGameDTOV2 converts a MarathonGameV2 aggregate to DTO
func ToMarathonGameDTOV2(game *solo_marathon.MarathonGameV2, now int64) MarathonGameDTO {
	// Get current question if game is in progress
	var currentQuestion *QuestionDTO
	if game.Status() == solo_marathon.GameStatusInProgress {
		if q, err := game.GetCurrentQuestion(); err == nil {
			dto := quiz.ToQuestionDTO(q)
			questionDTO := QuestionDTO{
				ID:       dto.ID,
				Text:     dto.Text,
				Answers:  toAnswerDTOs(dto.Answers),
				Points:   dto.Points,
				Position: dto.Position,
			}
			currentQuestion = &questionDTO
		}
	}

	questionNumber := game.QuestionNumber()
	timeLimit := game.Difficulty().GetTimeLimit(questionNumber)

	return MarathonGameDTO{
		ID:              game.ID().String(),
		PlayerID:        game.PlayerID().String(),
		Category:        ToCategoryDTO(game.Category()),
		Status:          string(game.Status()),
		Score:           game.Score(),
		TotalQuestions:  game.TotalQuestions(),
		Lives:           ToLivesDTO(game.Lives(), now),
		BonusInventory:  ToBonusInventoryDTO(game.BonusInventory()),
		ShieldActive:    game.ShieldActive(),
		DifficultyLevel: string(game.Difficulty().Level()),
		ContinueCount:   game.ContinueCount(),
		PersonalBest:    game.PersonalBestScore(),
		CurrentQuestion: currentQuestion,
		QuestionNumber:  questionNumber,
		TimeLimit:       timeLimit,
	}
}

// ToCategoryDTO converts a MarathonCategory to DTO
func ToCategoryDTO(category solo_marathon.MarathonCategory) CategoryDTO {
	return CategoryDTO{
		ID:              category.CategoryID().String(),
		Name:            category.Name(),
		IsAllCategories: category.IsAllCategories(),
	}
}

// ToLivesDTO converts LivesSystem to DTO
func ToLivesDTO(lives solo_marathon.LivesSystem, now int64) LivesDTO {
	return LivesDTO{
		CurrentLives:   lives.CurrentLives(),
		MaxLives:       lives.MaxLives(),
		TimeToNextLife: lives.TimeToNextLife(now),
		Label:          lives.Label(),
	}
}

// ToBonusInventoryDTO converts BonusInventory to DTO
func ToBonusInventoryDTO(bonuses solo_marathon.BonusInventory) BonusInventoryDTO {
	return BonusInventoryDTO{
		Shield:     bonuses.Shield(),
		FiftyFifty: bonuses.FiftyFifty(),
		Skip:       bonuses.Skip(),
		Freeze:     bonuses.Freeze(),
	}
}

// ToContinueOfferDTO converts domain ContinueOffer to DTO
func ToContinueOfferDTO(offer *solo_marathon.ContinueOffer) *ContinueOfferDTO {
	if offer == nil {
		return nil
	}
	return &ContinueOfferDTO{
		Available:     offer.Available,
		CostCoins:     offer.CostCoins,
		HasAd:         offer.HasAd,
		ContinueCount: offer.ContinueCount,
	}
}

// ToMilestoneDTO builds milestone progress from current score
func ToMilestoneDTO(currentScore int) *MilestoneDTO {
	next, remaining := solo_marathon.GetNextMilestone(currentScore)
	if next == 0 {
		return nil // Past all milestones
	}
	return &MilestoneDTO{
		Next:      next,
		Current:   currentScore,
		Remaining: remaining,
	}
}

// ToQuestionDTO converts a quiz Question to marathon QuestionDTO
func ToQuestionDTO(q *domainQuiz.Question) QuestionDTO {
	dto := quiz.ToQuestionDTO(q)
	return QuestionDTO{
		ID:       dto.ID,
		Text:     dto.Text,
		Answers:  toAnswerDTOs(dto.Answers),
		Points:   dto.Points,
		Position: dto.Position,
	}
}

// toAnswerDTOs converts quiz AnswerDTOs to marathon AnswerDTOs
func toAnswerDTOs(quizAnswers []quiz.AnswerDTO) []AnswerDTO {
	answers := make([]AnswerDTO, 0, len(quizAnswers))
	for _, a := range quizAnswers {
		answers = append(answers, AnswerDTO{
			ID:       a.ID,
			Text:     a.Text,
			Position: a.Position,
		})
	}
	return answers
}

// ToPersonalBestDTO converts a PersonalBest aggregate to DTO
func ToPersonalBestDTO(pb *solo_marathon.PersonalBest) PersonalBestDTO {
	return PersonalBestDTO{
		Category:   ToCategoryDTO(pb.Category()),
		BestStreak: pb.BestStreak(),
		BestScore:  pb.BestScore(),
		AchievedAt: pb.AchievedAt(),
	}
}

// ToPersonalBestDTOs converts a slice of PersonalBest to DTOs
func ToPersonalBestDTOs(personalBests []*solo_marathon.PersonalBest) []PersonalBestDTO {
	dtos := make([]PersonalBestDTO, 0, len(personalBests))
	for _, pb := range personalBests {
		dtos = append(dtos, ToPersonalBestDTO(pb))
	}
	return dtos
}

// ToLeaderboardEntryDTO converts PersonalBest to LeaderboardEntryDTO
func ToLeaderboardEntryDTO(pb *solo_marathon.PersonalBest, username string, rank int) LeaderboardEntryDTO {
	return LeaderboardEntryDTO{
		PlayerID:   pb.PlayerID().String(),
		Username:   username,
		BestStreak: pb.BestStreak(),
		BestScore:  pb.BestScore(),
		Rank:       rank,
		AchievedAt: pb.AchievedAt(),
	}
}

// ========================================
// Result Builders
// ========================================

// BuildGameOverResult builds a GameOverResultDTO from V1 game state
func BuildGameOverResult(
	game *solo_marathon.MarathonGame,
) GameOverResultDTO {
	return GameOverResultDTO{
		FinalScore:        game.MaxStreak(),
		TotalQuestions:    0, // V1 doesn't track
		IsNewPersonalBest: game.IsNewPersonalBest(),
		PreviousRecord:    game.PersonalBestStreak(),
	}
}

// BuildGameOverResultV2 builds a GameOverResultDTO from V2 game state
func BuildGameOverResultV2(game *solo_marathon.MarathonGameV2) GameOverResultDTO {
	result := GameOverResultDTO{
		FinalScore:        game.Score(),
		TotalQuestions:    game.TotalQuestions(),
		IsNewPersonalBest: game.IsNewPersonalBest(),
		PreviousRecord:    game.PersonalBestScore(),
	}

	// Build continue offer if game is in game_over state
	if game.IsWaitingForContinue() {
		costCalc := solo_marathon.ContinueCostCalculator{}
		result.ContinueOffer = &ContinueOfferDTO{
			Available:     true,
			CostCoins:     costCalc.GetCost(game.ContinueCount()),
			HasAd:         costCalc.HasAdOption(game.ContinueCount()),
			ContinueCount: game.ContinueCount(),
		}
	}

	return result
}

// ========================================
// Utility Functions
// ========================================

// GetTimeLimit calculates the time limit for a question based on question index
func GetTimeLimit(difficulty solo_marathon.DifficultyProgression, questionIndex int) int {
	return difficulty.GetTimeLimit(questionIndex)
}

// GetCurrentTimestamp returns current Unix timestamp
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// FindCorrectAnswerID finds the correct answer ID for a question
func FindCorrectAnswerID(q *domainQuiz.Question) string {
	return quiz.FindCorrectAnswerID(q)
}
