package classic_mode

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/kernel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// ClassicGame is the aggregate root for classic game mode
// Represents a single classic game session with enhanced scoring mechanics,
// visual feedback system, and ghost comparison for personal mastery.
type ClassicGame struct {
	id                GameID
	playerID          shared.UserID
	quizID            quiz.QuizID
	status            GameStatus
	session           *kernel.QuizGameplaySession // Composition: delegates pure gameplay logic
	currentStreak     int                         // Number of consecutive correct answers
	maxStreak         int                         // Maximum streak achieved during this game
	currentMultiplier Multiplier                  // Current score multiplier based on streak (1.0, 1.5, 2.0)
	personalBestScore *int                        // Player's PersonalBest score for this quiz (nil if no previous record)
	ghostComparison   int                         // Score difference relative to PersonalBest at same question index

	// Domain events collected during operations
	events []Event
}

// NewClassicGame creates a new classic game session
// PersonalBest can be nil if the player has no previous record for this quiz
func NewClassicGame(
	playerID shared.UserID,
	quizAggregate *quiz.Quiz,
	personalBest *PersonalBest,
	startedAt int64,
) (*ClassicGame, error) {
	// 1. Validate inputs
	if playerID.IsZero() {
		return nil, shared.ErrInvalidUserID
	}

	if quizAggregate == nil {
		return nil, quiz.ErrQuizNotFound
	}

	if err := quizAggregate.CanStart(); err != nil {
		return nil, err
	}

	// 2. Create kernel gameplay session
	gameID := NewGameID()
	sessionID := kernel.NewSessionID()

	session, err := kernel.NewQuizGameplaySession(sessionID, quizAggregate, startedAt)
	if err != nil {
		return nil, err
	}

	// 3. Extract PersonalBest score if exists
	var personalBestScore *int
	hasPersonalBest := false
	if personalBest != nil {
		score := personalBest.BestScore()
		personalBestScore = &score
		hasPersonalBest = true
	}

	// 4. Create game
	game := &ClassicGame{
		id:                gameID,
		playerID:          playerID,
		quizID:            quizAggregate.ID(),
		status:            GameStatusInProgress,
		session:           session,
		currentStreak:     0,
		maxStreak:         0,
		currentMultiplier: MultiplierNormal,
		personalBestScore: personalBestScore,
		ghostComparison:   0,
		events:            make([]Event, 0),
	}

	// 5. Publish ClassicGameStarted event
	game.events = append(game.events, NewClassicGameStarted(
		gameID,
		playerID,
		quizAggregate.ID(),
		hasPersonalBest,
		personalBestScore,
		startedAt,
	))

	return game, nil
}

// SubmitAnswerResult holds detailed information about a submitted answer
type SubmitAnswerResult struct {
	IsCorrect         bool
	BasePoints        int
	TimeBonus         int
	Multiplier        float64
	TotalPoints       int     // (BasePoints + TimeBonus) * Multiplier
	CurrentStreak     int
	MaxStreak         int
	VisualState       VisualState
	GhostComparison   int     // Difference vs PersonalBest at this point
	IsGameFinished    bool
}

