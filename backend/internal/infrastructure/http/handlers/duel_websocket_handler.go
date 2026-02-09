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
	// Match ID -> player connections
	matches map[string]*DuelMatch
	mu      sync.RWMutex

	// Use cases
	startMatchUC  *appDuel.StartMatchUseCase
	submitAnswerUC *appDuel.SubmitDuelAnswerUseCase
}

// DuelMatch represents an active duel match with two players
type DuelMatch struct {
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
	MatchID  string `json:"matchId"`
}

// DuelSubmitAnswerData for submit_answer message
type DuelSubmitAnswerData struct {
	PlayerID   string `json:"playerId"`
	MatchID    string `json:"matchId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	TimeTaken  int    `json:"timeTaken"` // milliseconds
}

// NewDuelWebSocketHub creates a new duel WebSocket hub
func NewDuelWebSocketHub(
	startMatchUC *appDuel.StartMatchUseCase,
	submitAnswerUC *appDuel.SubmitDuelAnswerUseCase,
) *DuelWebSocketHub {
	return &DuelWebSocketHub{
		matches:        make(map[string]*DuelMatch),
		startMatchUC:   startMatchUC,
		submitAnswerUC: submitAnswerUC,
	}
}

// HandleDuelWebSocket handles WebSocket connections for duels
func (h *DuelWebSocketHub) HandleDuelWebSocket(c *websocket.Conn) {
	matchID := c.Params("matchId")
	playerID := c.Query("playerId")

	if matchID == "" || playerID == "" {
		log.Printf("Missing matchId or playerId")
		c.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Missing matchId or playerId",
		})
		c.Close()
		return
	}

	// Register player to match
	if err := h.registerPlayer(matchID, playerID, c); err != nil {
		log.Printf("Failed to register player: %v", err)
		c.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": err.Error(),
		})
		c.Close()
		return
	}

	// Clean up on disconnect
	defer h.unregisterPlayer(matchID, playerID, c)

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

		h.handleMessage(matchID, playerID, c, msg)
	}
}

func (h *DuelWebSocketHub) registerPlayer(matchID, playerID string, conn *websocket.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	match, exists := h.matches[matchID]
	if !exists {
		// Create new match entry
		match = &DuelMatch{
			ID:           matchID,
			CurrentRound: 0,
		}
		h.matches[matchID] = match
	}

	match.mu.Lock()
	defer match.mu.Unlock()

	// Assign connection to appropriate player slot
	if match.Player1ID == "" || match.Player1ID == playerID {
		match.Player1ID = playerID
		match.Player1Conn = conn
	} else if match.Player2ID == "" || match.Player2ID == playerID {
		match.Player2ID = playerID
		match.Player2Conn = conn
	} else {
		return quick_duel.ErrAlreadyInMatch
	}

	log.Printf("Player %s connected to match %s", playerID, matchID)

	// Send connection confirmation
	conn.WriteJSON(map[string]interface{}{
		"type": "connected",
		"data": map[string]interface{}{
			"matchId":  matchID,
			"playerId": playerID,
		},
	})

	// Check if both players are connected
	if match.Player1Conn != nil && match.Player2Conn != nil {
		h.notifyBothPlayersReady(match)
	}

	return nil
}

func (h *DuelWebSocketHub) unregisterPlayer(matchID, playerID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	match, exists := h.matches[matchID]
	if !exists {
		return
	}

	match.mu.Lock()
	defer match.mu.Unlock()

	// Notify opponent of disconnect
	var opponentConn *websocket.Conn
	if match.Player1ID == playerID {
		match.Player1Conn = nil
		opponentConn = match.Player2Conn
	} else if match.Player2ID == playerID {
		match.Player2Conn = nil
		opponentConn = match.Player1Conn
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

	// Clean up match if both disconnected
	if match.Player1Conn == nil && match.Player2Conn == nil {
		delete(h.matches, matchID)
		log.Printf("Match %s cleaned up - both players disconnected", matchID)
	}

	log.Printf("Player %s disconnected from match %s", playerID, matchID)
}

func (h *DuelWebSocketHub) notifyBothPlayersReady(match *DuelMatch) {
	// Both players connected - start the match
	readyMsg := map[string]interface{}{
		"type": "match_ready",
		"data": map[string]interface{}{
			"matchId":     match.ID,
			"player1Id":   match.Player1ID,
			"player2Id":   match.Player2ID,
			"startsIn":    3, // countdown seconds
			"totalRounds": quick_duel.QuestionsPerDuel,
		},
	}

	if match.Player1Conn != nil {
		match.Player1Conn.WriteJSON(readyMsg)
	}
	if match.Player2Conn != nil {
		match.Player2Conn.WriteJSON(readyMsg)
	}

	// Start the match after countdown
	go func() {
		time.Sleep(3 * time.Second)
		h.startRound(match, 1)
	}()
}

func (h *DuelWebSocketHub) handleMessage(matchID, playerID string, conn *websocket.Conn, msg DuelMessage) {
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
		h.handleSubmitAnswer(matchID, playerID, data)

	case "player_ready":
		// Player signals ready for next round
		log.Printf("Player %s ready in match %s", playerID, matchID)

	case "ping":
		conn.WriteJSON(map[string]interface{}{"type": "pong"})

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *DuelWebSocketHub) handleSubmitAnswer(matchID, playerID string, data DuelSubmitAnswerData) {
	if h.submitAnswerUC == nil {
		log.Printf("SubmitDuelAnswerUseCase not initialized")
		return
	}

	output, err := h.submitAnswerUC.Execute(appDuel.SubmitDuelAnswerInput{
		MatchID:    matchID,
		PlayerID:   playerID,
		QuestionID: data.QuestionID,
		AnswerID:   data.AnswerID,
		TimeTaken:  data.TimeTaken,
	})

	h.mu.RLock()
	match, exists := h.matches[matchID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	match.mu.Lock()
	defer match.mu.Unlock()

	// Send result to the player who answered
	var playerConn *websocket.Conn
	if match.Player1ID == playerID {
		playerConn = match.Player1Conn
	} else {
		playerConn = match.Player2Conn
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

	if match.Player1Conn != nil {
		match.Player1Conn.WriteJSON(answerResult)
	}
	if match.Player2Conn != nil {
		match.Player2Conn.WriteJSON(answerResult)
	}

	// Check if round is complete (both players answered)
	if output.RoundComplete {
		h.handleRoundComplete(match, output)
	}

	// Check if match is finished
	if output.MatchComplete {
		h.handleMatchComplete(match, output)
	}
}

func (h *DuelWebSocketHub) startRound(match *DuelMatch, roundNum int) {
	if h.startMatchUC == nil {
		log.Printf("StartMatchUseCase not initialized")
		return
	}

	match.mu.Lock()
	defer match.mu.Unlock()

	match.CurrentRound = roundNum

	// Get question for this round from the use case
	output, err := h.startMatchUC.GetRoundQuestion(match.ID, roundNum)
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

	if match.Player1Conn != nil {
		match.Player1Conn.WriteJSON(questionMsg)
	}
	if match.Player2Conn != nil {
		match.Player2Conn.WriteJSON(questionMsg)
	}

	// Auto-advance after time limit
	go func() {
		time.Sleep(time.Duration(quick_duel.TimePerQuestionSec+2) * time.Second)
		h.checkRoundTimeout(match, roundNum)
	}()
}

func (h *DuelWebSocketHub) checkRoundTimeout(match *DuelMatch, roundNum int) {
	match.mu.Lock()
	defer match.mu.Unlock()

	// If still on the same round, force advance
	if match.CurrentRound == roundNum {
		// Notify players of timeout
		timeoutMsg := map[string]interface{}{
			"type": "round_timeout",
			"data": map[string]interface{}{
				"roundNum": roundNum,
			},
		}

		if match.Player1Conn != nil {
			match.Player1Conn.WriteJSON(timeoutMsg)
		}
		if match.Player2Conn != nil {
			match.Player2Conn.WriteJSON(timeoutMsg)
		}

		// Move to next round
		if roundNum < quick_duel.QuestionsPerDuel {
			go h.startRound(match, roundNum+1)
		}
	}
}

func (h *DuelWebSocketHub) handleRoundComplete(match *DuelMatch, output *appDuel.SubmitDuelAnswerOutput) {
	roundCompleteMsg := map[string]interface{}{
		"type": "round_complete",
		"data": map[string]interface{}{
			"roundNum":     match.CurrentRound,
			"player1Score": output.Player1Score,
			"player2Score": output.Player2Score,
			"nextRoundIn":  2, // seconds
		},
	}

	if match.Player1Conn != nil {
		match.Player1Conn.WriteJSON(roundCompleteMsg)
	}
	if match.Player2Conn != nil {
		match.Player2Conn.WriteJSON(roundCompleteMsg)
	}

	// Start next round after delay
	if match.CurrentRound < quick_duel.QuestionsPerDuel {
		go func() {
			time.Sleep(2 * time.Second)
			h.startRound(match, match.CurrentRound+1)
		}()
	}
}

func (h *DuelWebSocketHub) handleMatchComplete(match *DuelMatch, output *appDuel.SubmitDuelAnswerOutput) {
	matchCompleteMsg := map[string]interface{}{
		"type": "match_complete",
		"data": map[string]interface{}{
			"winnerId":        output.WinnerID,
			"player1Score":    output.Player1Score,
			"player2Score":    output.Player2Score,
			"player1MMRChange": output.Player1MMRChange,
			"player2MMRChange": output.Player2MMRChange,
			"player1NewMMR":   output.Player1NewMMR,
			"player2NewMMR":   output.Player2NewMMR,
		},
	}

	if match.Player1Conn != nil {
		match.Player1Conn.WriteJSON(matchCompleteMsg)
	}
	if match.Player2Conn != nil {
		match.Player2Conn.WriteJSON(matchCompleteMsg)
	}

	// Clean up match after a delay
	go func() {
		time.Sleep(30 * time.Second)
		h.mu.Lock()
		delete(h.matches, match.ID)
		h.mu.Unlock()
	}()
}

// NotifyMatchCreated notifies players that a match has been created
func (h *DuelWebSocketHub) NotifyMatchCreated(matchID, player1ID, player2ID string) {
	// This would be called by the matchmaking service when a match is found
	log.Printf("Match %s created: %s vs %s", matchID, player1ID, player2ID)
}
