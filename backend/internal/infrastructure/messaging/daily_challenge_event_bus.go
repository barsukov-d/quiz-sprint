package messaging

import (
	"log"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
)

// DailyChallengeEventBus is an in-memory implementation of daily_challenge.EventBus
type DailyChallengeEventBus struct {
	enableLogging bool
}

// NewDailyChallengeEventBus creates a new daily challenge event bus
func NewDailyChallengeEventBus(enableLogging bool) *DailyChallengeEventBus {
	return &DailyChallengeEventBus{
		enableLogging: enableLogging,
	}
}

// Publish publishes a single daily challenge event
func (eb *DailyChallengeEventBus) Publish(event daily_challenge.Event) {
	go eb.dispatch(event)
}

func (eb *DailyChallengeEventBus) dispatch(event daily_challenge.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[DAILY CHALLENGE EVENT BUS] Panic: %v", r)
		}
	}()

	if eb.enableLogging {
		eb.logEvent(event)
	}
}

func (eb *DailyChallengeEventBus) logEvent(event daily_challenge.Event) {
	switch e := event.(type) {
	case daily_challenge.DailyQuizCreatedEvent:
		log.Printf("[DAILY CHALLENGE] Quiz Created: date=%s, questions=%d", e.Date().String(), len(e.QuestionIDs()))

	case daily_challenge.DailyGameStartedEvent:
		log.Printf("[DAILY CHALLENGE] Game Started: gameId=%s, playerId=%s, date=%s, streak=%d",
			e.GameID().String(), e.PlayerID().String(), e.Date().String(), e.CurrentStreak())

	case daily_challenge.DailyQuestionAnsweredEvent:
		log.Printf("[DAILY CHALLENGE] Question Answered: gameId=%s, questionId=%s",
			e.GameID().String(), e.QuestionID().String())

	case daily_challenge.DailyGameCompletedEvent:
		log.Printf("[DAILY CHALLENGE] Game Completed: gameId=%s, playerId=%s, score=%d, correct=%d/%d, streak=%d",
			e.GameID().String(), e.PlayerID().String(), e.FinalScore(), e.CorrectAnswers(), e.TotalQuestions(), e.NewStreak())

	case daily_challenge.StreakMilestoneReachedEvent:
		log.Printf("[DAILY CHALLENGE] Streak Milestone: playerId=%s, streak=%d days, bonus=%d%%",
			e.PlayerID().String(), e.StreakDays(), e.BonusPercent())

	default:
		log.Printf("[DAILY CHALLENGE] Unknown event: %s at %d", event.EventType(), event.OccurredAt())
	}
}
