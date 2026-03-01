package handlers

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/websocket"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// DuelWebSocketHub manages WebSocket connections for real-time duels
type DuelWebSocketHub struct {
	// Game ID -> player connections
	games map[string]*DuelGame
	mu    sync.RWMutex

	// Use cases
	startGameUC    *appDuel.StartGameUseCase
	submitAnswerUC *appDuel.SubmitDuelAnswerUseCase
}

// DuelGame represents an active duel game with two players
type DuelGame struct {
	ID              string
	Player1Conn     *websocket.Conn
	Player2Conn     *websocket.Conn
	Player1ID       string
	Player2ID       string
	CurrentRound    int
	Finished        bool
	GameCompleteMsg map[string]interface{} // cached for reconnecting players
	mu              sync.Mutex
}

// DuelMessage represents a WebSocket message
type DuelMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// PlayerReadyData for player_ready message
type PlayerReadyData struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

// DuelSubmitAnswerData for submit_answer message
type DuelSubmitAnswerData struct {
	PlayerID   string `json:"playerId"`
	GameID     string `json:"gameId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	TimeTaken  int    `json:"timeTaken"` // milliseconds
}

// NewDuelWebSocketHub creates a new duel WebSocket hub
func NewDuelWebSocketHub(
	startGameUC *appDuel.StartGameUseCase,
	submitAnswerUC *appDuel.SubmitDuelAnswerUseCase,
) *DuelWebSocketHub {
	return &DuelWebSocketHub{
		games:          make(map[string]*DuelGame),
		startGameUC:    startGameUC,
		submitAnswerUC: submitAnswerUC,
	}
}

// HandleDuelWebSocket handles WebSocket connections for duels
func (h *DuelWebSocketHub) HandleDuelWebSocket(c *websocket.Conn) {
	gameID := c.Params("gameId")
	playerID := c.Query("playerId")

	if gameID == "" || playerID == "" {
		log.Printf("Missing gameId or playerId")
		c.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Missing gameId or playerId",
		})
		c.Close()
		return
	}

	// Register player to game
	if err := h.registerPlayer(gameID, playerID, c); err != nil {
		log.Printf("Failed to register player: %v", err)
		c.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": err.Error(),
		})
		c.Close()
		return
	}

	// Clean up on disconnect
	defer h.unregisterPlayer(gameID, playerID, c)

	// Listen for messages
	for {
		_, msgBytes, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Duel WebSocket error: %v", err)
			}
			break
		}

		var msg DuelMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		h.handleMessage(gameID, playerID, c, msg)
	}
}

func (h *DuelWebSocketHub) registerPlayer(gameID, playerID string, conn *websocket.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	game, exists := h.games[gameID]
	if !exists {
		// Create new game entry
		game = &DuelGame{
			ID:           gameID,
			CurrentRound: 0,
		}
		h.games[gameID] = game
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Reconnect: existing player slot reconnects
	if game.Player1ID == playerID {
		game.Player1Conn = conn
		log.Printf("Player %s reconnected to game %s (slot 1)", playerID, gameID)
		conn.WriteJSON(map[string]interface{}{
			"type": "connected",
			"data": map[string]interface{}{
				"gameId":   gameID,
				"playerId": playerID,
			},
		})
		if game.Finished && game.GameCompleteMsg != nil {
			conn.WriteJSON(game.GameCompleteMsg)
		} else if game.CurrentRound > 0 {
			go h.resendCurrentQuestion(game, game.CurrentRound)
		}
		return nil
	}

	if game.Player2ID == playerID {
		game.Player2Conn = conn
		log.Printf("Player %s reconnected to game %s (slot 2)", playerID, gameID)
		conn.WriteJSON(map[string]interface{}{
			"type": "connected",
			"data": map[string]interface{}{
				"gameId":   gameID,
				"playerId": playerID,
			},
		})
		if game.Finished && game.GameCompleteMsg != nil {
			conn.WriteJSON(game.GameCompleteMsg)
		} else if game.CurrentRound > 0 {
			go h.resendCurrentQuestion(game, game.CurrentRound)
		}
		return nil
	}

	// New player: assign to empty slot
	if game.Player1ID == "" {
		game.Player1ID = playerID
		game.Player1Conn = conn
	} else if game.Player2ID == "" {
		game.Player2ID = playerID
		game.Player2Conn = conn
	} else {
		return quick_duel.ErrAlreadyInGame
	}

	log.Printf("Player %s connected to game %s", playerID, gameID)

	// Send connection confirmation
	conn.WriteJSON(map[string]interface{}{
		"type": "connected",
		"data": map[string]interface{}{
			"gameId":   gameID,
			"playerId": playerID,
		},
	})

	// Check if both players are connected
	if game.Player1Conn != nil && game.Player2Conn != nil {
		h.notifyBothPlayersReady(game)
	}

	return nil
}

func (h *DuelWebSocketHub) unregisterPlayer(gameID, playerID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	game, exists := h.games[gameID]
	if !exists {
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Notify opponent of disconnect
	var opponentConn *websocket.Conn
	if game.Player1ID == playerID && game.Player1Conn == conn {
		game.Player1Conn = nil
		opponentConn = game.Player2Conn
	} else if game.Player2ID == playerID && game.Player2Conn == conn {
		game.Player2Conn = nil
		opponentConn = game.Player1Conn
	}

	if opponentConn != nil {
		opponentConn.WriteJSON(map[string]interface{}{
			"type": "opponent_disconnected",
			"data": map[string]interface{}{
				"playerId":    playerID,
				"reconnectIn": 10, // seconds — per spec
			},
		})
		// Grace period: if player doesn't reconnect within 10s, forfeit remaining rounds
		go h.handleDisconnectGracePeriod(gameID, playerID)
	}

	// Clean up game if both disconnected
	if game.Player1Conn == nil && game.Player2Conn == nil {
		delete(h.games, gameID)
		log.Printf("Game %s cleaned up - both players disconnected", gameID)
	}

	log.Printf("Player %s disconnected from game %s", playerID, gameID)
}

// handleDisconnectGracePeriod waits 10 s; if the player is still gone, it submits
// timeout answers for all remaining rounds so the opponent gets a proper result.
func (h *DuelWebSocketHub) handleDisconnectGracePeriod(gameID, playerID string) {
	time.Sleep(10 * time.Second)

	h.mu.RLock()
	game, exists := h.games[gameID]
	h.mu.RUnlock()

	if !exists {
		return // Game already cleaned up
	}

	game.mu.Lock()
	reconnected := (game.Player1ID == playerID && game.Player1Conn != nil) ||
		(game.Player2ID == playerID && game.Player2Conn != nil)
	currentRound := game.CurrentRound
	game.mu.Unlock()

	if reconnected {
		return
	}

	if h.submitAnswerUC == nil {
		return
	}

	// Submit timeout answers for every remaining round
	for round := currentRound; round <= quick_duel.QuestionsPerDuel; round++ {
		output, err := h.submitAnswerUC.TimeoutRound(gameID, round)
		if err != nil {
			log.Printf("Game %s: grace period TimeoutRound(%d) error: %v", gameID, round, err)
			break
		}
		if output == nil {
			break // Round already past or game finished before us
		}
		if output.GameComplete {
			h.mu.RLock()
			g, ok := h.games[gameID]
			h.mu.RUnlock()
			if ok {
				g.mu.Lock()
				h.broadcastGameComplete(g, output)
				g.mu.Unlock()
			}
			break
		}
	}
}

func (h *DuelWebSocketHub) notifyBothPlayersReady(game *DuelGame) {
	// Use domain player order so game_ready.player1Id matches answer_result.player1Score.
	// Domain order: player1 = challenger (set at game creation), player2 = accepter.
	// Hub order (WS connection order) may differ — accepter usually connects first.
	domainPlayer1ID := game.Player1ID
	domainPlayer2ID := game.Player2ID
	if h.startGameUC != nil {
		if p1, p2, err := h.startGameUC.GetDomainPlayerOrder(game.ID); err == nil {
			domainPlayer1ID = p1
			domainPlayer2ID = p2
		} else {
			log.Printf("[DuelWS] GetDomainPlayerOrder failed for %s: %v (falling back to hub order)", game.ID, err)
		}
	}

	readyMsg := map[string]interface{}{
		"type": "game_ready",
		"data": map[string]interface{}{
			"gameId":      game.ID,
			"player1Id":   domainPlayer1ID,
			"player2Id":   domainPlayer2ID,
			"startsIn":    3,
			"totalRounds": quick_duel.QuestionsPerDuel,
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(readyMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(readyMsg)
	}

	go func() {
		time.Sleep(3 * time.Second)
		h.startRound(game, 1)
	}()
}

func (h *DuelWebSocketHub) handleMessage(gameID, playerID string, conn *websocket.Conn, msg DuelMessage) {
	switch msg.Type {
	case "submit_answer":
		var data DuelSubmitAnswerData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			conn.WriteJSON(map[string]interface{}{
				"type":  "error",
				"error": "Invalid answer data",
			})
			return
		}
		h.handleSubmitAnswer(gameID, playerID, data)

	case "player_ready":
		// Player signals ready for next round
		log.Printf("Player %s ready in game %s", playerID, gameID)

	case "ping":
		conn.WriteJSON(map[string]interface{}{"type": "pong"})

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *DuelWebSocketHub) handleSubmitAnswer(gameID, playerID string, data DuelSubmitAnswerData) {
	if h.submitAnswerUC == nil {
		log.Printf("SubmitDuelAnswerUseCase not initialized")
		return
	}

	output, err := h.submitAnswerUC.Execute(appDuel.SubmitDuelAnswerInput{
		GameID:     gameID,
		PlayerID:   playerID,
		QuestionID: data.QuestionID,
		AnswerID:   data.AnswerID,
		TimeTaken:  data.TimeTaken,
	})

	h.mu.RLock()
	game, exists := h.games[gameID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Find the connection for the player who answered (for error reporting)
	var playerConn *websocket.Conn
	if game.Player1ID == playerID {
		playerConn = game.Player1Conn
	} else {
		playerConn = game.Player2Conn
	}

	if err != nil {
		if playerConn != nil {
			playerConn.WriteJSON(map[string]interface{}{
				"type":  "error",
				"error": err.Error(),
			})
		}
		return
	}

	// Broadcast answer_result to BOTH players.
	// Per spec and commit d886947: no pointsEarned field.
	answerResult := map[string]interface{}{
		"type": "answer_result",
		"data": map[string]interface{}{
			"playerId":      playerID,
			"questionId":    data.QuestionID,
			"isCorrect":     output.IsCorrect,
			"correctAnswer": output.CorrectAnswerID,
			"timeTaken":     data.TimeTaken,
			"player1Score":  output.Player1Score,
			"player2Score":  output.Player2Score,
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(answerResult)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(answerResult)
	}

	// When game is finished, send game_complete (takes priority over round_complete).
	// round_complete is NOT sent for the final round — game_complete carries final scores.
	if output.GameComplete {
		h.broadcastGameComplete(game, output)
		return
	}

	// Round complete — both players answered but the game continues
	if output.RoundComplete {
		h.broadcastRoundComplete(game, output)
		// Schedule next question after delay (outside the lock)
		nextRound := game.CurrentRound + 1
		go func() {
			time.Sleep(2 * time.Second)
			h.startRound(game, nextRound)
		}()
	}
}

// broadcastRoundComplete sends round_complete to both players.
// Must be called with game.mu held.
func (h *DuelWebSocketHub) broadcastRoundComplete(game *DuelGame, output *appDuel.SubmitDuelAnswerOutput) {
	roundCompleteMsg := map[string]interface{}{
		"type": "round_complete",
		"data": map[string]interface{}{
			"roundNum":     game.CurrentRound,
			"player1Score": output.Player1Score,
			"player2Score": output.Player2Score,
			"nextRoundIn":  2, // seconds — per spec
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(roundCompleteMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(roundCompleteMsg)
	}
}

// broadcastGameComplete sends game_complete to both players.
// Must be called with game.mu held.
func (h *DuelWebSocketHub) broadcastGameComplete(game *DuelGame, output *appDuel.SubmitDuelAnswerOutput) {
	gameCompleteMsg := map[string]interface{}{
		"type": "game_complete",
		"data": map[string]interface{}{
			"winnerId":         output.WinnerID,
			"player1Score":     output.Player1Score,
			"player2Score":     output.Player2Score,
			"player1MMRChange": output.Player1MMRChange,
			"player2MMRChange": output.Player2MMRChange,
			"player1NewMMR":    output.Player1NewMMR,
			"player2NewMMR":    output.Player2NewMMR,
		},
	}

	// Cache for reconnecting players
	game.Finished = true
	game.GameCompleteMsg = gameCompleteMsg

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(gameCompleteMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(gameCompleteMsg)
	}

	// Schedule game clean-up — do NOT hold the lock during the sleep
	gameID := game.ID
	go func() {
		time.Sleep(30 * time.Second)
		h.mu.Lock()
		delete(h.games, gameID)
		h.mu.Unlock()
	}()
}

// startRound sends new_question to both players and arms the timeout goroutine.
// Safe to call from any goroutine — acquires game.mu internally.
func (h *DuelWebSocketHub) startRound(game *DuelGame, roundNum int) {
	if h.startGameUC == nil {
		log.Printf("StartGameUseCase not initialized")
		return
	}

	// Get question data BEFORE acquiring the lock to avoid holding it during I/O.
	output, err := h.startGameUC.GetRoundQuestion(game.ID, roundNum)
	if err != nil {
		log.Printf("Failed to get question for round %d: %v", roundNum, err)
		return
	}

	game.mu.Lock()

	game.CurrentRound = roundNum

	questionMsg := map[string]interface{}{
		"type": "new_question",
		"data": map[string]interface{}{
			"roundNum":    roundNum,
			"totalRounds": quick_duel.QuestionsPerDuel,
			"question": map[string]interface{}{
				"id":        output.QuestionID,
				"text":      output.QuestionText,
				"answers":   output.Answers,
				"timeLimit": quick_duel.TimePerQuestionSec,
			},
			"serverTime": time.Now().UnixMilli(),
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(questionMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(questionMsg)
	}

	game.mu.Unlock()

	// Arm the 10-second timeout AFTER releasing the lock
	go func() {
		time.Sleep(time.Duration(quick_duel.TimePerQuestionSec) * time.Second)
		h.checkRoundTimeout(game, roundNum)
	}()
}

// resendCurrentQuestion sends the current question to a reconnecting player.
// Must be called WITHOUT game.mu held (it calls startRound which acquires the lock).
func (h *DuelWebSocketHub) resendCurrentQuestion(game *DuelGame, roundNum int) {
	if h.startGameUC == nil {
		return
	}

	output, err := h.startGameUC.GetRoundQuestion(game.ID, roundNum)
	if err != nil {
		log.Printf("Failed to resend question for round %d: %v", roundNum, err)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Only resend if we're still on the same round
	if game.CurrentRound != roundNum {
		return
	}

	questionMsg := map[string]interface{}{
		"type": "new_question",
		"data": map[string]interface{}{
			"roundNum":    roundNum,
			"totalRounds": quick_duel.QuestionsPerDuel,
			"question": map[string]interface{}{
				"id":        output.QuestionID,
				"text":      output.QuestionText,
				"answers":   output.Answers,
				"timeLimit": quick_duel.TimePerQuestionSec,
			},
			"serverTime": time.Now().UnixMilli(),
		},
	}

	// Reconnecting player is whichever slot just regained a non-nil conn.
	// Broadcast to both: a player who already got it will simply ignore a duplicate.
	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(questionMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(questionMsg)
	}
}

// checkRoundTimeout fires when the per-question timer expires.
// If the round hasn't advanced yet (CurrentRound == roundNum), it records
// a timeout for any player who hasn't answered and broadcasts round_timeout.
func (h *DuelWebSocketHub) checkRoundTimeout(game *DuelGame, roundNum int) {
	game.mu.Lock()

	// Guard: round already advanced (both answered in time) or game finished
	if game.CurrentRound != roundNum {
		game.mu.Unlock()
		return
	}

	// Notify players of timeout
	timeoutMsg := map[string]interface{}{
		"type": "round_timeout",
		"data": map[string]interface{}{
			"roundNum": roundNum,
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(timeoutMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(timeoutMsg)
	}

	isLastRound := roundNum >= quick_duel.QuestionsPerDuel
	nextRound := roundNum + 1
	gameID := game.ID

	game.mu.Unlock()

	// Record timeout answers for any players who haven't answered yet.
	// This also advances game.currentRound in the domain so future answers
	// are validated against the correct question.
	var timeoutOutput *appDuel.SubmitDuelAnswerOutput
	if h.submitAnswerUC != nil {
		out, err := h.submitAnswerUC.TimeoutRound(gameID, roundNum)
		if err != nil {
			log.Printf("Game %s: TimeoutRound(%d) error: %v", gameID, roundNum, err)
		}
		timeoutOutput = out
	}

	if timeoutOutput != nil && timeoutOutput.GameComplete {
		h.mu.RLock()
		g, ok := h.games[gameID]
		h.mu.RUnlock()
		if ok {
			g.mu.Lock()
			h.broadcastGameComplete(g, timeoutOutput)
			g.mu.Unlock()
		}
		return
	}

	if !isLastRound {
		go func() {
			time.Sleep(2 * time.Second)
			h.startRound(game, nextRound)
		}()
	}
}

// NotifyGameCreated notifies players that a game has been created
func (h *DuelWebSocketHub) NotifyGameCreated(gameID, player1ID, player2ID string) {
	// This would be called by the matchmaking service when a game is found
	log.Printf("Game %s created: %s vs %s", gameID, player1ID, player2ID)
}

// handleRoundComplete is kept for backward compatibility but is no longer called
// directly — logic moved inline into handleSubmitAnswer.
func (h *DuelWebSocketHub) handleRoundComplete(game *DuelGame, output *appDuel.SubmitDuelAnswerOutput) {
	h.broadcastRoundComplete(game, output)
}

// handleGameComplete is kept for backward compatibility but is no longer called
// directly — logic moved inline into handleSubmitAnswer.
func (h *DuelWebSocketHub) handleGameComplete(game *DuelGame, output *appDuel.SubmitDuelAnswerOutput) {
	h.broadcastGameComplete(game, output)
}
