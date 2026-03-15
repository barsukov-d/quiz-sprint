package quick_duel

import (
	"log"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

const (
	// BotUserID is the sentinel user ID for the bot opponent.
	// Must not collide with real Telegram user IDs (Telegram IDs are always > 0).
	BotUserID    = "0"
	BotUsername  = "QuizBot"
	// QueueTimeoutSec is how long a player can wait before getting a bot opponent.
	QueueTimeoutSec = 60
)

// BotFallbackUseCase pairs long-waiting queue players with a bot opponent.
// It is driven by an external scheduler (e.g. a goroutine ticking every 5s).
type BotFallbackUseCase struct {
	matchmakingQueue quick_duel.MatchmakingQueue
	duelGameRepo     quick_duel.DuelGameRepository
	playerRatingRepo quick_duel.PlayerRatingRepository
	questionRepo     QuestionRepository
	seasonRepo       quick_duel.SeasonRepository
	eventBus         EventBus
}

func NewBotFallbackUseCase(
	matchmakingQueue quick_duel.MatchmakingQueue,
	duelGameRepo quick_duel.DuelGameRepository,
	playerRatingRepo quick_duel.PlayerRatingRepository,
	questionRepo QuestionRepository,
	seasonRepo quick_duel.SeasonRepository,
	eventBus EventBus,
) *BotFallbackUseCase {
	return &BotFallbackUseCase{
		matchmakingQueue: matchmakingQueue,
		duelGameRepo:     duelGameRepo,
		playerRatingRepo: playerRatingRepo,
		questionRepo:     questionRepo,
		seasonRepo:       seasonRepo,
		eventBus:         eventBus,
	}
}

// Execute scans the queue for players waiting more than QueueTimeoutSec seconds
// and creates a bot game for each of them.
func (uc *BotFallbackUseCase) Execute() {
	now := time.Now().UTC().Unix()
	cutoff := now - QueueTimeoutSec

	stale, err := uc.matchmakingQueue.GetStaleQueueEntries(cutoff)
	if err != nil {
		log.Printf("[BotFallback] Failed to get stale queue entries: %v", err)
		return
	}

	for _, playerID := range stale {
		if err := uc.spawnBotGame(playerID, now); err != nil {
			log.Printf("[BotFallback] Failed to spawn bot game for player %s: %v", playerID.String(), err)
			continue
		}
		// Remove from queue AFTER successfully creating the game
		if err := uc.matchmakingQueue.RemoveFromQueue(playerID); err != nil {
			log.Printf("[BotFallback] Failed to remove player %s from queue: %v", playerID.String(), err)
		}
	}
}

func (uc *BotFallbackUseCase) spawnBotGame(playerID quick_duel.UserID, now int64) error {
	// Get season
	seasonID, _ := uc.seasonRepo.GetCurrentSeason()

	// Get player rating
	rating, err := uc.playerRatingRepo.FindOrCreate(playerID, seasonID, now)
	if err != nil {
		return err
	}

	// Bot player: use player's own MMR so the ELO change is symmetric
	botID, err := shared.NewUserID(BotUserID)
	if err != nil {
		return err
	}

	player := quick_duel.NewDuelPlayer(
		playerID,
		"", // username looked up in GetGameResult; empty is fine here
		quick_duel.ReconstructEloRating(rating.MMR(), 0),
	)
	bot := quick_duel.NewDuelPlayer(
		botID,
		BotUsername,
		quick_duel.ReconstructEloRating(rating.MMR(), 0), // match player's MMR for fair ELO
	)

	// Get questions
	questions, err := uc.questionRepo.FindRandomByDifficulty(quick_duel.QuestionsPerDuel, "medium")
	if err != nil {
		return err
	}

	questionIDs := make([]quick_duel.QuestionID, 0, len(questions))
	for _, q := range questions {
		qid, err := quiz.NewQuestionIDFromString(q.ID)
		if err != nil {
			return err
		}
		questionIDs = append(questionIDs, qid)
	}

	// Create and start the game
	game, err := quick_duel.NewDuelGame(player, bot, questionIDs, now)
	if err != nil {
		return err
	}
	if err := game.Start(now); err != nil {
		return err
	}

	if err := uc.duelGameRepo.Save(game); err != nil {
		return err
	}

	// Publish events (game_created + game_started)
	for _, event := range game.Events() {
		uc.eventBus.Publish(event)
	}

	log.Printf("[BotFallback] Created bot game %s for player %s (waited >%ds)",
		game.ID().String(), playerID.String(), QueueTimeoutSec)
	return nil
}