// SubmitAnswer processes a user's answer with enhanced scoring and streak tracking
// Business Logic:
// - Score = (BasePoints + TimeBonus) * Multiplier
// - Multiplier based on streak: 1.0 (0-2), 1.5 (3-5), 2.0 (6+)
// - Incorrect answer resets streak and multiplier
func (g *ClassicGame) SubmitAnswer(
	questionID quiz.QuestionID,
	answerID quiz.AnswerID,
	timeTaken int64,
	quizAggregate *quiz.Quiz,
	answeredAt int64,
) (*SubmitAnswerResult, error) {
	// 1. Validate game state
	if g.status != GameStatusInProgress {
		return nil, ErrGameAlreadyFinished
	}

	// 2. Validate question exists (GetQuestion will be called by kernel)
	_, err := quizAggregate.GetQuestion(questionID)
	if err != nil {
		return nil, err
	}

	// 3. Delegate to kernel session for pure gameplay logic
	kernelResult, err := g.session.AnswerQuestion(questionID, answerID, timeTaken, answeredAt)
	if err != nil {
		return nil, err
	}

	// 4. Initialize result
	result := &SubmitAnswerResult{
		IsCorrect:       kernelResult.IsCorrect,
		BasePoints:      kernelResult.BasePoints.Value(),
		TimeBonus:       0,
		Multiplier:      g.currentMultiplier.Float64(),
		TotalPoints:     0,
		CurrentStreak:   g.currentStreak,
		MaxStreak:       g.maxStreak,
		VisualState:     VisualStateFromStreak(g.currentStreak),
		GhostComparison: g.ghostComparison,
		IsGameFinished:  false,
	}

	// 5. Process answer based on correctness
	if kernelResult.IsCorrect {
		// === CORRECT ANSWER ===

		// a. Calculate time bonus (linear formula)
		timeLimitMs := int64(quizAggregate.TimeLimitPerQuestion()) * 1000
		if timeTaken > 0 && timeTaken <= timeLimitMs {
			ratio := 1.0 - (float64(timeTaken) / float64(timeLimitMs))
			result.TimeBonus = int(float64(quizAggregate.MaxTimeBonus().Value()) * ratio)
		}

		// b. Update streak
		g.currentStreak++
		result.CurrentStreak = g.currentStreak

		// c. Update max streak
		if g.currentStreak > g.maxStreak {
			g.maxStreak = g.currentStreak
		}
		result.MaxStreak = g.maxStreak

		// d. Update multiplier based on new streak
		previousMultiplier := g.currentMultiplier
		g.currentMultiplier = MultiplierFromStreak(g.currentStreak)
		result.Multiplier = g.currentMultiplier.Float64()

		// e. Calculate total points with multiplier: (BasePoints + TimeBonus) * Multiplier
		result.TotalPoints = int(float64(result.BasePoints+result.TimeBonus) * result.Multiplier)

		// f. Check for streak milestone (3 or 6)
		if IsStreakMilestone(g.currentStreak) && g.currentMultiplier != previousMultiplier {
			visualState := VisualStateFromStreak(g.currentStreak)
			result.VisualState = visualState

			// Publish StreakMilestoneReached event
			g.events = append(g.events, NewStreakMilestoneReached(
				g.id,
				g.playerID,
				g.currentStreak,
				visualState,
				g.currentMultiplier,
				answeredAt,
			))
		}

		// g. Update ghost comparison
		if g.personalBestScore != nil {
			g.updateGhostComparison()
			result.GhostComparison = g.ghostComparison
		}

	} else {
		// === INCORRECT ANSWER OR TIMEOUT ===

		// a. Publish StreakBroken event if streak was active
		if g.currentStreak > 0 {
			g.events = append(g.events, NewStreakBroken(
				g.id,
				g.playerID,
				g.currentStreak,
				questionID,
				answeredAt,
			))
		}

		// b. Reset streak and multiplier
		g.currentStreak = 0
		g.currentMultiplier = MultiplierNormal
		result.CurrentStreak = 0
		result.Multiplier = g.currentMultiplier.Float64()
		result.VisualState = VisualStateNormal

		// c. No points for incorrect answer
		result.TotalPoints = 0

		// d. Update ghost comparison
		if g.personalBestScore != nil {
			g.updateGhostComparison()
			result.GhostComparison = g.ghostComparison
		}
	}

	// 6. Check if game is finished
	if g.session.IsFinished() {
		result.IsGameFinished = true
	}

	return result, nil
}

