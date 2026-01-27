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

// ToMarathonGameDTO converts a MarathonGame aggregate to DTO
func ToMarathonGameDTO(game *solo_marathon.MarathonGame, now int64) MarathonGameDTO {
	// Get current question if game is in progress
	var currentQuestion *QuestionDTO
	if game.Status() == solo_marathon.GameStatusInProgress {
		if q, err := game.GetCurrentQuestion(); err == nil {
			dto := quiz.ToQuestionDTO(q)
			currentQuestion = &QuestionDTO{
				ID:       dto.ID,
				Text:     dto.Text,
				Answers:  toAnswerDTOs(dto.Answers),
				Points:   dto.Points,
				Position: dto.Position,
			}
		}
	}

	return MarathonGameDTO{
		ID:                 game.ID().String(),
		PlayerID:           game.PlayerID().String(),
		Category:           ToCategoryDTO(game.Category()),
		Status:             string(game.Status()),
		CurrentStreak:      game.CurrentStreak(),
		MaxStreak:          game.MaxStreak(),
		Lives:              ToLivesDTO(game.Lives(), now),
		Hints:              ToHintsDTO(game.Hints()),
		DifficultyLevel:    string(game.Difficulty().Level()),
		PersonalBestStreak: game.PersonalBestStreak(),
		CurrentQuestion:    currentQuestion,
		BaseScore:          game.Session().BaseScore().Value(),
	}
}

// ToMarathonGameDTOV2 converts a MarathonGameV2 aggregate to DTO
func ToMarathonGameDTOV2(game *solo_marathon.MarathonGameV2, now int64) MarathonGameDTO {
	// Get current question if game is in progress
	var currentQuestion *QuestionDTO
	if game.Status() == solo_marathon.GameStatusInProgress {
		if q, err := game.GetCurrentQuestion(); err == nil {
			dto := quiz.ToQuestionDTO(q)
			currentQuestion = &QuestionDTO{
				ID:       dto.ID,
				Text:     dto.Text,
				Answers:  toAnswerDTOs(dto.Answers),
				Points:   dto.Points,
				Position: dto.Position,
			}
		}
	}

	return MarathonGameDTO{
		ID:                 game.ID().String(),
		PlayerID:           game.PlayerID().String(),
		Category:           ToCategoryDTO(game.Category()),
		Status:             string(game.Status()),
		CurrentStreak:      game.CurrentStreak(),
		MaxStreak:          game.MaxStreak(),
		Lives:              ToLivesDTO(game.Lives(), now),
		Hints:              ToHintsDTO(game.Hints()),
		DifficultyLevel:    string(game.Difficulty().Level()),
		PersonalBestStreak: game.PersonalBestStreak(),
		CurrentQuestion:    currentQuestion,
		BaseScore:          game.BaseScore(), // Note: V2 stores baseScore directly
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
	}
}

// ToHintsDTO converts HintsSystem to DTO
func ToHintsDTO(hints solo_marathon.HintsSystem) HintsDTO {
	return HintsDTO{
		FiftyFifty: hints.FiftyFifty(),
		ExtraTime:  hints.ExtraTime(),
		Skip:       hints.Skip(),
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
// Note: Username must be provided separately from user repository
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

// BuildGameOverResult builds a GameOverResultDTO from game state (V1)
func BuildGameOverResult(
	game *solo_marathon.MarathonGame,
	globalRank *int,
) GameOverResultDTO {
	return GameOverResultDTO{
		FinalStreak:       game.MaxStreak(),
		IsNewPersonalBest: game.IsNewPersonalBest(),
		PreviousRecord:    game.PersonalBestStreak(),
		TotalBaseScore:    game.Session().BaseScore().Value(),
		GlobalRank:        globalRank,
	}
}

// BuildGameOverResultV2 builds a GameOverResultDTO from game state (V2)
func BuildGameOverResultV2(
	game *solo_marathon.MarathonGameV2,
	globalRank *int,
) GameOverResultDTO {
	return GameOverResultDTO{
		FinalStreak:       game.MaxStreak(),
		IsNewPersonalBest: game.IsNewPersonalBest(),
		PreviousRecord:    game.PersonalBestStreak(),
		TotalBaseScore:    game.BaseScore(),
		GlobalRank:        globalRank,
	}
}

// ========================================
// Utility Functions
// ========================================

// GetTimeLimit calculates the time limit for current question based on streak
func GetTimeLimit(difficulty solo_marathon.DifficultyProgression, currentStreak int) int {
	return difficulty.GetTimeLimit(currentStreak)
}

// GetCurrentTimestamp returns current Unix timestamp
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// FindCorrectAnswerID finds the correct answer ID for a question
func FindCorrectAnswerID(q *domainQuiz.Question) string {
	return quiz.FindCorrectAnswerID(q)
}
