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
	ID           string
	Player1Conn  *websocket.Conn
	Player2Conn  *websocket.Conn
	Player1ID    string
	Player2ID    string
	CurrentRound int
	mu           sync.Mutex
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

	// Assign connection to appropriate player slot
	if game.Player1ID == "" || game.Player1ID == playerID {
		game.Player1ID = playerID
		game.Player1Conn = conn
	} else if game.Player2ID == "" || game.Player2ID == playerID {
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
	if game.Player1ID == playerID {
		game.Player1Conn = nil
		opponentConn = game.Player2Conn
	} else if game.Player2ID == playerID {
		game.Player2Conn = nil
		opponentConn = game.Player1Conn
	}

	if opponentConn != nil {
		opponentConn.WriteJSON(map[string]interface{}{
			"type": "opponent_disconnected",
			"data": map[string]interface{}{
				"playerId":    playerID,
				"reconnectIn": 30, // seconds to wait for reconnect
			},
		})
	}

	// Clean up game if both disconnected
	if game.Player1Conn == nil && game.Player2Conn == nil {
		delete(h.games, gameID)
		log.Printf("Game %s cleaned up - both players disconnected", gameID)
	}

	log.Printf("Player %s disconnected from game %s", playerID, gameID)
}

func (h *DuelWebSocketHub) notifyBothPlayersReady(game *DuelGame) {
	// Both players connected - start the game
	readyMsg := map[string]interface{}{
		"type": "game_ready",
		"data": map[string]interface{}{
			"gameId":      game.ID,
			"player1Id":   game.Player1ID,
			"player2Id":   game.Player2ID,
			"startsIn":    3, // countdown seconds
			"totalRounds": quick_duel.QuestionsPerDuel,
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(readyMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(readyMsg)
	}

	// Start the game after countdown
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

	// Send result to the player who answered
	var playerConn *websocket.Conn
	if game.Player1ID == playerID {
		playerConn = game.Player1Conn
	} else {
		playerConn = game.Player2Conn
	}

	if err != nil {
		if playerConn != nil {
			playerConn.WriteJSON(map[string]interface{}{
				"type":  "answer_error",
				"error": err.Error(),
			})
		}
		return
	}

	// Send answer result to both players
	answerResult := map[string]interface{}{
		"type": "answer_result",
		"data": map[string]interface{}{
			"playerId":      playerID,
			"questionId":    data.QuestionID,
			"isCorrect":     output.IsCorrect,
			"correctAnswer": output.CorrectAnswerID,
			"pointsEarned":  output.PointsEarned,
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

	// Check if round is complete (both players answered)
	if output.RoundComplete {
		h.handleRoundComplete(game, output)
	}

	// Check if game is finished
	if output.GameComplete {
		h.handleGameComplete(game, output)
	}
}

func (h *DuelWebSocketHub) startRound(game *DuelGame, roundNum int) {
	if h.startGameUC == nil {
		log.Printf("StartGameUseCase not initialized")
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	game.CurrentRound = roundNum

	// Get question for this round from the use case
	output, err := h.startGameUC.GetRoundQuestion(game.ID, roundNum)
	if err != nil {
		log.Printf("Failed to get question for round %d: %v", roundNum, err)
		return
	}

	// Send question to both players
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

	// Auto-advance after time limit
	go func() {
		time.Sleep(time.Duration(quick_duel.TimePerQuestionSec+2) * time.Second)
		h.checkRoundTimeout(game, roundNum)
	}()
}

func (h *DuelWebSocketHub) checkRoundTimeout(game *DuelGame, roundNum int) {
	game.mu.Lock()
	defer game.mu.Unlock()

	// If still on the same round, force advance
	if game.CurrentRound == roundNum {
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

		// Move to next round
		if roundNum < quick_duel.QuestionsPerDuel {
			go h.startRound(game, roundNum+1)
		}
	}
}

func (h *DuelWebSocketHub) handleRoundComplete(game *DuelGame, output *appDuel.SubmitDuelAnswerOutput) {
	roundCompleteMsg := map[string]interface{}{
		"type": "round_complete",
		"data": map[string]interface{}{
			"roundNum":     game.CurrentRound,
			"player1Score": output.Player1Score,
			"player2Score": output.Player2Score,
			"nextRoundIn":  2, // seconds
		},
	}

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(roundCompleteMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(roundCompleteMsg)
	}

	// Start next round after delay
	if game.CurrentRound < quick_duel.QuestionsPerDuel {
		go func() {
			time.Sleep(2 * time.Second)
			h.startRound(game, game.CurrentRound+1)
		}()
	}
}

func (h *DuelWebSocketHub) handleGameComplete(game *DuelGame, output *appDuel.SubmitDuelAnswerOutput) {
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

	if game.Player1Conn != nil {
		game.Player1Conn.WriteJSON(gameCompleteMsg)
	}
	if game.Player2Conn != nil {
		game.Player2Conn.WriteJSON(gameCompleteMsg)
	}

	// Clean up game after a delay
	go func() {
		time.Sleep(30 * time.Second)
		h.mu.Lock()
		delete(h.games, game.ID)
		h.mu.Unlock()
	}()
}

// NotifyGameCreated notifies players that a game has been created
func (h *DuelWebSocketHub) NotifyGameCreated(gameID, player1ID, player2ID string) {
	// This would be called by the matchmaking service when a game is found
	log.Printf("Game %s created: %s vs %s", gameID, player1ID, player2ID)
}