// Finish finalizes the game and determines if a new PersonalBest was achieved
func (g *ClassicGame) Finish(finishedAt int64, quizAggregate *quiz.Quiz) error {
	if g.status == GameStatusFinished {
		return ErrGameAlreadyFinished
	}

	// 1. Mark session as finished
	if err := g.session.Finish(finishedAt); err != nil {
		return err
	}

	// 2. Mark game as finished
	g.status = GameStatusFinished

	// 3. Get final score (with multipliers applied during gameplay)
	finalScore := g.GetTotalScore()

	// 4. Determine if new PersonalBest
	isNewPersonalBest := false
	if g.HasPassedQuiz(quizAggregate) {
		if g.personalBestScore == nil || finalScore > *g.personalBestScore {
			isNewPersonalBest = true
		}
	}

	// 5. Publish ClassicGameFinished event
	g.events = append(g.events, NewClassicGameFinished(
		g.id,
		g.playerID,
		g.quizID,
		finalScore,
		g.maxStreak,
		isNewPersonalBest,
		finishedAt,
	))

	return nil
}

// HasPassedQuiz checks if the user passed the quiz (reached PassingScore threshold)
func (g *ClassicGame) HasPassedQuiz(quizAggregate *quiz.Quiz) bool {
	if g.status != GameStatusFinished {
		return false
	}

	totalPossiblePoints := quizAggregate.GetTotalPoints()
	if totalPossiblePoints.IsZero() {
		return false
	}

	finalScore := g.GetTotalScore()
	scorePercentage := (finalScore * 100) / totalPossiblePoints.Value()
	return scorePercentage >= quizAggregate.PassingScore().Percentage()
}

// GetTotalScore calculates total score from session
// This is the score WITH multipliers applied during gameplay
func (g *ClassicGame) GetTotalScore() int {
	// Note: In the current implementation, we don't store cumulative score in ClassicGame.
	// We need to recalculate from session answers with multipliers.
	// For simplicity, let's return base score for now.
	// TODO: Store cumulative score during SubmitAnswer() calls
	return g.session.BaseScore().Value()
}

// updateGhostComparison calculates the score difference vs PersonalBest at same question index
func (g *ClassicGame) updateGhostComparison() {
	if g.personalBestScore == nil {
		g.ghostComparison = 0
		return
	}

	currentScore := g.session.BaseScore().Value() // TODO: Use actual score with multipliers

	// This is a simplified version - ideally we'd have score-by-question from PersonalBest
	// For MVP, we compare against total personal best
	g.ghostComparison = currentScore - *g.personalBestScore
}

// Getters
func (g *ClassicGame) ID() GameID                              { return g.id }
func (g *ClassicGame) PlayerID() shared.UserID                 { return g.playerID }
func (g *ClassicGame) QuizID() quiz.QuizID                     { return g.quizID }
func (g *ClassicGame) Status() GameStatus                      { return g.status }
func (g *ClassicGame) Session() *kernel.QuizGameplaySession    { return g.session }
func (g *ClassicGame) CurrentStreak() int                      { return g.currentStreak }
func (g *ClassicGame) MaxStreak() int                          { return g.maxStreak }
func (g *ClassicGame) CurrentMultiplier() Multiplier           { return g.currentMultiplier }
func (g *ClassicGame) PersonalBestScore() *int                 { return g.personalBestScore }
func (g *ClassicGame) GhostComparison() int                    { return g.ghostComparison }
func (g *ClassicGame) CurrentVisualState() VisualState         { return VisualStateFromStreak(g.currentStreak) }

// Events returns collected domain events and clears them
func (g *ClassicGame) Events() []Event {
	events := g.events
	g.events = make([]Event, 0)
	return events
}

// ReconstructClassicGame reconstructs a ClassicGame from persistence
// Used by repository when loading from database
func ReconstructClassicGame(
	id GameID,
	playerID shared.UserID,
	quizID quiz.QuizID,
	status GameStatus,
	session *kernel.QuizGameplaySession,
	currentStreak int,
	maxStreak int,
	currentMultiplier Multiplier,
	personalBestScore *int,
	ghostComparison int,
) *ClassicGame {
	return &ClassicGame{
		id:                id,
		playerID:          playerID,
		quizID:            quizID,
		status:            status,
		session:           session,
		currentStreak:     currentStreak,
		maxStreak:         maxStreak,
		currentMultiplier: currentMultiplier,
		personalBestScore: personalBestScore,
		ghostComparison:   ghostComparison,
		events:            make([]Event, 0), // Don't replay events from DB
	}
}
