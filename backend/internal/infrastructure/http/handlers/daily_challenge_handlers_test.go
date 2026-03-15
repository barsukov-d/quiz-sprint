package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	appDaily "github.com/barsukov/quiz-sprint/backend/internal/application/daily_challenge"
	domainDaily "github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	domainQuiz "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// Test Infrastructure: Mock Repositories
// ========================================

// mockDailyQuizRepo is an in-memory DailyQuizRepository for handler tests
type mockDailyQuizRepo struct {
	quizzes map[string]*domainDaily.DailyQuiz
	byDate  map[string]*domainDaily.DailyQuiz
}

func newMockDailyQuizRepo() *mockDailyQuizRepo {
	return &mockDailyQuizRepo{
		quizzes: make(map[string]*domainDaily.DailyQuiz),
		byDate:  make(map[string]*domainDaily.DailyQuiz),
	}
}

func (m *mockDailyQuizRepo) Save(dq *domainDaily.DailyQuiz) error {
	m.quizzes[dq.ID().String()] = dq
	m.byDate[dq.Date().String()] = dq
	return nil
}

func (m *mockDailyQuizRepo) FindByID(id domainDaily.DailyQuizID) (*domainDaily.DailyQuiz, error) {
	if q, ok := m.quizzes[id.String()]; ok {
		return q, nil
	}
	return nil, domainDaily.ErrDailyQuizNotFound
}

func (m *mockDailyQuizRepo) FindByDate(date domainDaily.Date) (*domainDaily.DailyQuiz, error) {
	if q, ok := m.byDate[date.String()]; ok {
		return q, nil
	}
	return nil, domainDaily.ErrDailyQuizNotFound
}

func (m *mockDailyQuizRepo) Delete(id domainDaily.DailyQuizID) error {
	if q, ok := m.quizzes[id.String()]; ok {
		delete(m.byDate, q.Date().String())
		delete(m.quizzes, id.String())
	}
	return nil
}

// mockDailyGameRepo is an in-memory DailyGameRepository for handler tests
type mockDailyGameRepo struct {
	games map[string]*domainDaily.DailyGame
}

func newMockDailyGameRepo() *mockDailyGameRepo {
	return &mockDailyGameRepo{games: make(map[string]*domainDaily.DailyGame)}
}

func (m *mockDailyGameRepo) Save(game *domainDaily.DailyGame) error {
	m.games[game.ID().String()] = game
	return nil
}

func (m *mockDailyGameRepo) FindByID(id domainDaily.GameID) (*domainDaily.DailyGame, error) {
	if g, ok := m.games[id.String()]; ok {
		return g, nil
	}
	return nil, domainDaily.ErrGameNotFound
}

func (m *mockDailyGameRepo) FindByPlayerAndDate(playerID domainDaily.UserID, date domainDaily.Date) (*domainDaily.DailyGame, error) {
	var best *domainDaily.DailyGame
	for _, g := range m.games {
		if g.PlayerID() == playerID && g.Date().Equals(date) {
			if best == nil || g.GetFinalScore() > best.GetFinalScore() {
				best = g
			}
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, domainDaily.ErrGameNotFound
}

func (m *mockDailyGameRepo) FindAllAttemptsByPlayerAndDate(playerID domainDaily.UserID, date domainDaily.Date) ([]*domainDaily.DailyGame, error) {
	var result []*domainDaily.DailyGame
	for _, g := range m.games {
		if g.PlayerID() == playerID && g.Date().Equals(date) {
			result = append(result, g)
		}
	}
	return result, nil
}

func (m *mockDailyGameRepo) CountAttemptsByPlayerAndDate(playerID domainDaily.UserID, date domainDaily.Date) (int, error) {
	count := 0
	for _, g := range m.games {
		if g.PlayerID() == playerID && g.Date().Equals(date) {
			count++
		}
	}
	return count, nil
}

func (m *mockDailyGameRepo) FindTopByDate(date domainDaily.Date, limit int) ([]*domainDaily.DailyGame, error) {
	var games []*domainDaily.DailyGame
	for _, g := range m.games {
		if g.Date().Equals(date) && g.IsCompleted() {
			games = append(games, g)
		}
	}
	for i := 0; i < len(games); i++ {
		for j := i + 1; j < len(games); j++ {
			if games[j].GetFinalScore() > games[i].GetFinalScore() {
				games[i], games[j] = games[j], games[i]
			}
		}
	}
	if limit > 0 && len(games) > limit {
		games = games[:limit]
	}
	return games, nil
}

func (m *mockDailyGameRepo) FindTopByDateAndFriends(date domainDaily.Date, playerID domainDaily.UserID, limit int) ([]*domainDaily.DailyGame, error) {
	return m.FindTopByDate(date, limit)
}

func (m *mockDailyGameRepo) FindTopByDateAndCountry(date domainDaily.Date, playerID domainDaily.UserID, limit int) ([]*domainDaily.DailyGame, error) {
	return m.FindTopByDate(date, limit)
}

func (m *mockDailyGameRepo) GetPlayerRankByDate(playerID domainDaily.UserID, date domainDaily.Date) (int, error) {
	top, _ := m.FindTopByDate(date, 0)
	for i, g := range top {
		if g.PlayerID() == playerID {
			return i + 1, nil
		}
	}
	return 0, nil
}

func (m *mockDailyGameRepo) GetTotalPlayersByDate(date domainDaily.Date) (int, error) {
	seen := make(map[string]bool)
	for _, g := range m.games {
		if g.Date().Equals(date) && g.IsCompleted() {
			seen[g.PlayerID().String()] = true
		}
	}
	return len(seen), nil
}

func (m *mockDailyGameRepo) Delete(id domainDaily.GameID) error {
	delete(m.games, id.String())
	return nil
}

func (m *mockDailyGameRepo) MarkAbandonedGames() (int, error) {
	return 0, nil
}

// mockQuestionRepo is an in-memory QuestionRepository
type mockQuestionRepo struct {
	questions map[string]*domainQuiz.Question
}

func newMockQuestionRepo() *mockQuestionRepo {
	return &mockQuestionRepo{questions: make(map[string]*domainQuiz.Question)}
}

func (m *mockQuestionRepo) FindByID(id domainQuiz.QuestionID) (*domainQuiz.Question, error) {
	if q, ok := m.questions[id.String()]; ok {
		return q, nil
	}
	return nil, domainQuiz.ErrQuestionNotFound
}

func (m *mockQuestionRepo) FindByIDs(ids []domainQuiz.QuestionID) ([]*domainQuiz.Question, error) {
	result := make([]*domainQuiz.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := m.questions[id.String()]; ok {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *mockQuestionRepo) FindByFilter(_ domainQuiz.QuestionFilter) ([]*domainQuiz.Question, error) {
	var result []*domainQuiz.Question
	for _, q := range m.questions {
		result = append(result, q)
	}
	return result, nil
}

func (m *mockQuestionRepo) FindRandomQuestions(_ domainQuiz.QuestionFilter, limit int) ([]*domainQuiz.Question, error) {
	var result []*domainQuiz.Question
	for _, q := range m.questions {
		result = append(result, q)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockQuestionRepo) FindQuestionsBySeed(f domainQuiz.QuestionFilter, limit int, _ int64) ([]*domainQuiz.Question, error) {
	return m.FindRandomQuestions(f, limit)
}

func (m *mockQuestionRepo) FindQuestionsByQuizSeed(n int, _ int64, _ *domainQuiz.CategoryID) ([]*domainQuiz.Question, error) {
	var result []*domainQuiz.Question
	for _, q := range m.questions {
		result = append(result, q)
		if len(result) >= n {
			break
		}
	}
	if len(result) < n {
		return nil, fmt.Errorf("not enough questions")
	}
	return result, nil
}

func (m *mockQuestionRepo) CountByFilter(_ domainQuiz.QuestionFilter) (int, error) { return len(m.questions), nil }
func (m *mockQuestionRepo) Save(q *domainQuiz.Question) error                      { m.questions[q.ID().String()] = q; return nil }
func (m *mockQuestionRepo) SaveAll(qs []*domainQuiz.Question) error {
	for _, q := range qs {
		m.questions[q.ID().String()] = q
	}
	return nil
}
func (m *mockQuestionRepo) Delete(id domainQuiz.QuestionID) error { delete(m.questions, id.String()); return nil }

// mockQuizRepo is an in-memory QuizRepository
type mockQuizRepo struct{}

func (m *mockQuizRepo) FindByID(_ domainQuiz.QuizID) (*domainQuiz.Quiz, error)               { return nil, domainQuiz.ErrQuizNotFound }
func (m *mockQuizRepo) FindAll() ([]domainQuiz.Quiz, error)                                  { return nil, nil }
func (m *mockQuizRepo) FindAllSummaries() ([]*domainQuiz.QuizSummary, error)                 { return nil, nil }
func (m *mockQuizRepo) FindSummariesByCategory(_ domainQuiz.CategoryID) ([]*domainQuiz.QuizSummary, error) { return nil, nil }
func (m *mockQuizRepo) Save(_ *domainQuiz.Quiz) error                                       { return nil }
func (m *mockQuizRepo) Delete(_ domainQuiz.QuizID) error                                    { return nil }

// mockUserRepo for handler tests
type mockUserRepo struct {
	users map[string]*domainUser.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domainUser.User)}
}

func (m *mockUserRepo) FindByID(id domainUser.UserID) (*domainUser.User, error) {
	if u, ok := m.users[id.String()]; ok {
		return u, nil
	}
	return nil, domainUser.ErrUserNotFound
}

func (m *mockUserRepo) FindByTelegramUsername(_ domainUser.TelegramUsername) (*domainUser.User, error) {
	return nil, domainUser.ErrUserNotFound
}
func (m *mockUserRepo) FindAll(_, _ int) ([]domainUser.User, error) { return nil, nil }
func (m *mockUserRepo) Save(u *domainUser.User) error               { m.users[u.ID().String()] = u; return nil }
func (m *mockUserRepo) Delete(_ domainUser.UserID) error            { return nil }
func (m *mockUserRepo) Exists(id domainUser.UserID) (bool, error)   { _, ok := m.users[id.String()]; return ok, nil }

// mockEventBus collects events
type mockEventBus struct {
	events []domainDaily.Event
}

func (m *mockEventBus) Publish(event domainDaily.Event) {
	m.events = append(m.events, event)
}

// ========================================
// Test Helpers
// ========================================

func createTestQuestion(t *testing.T, position int) *domainQuiz.Question {
	t.Helper()
	text, _ := domainQuiz.NewQuestionText(fmt.Sprintf("Question %d", position))
	points, _ := domainQuiz.NewPoints(100)
	q, _ := domainQuiz.NewQuestion(domainQuiz.NewQuestionID(), text, points, position)

	correctText, _ := domainQuiz.NewAnswerText("Correct")
	correct, _ := domainQuiz.NewAnswer(domainQuiz.NewAnswerID(), correctText, true, 1)
	q.AddAnswer(*correct)

	for i := 2; i <= 4; i++ {
		wrongText, _ := domainQuiz.NewAnswerText(fmt.Sprintf("Wrong %d", i))
		wrong, _ := domainQuiz.NewAnswer(domainQuiz.NewAnswerID(), wrongText, false, i)
		q.AddAnswer(*wrong)
	}
	return q
}

func createTestQuestions(t *testing.T, count int) []*domainQuiz.Question {
	t.Helper()
	qs := make([]*domainQuiz.Question, count)
	for i := 0; i < count; i++ {
		qs[i] = createTestQuestion(t, i+1)
	}
	return qs
}

// handlerFixture holds everything needed for handler tests
type handlerFixture struct {
	app           *fiber.App
	handler       *DailyChallengeHandler
	dailyQuizRepo *mockDailyQuizRepo
	dailyGameRepo *mockDailyGameRepo
	questionRepo  *mockQuestionRepo
	userRepo      *mockUserRepo
	eventBus      *mockEventBus
	questions     []*domainQuiz.Question
	dailyQuiz     *domainDaily.DailyQuiz
	date          domainDaily.Date
}

func setupHandlerFixture(t *testing.T) *handlerFixture {
	t.Helper()

	date := domainDaily.NewDate(2026, time.January, 25)
	questions := createTestQuestions(t, 10)

	questionIDs := make([]domainDaily.QuestionID, 10)
	for i, q := range questions {
		questionIDs[i] = q.ID()
	}

	dateTime, _ := time.Parse("2006-01-02", date.String())
	expiresAt := dateTime.AddDate(0, 0, 1).Unix()
	dailyQuiz, _ := domainDaily.NewDailyQuiz(date, questionIDs, expiresAt, int64(1000000))

	dailyQuizRepo := newMockDailyQuizRepo()
	dailyQuizRepo.Save(dailyQuiz)

	dailyGameRepo := newMockDailyGameRepo()

	questionRepo := newMockQuestionRepo()
	for _, q := range questions {
		questionRepo.questions[q.ID().String()] = q
	}

	quizRepo := &mockQuizRepo{}

	userRepo := newMockUserRepo()
	uid, _ := shared.NewUserID("player123")
	uname, _ := domainUser.NewUsername("TestPlayer")
	u, _ := domainUser.NewUser(uid, uname, int64(1000000))
	userRepo.users[uid.String()] = u

	eventBus := &mockEventBus{events: make([]domainDaily.Event, 0)}

	// Create use cases
	getOrCreateQuizUC := appDaily.NewGetOrCreateDailyQuizUseCase(dailyQuizRepo, dailyGameRepo, questionRepo, eventBus)
	leaderboardUC := appDaily.NewGetDailyLeaderboardUseCase(dailyGameRepo, userRepo)

	rng := rand.New(rand.NewSource(42))
	chestCalc := domainDaily.NewChestRewardCalculator(rng)

	startUC := appDaily.NewStartDailyChallengeUseCase(dailyQuizRepo, dailyGameRepo, questionRepo, quizRepo, eventBus, getOrCreateQuizUC)
	submitUC := appDaily.NewSubmitDailyAnswerUseCase(dailyGameRepo, eventBus, leaderboardUC, chestCalc)
	statusUC := appDaily.NewGetDailyGameStatusUseCase(dailyQuizRepo, dailyGameRepo, leaderboardUC)
	streakUC := appDaily.NewGetPlayerStreakUseCase(dailyGameRepo)
	openChestUC := appDaily.NewOpenChestUseCase(dailyGameRepo, nil)
	retryUC := appDaily.NewRetryChallengeUseCase(dailyGameRepo, dailyQuizRepo, questionRepo, eventBus, nil, nil)

	handler := NewDailyChallengeHandler(getOrCreateQuizUC, startUC, submitUC, statusUC, leaderboardUC, streakUC, openChestUC, retryUC, nil)

	// Create Fiber app and register routes
	app := fiber.New()
	daily := app.Group("/api/v1/daily-challenge")
	daily.Post("/start", handler.StartDailyChallenge)
	daily.Post("/:gameId/answer", handler.SubmitDailyAnswer)
	daily.Post("/:gameId/chest/open", handler.OpenChest)
	daily.Post("/:gameId/retry", handler.RetryChallenge)
	daily.Get("/status", handler.GetDailyStatus)
	daily.Get("/leaderboard", handler.GetDailyLeaderboard)
	daily.Get("/streak", handler.GetPlayerStreak)

	return &handlerFixture{
		app:           app,
		handler:       handler,
		dailyQuizRepo: dailyQuizRepo,
		dailyGameRepo: dailyGameRepo,
		questionRepo:  questionRepo,
		userRepo:      userRepo,
		eventBus:      eventBus,
		questions:     questions,
		dailyQuiz:     dailyQuiz,
		date:          date,
	}
}

func jsonBody(t *testing.T, v interface{}) io.Reader {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	return bytes.NewReader(data)
}

func parseJSONResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v\nBody: %s", err, string(body))
	}
	return result
}

// ========================================
// StartDailyChallenge Handler Tests
// ========================================

func TestHandler_StartDailyChallenge_201(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/start", jsonBody(t, map[string]string{
		"playerId": "player123",
		"date":     f.date.String(),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 201. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'data' field in response")
	}

	game, ok := data["game"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'game' in data")
	}
	if game["gameId"] == nil || game["gameId"] == "" {
		t.Error("Expected gameId in game")
	}
	if game["status"] != "in_progress" {
		t.Errorf("status = %v, want in_progress", game["status"])
	}

	firstQuestion, ok := data["firstQuestion"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'firstQuestion' in data")
	}
	if firstQuestion["id"] == nil {
		t.Error("Expected question id")
	}
}

func TestHandler_StartDailyChallenge_400_MissingPlayerID(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/start", jsonBody(t, map[string]string{
		"date": f.date.String(),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

func TestHandler_StartDailyChallenge_409_AlreadyPlayed(t *testing.T) {
	f := setupHandlerFixture(t)

	body := map[string]string{
		"playerId": "player123",
		"date":     f.date.String(),
	}

	// First request - should succeed
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/start", jsonBody(t, body))
	req1.Header.Set("Content-Type", "application/json")
	resp1, _ := f.app.Test(req1)
	if resp1.StatusCode != 201 {
		t.Fatalf("First request: status = %d, want 201", resp1.StatusCode)
	}

	// Second request - should be 409
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/start", jsonBody(t, body))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := f.app.Test(req2)

	if resp2.StatusCode != 409 {
		body, _ := io.ReadAll(resp2.Body)
		t.Errorf("Status = %d, want 409. Body: %s", resp2.StatusCode, string(body))
	}
}

// ========================================
// SubmitDailyAnswer Handler Tests
// ========================================

func TestHandler_SubmitDailyAnswer_200(t *testing.T) {
	f := setupHandlerFixture(t)

	// Start a game first
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/start", jsonBody(t, map[string]string{
		"playerId": "player123",
		"date":     f.date.String(),
	}))
	startReq.Header.Set("Content-Type", "application/json")
	startResp, _ := f.app.Test(startReq)
	startResult := parseJSONResponse(t, startResp)
	startData := startResult["data"].(map[string]interface{})
	game := startData["game"].(map[string]interface{})
	gameID := game["gameId"].(string)
	firstQ := startData["firstQuestion"].(map[string]interface{})
	questionID := firstQ["id"].(string)
	answers := firstQ["answers"].([]interface{})
	answerID := answers[0].(map[string]interface{})["id"].(string)

	// Submit answer
	answerReq := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/"+gameID+"/answer", jsonBody(t, map[string]interface{}{
		"questionId": questionID,
		"answerId":   answerID,
		"playerId":   "player123",
		"timeTaken":  2000,
	}))
	answerReq.Header.Set("Content-Type", "application/json")

	resp, err := f.app.Test(answerReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 200. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data := result["data"].(map[string]interface{})
	if data["totalQuestions"] != float64(10) {
		t.Errorf("totalQuestions = %v, want 10", data["totalQuestions"])
	}
}

func TestHandler_SubmitDailyAnswer_400_MissingFields(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/some-game-id/answer", jsonBody(t, map[string]string{
		"playerId": "player123",
		// Missing questionId and answerId
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

func TestHandler_SubmitDailyAnswer_404_GameNotFound(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/nonexistent-id/answer", jsonBody(t, map[string]interface{}{
		"questionId": "q1",
		"answerId":   "a1",
		"playerId":   "player123",
		"timeTaken":  2000,
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 404. Body: %s", resp.StatusCode, string(body))
	}
}

// ========================================
// GetDailyStatus Handler Tests
// ========================================

func TestHandler_GetDailyStatus_200(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-challenge/status?playerId=player123&date="+f.date.String(), nil)

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 200. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data := result["data"].(map[string]interface{})
	if data["hasPlayed"] != false {
		t.Errorf("hasPlayed = %v, want false (no game played yet)", data["hasPlayed"])
	}
}

func TestHandler_GetDailyStatus_400_MissingPlayerID(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-challenge/status", nil)

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

// ========================================
// GetDailyLeaderboard Handler Tests
// ========================================

func TestHandler_GetDailyLeaderboard_200(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-challenge/leaderboard?date="+f.date.String(), nil)

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 200. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data := result["data"].(map[string]interface{})
	if data["date"] == nil {
		t.Error("Expected 'date' in response")
	}
	if data["entries"] == nil {
		t.Error("Expected 'entries' in response")
	}
}

func TestHandler_GetDailyLeaderboard_LimitValidation(t *testing.T) {
	f := setupHandlerFixture(t)

	// Negative limit should still work (clamped to 10)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-challenge/leaderboard?limit=-5", nil)
	resp, _ := f.app.Test(req)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 200. Body: %s", resp.StatusCode, string(body))
	}
}

// ========================================
// GetPlayerStreak Handler Tests
// ========================================

func TestHandler_GetPlayerStreak_200(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-challenge/streak?playerId=player123", nil)

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 200. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data := result["data"].(map[string]interface{})
	streak := data["streak"].(map[string]interface{})
	if streak["currentStreak"] != float64(0) {
		t.Errorf("currentStreak = %v, want 0 (no games)", streak["currentStreak"])
	}
}

func TestHandler_GetPlayerStreak_400_MissingPlayerID(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/daily-challenge/streak", nil)
	resp, _ := f.app.Test(req)

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

// ========================================
// OpenChest Handler Tests
// ========================================

func TestHandler_OpenChest_200(t *testing.T) {
	f := setupHandlerFixture(t)

	// Create a completed game
	pid, _ := shared.NewUserID("player123")
	dailyQuizID := domainDaily.NewDailyQuizID()
	quizAgg := makeQuizAggregate(t, f.questions)
	streak := domainDaily.NewStreakSystem()

	game, _ := domainDaily.NewDailyGame(pid, dailyQuizID, f.date, quizAgg, streak, int64(1000000))
	game.Events()

	// Answer all questions
	quizQuestions := quizAgg.Questions()
	for i, q := range quizQuestions {
		correct := q.Answers()[0]
		game.AnswerQuestion(q.ID(), correct.ID(), 2000, int64(1000000+(i+1)*2000))
	}
	game.Events()

	// Set chest reward
	rng := rand.New(rand.NewSource(42))
	calc := domainDaily.NewChestRewardCalculator(rng)
	chestType := domainDaily.CalculateChestType(game.GetCorrectAnswersCount(), quizAgg.QuestionsCount())
	reward := calc.CalculateRewards(chestType, game.Streak().GetBonus())
	game.SetChestReward(reward)

	f.dailyGameRepo.Save(game)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/"+game.ID().String()+"/chest/open", jsonBody(t, map[string]string{
		"playerId": "player123",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 200. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data := result["data"].(map[string]interface{})
	if data["chestType"] == nil || data["chestType"] == "" {
		t.Error("Expected chestType in response")
	}
}

func TestHandler_OpenChest_400_NotCompleted(t *testing.T) {
	f := setupHandlerFixture(t)

	// Start a game but don't complete it
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/start", jsonBody(t, map[string]string{
		"playerId": "player123",
		"date":     f.date.String(),
	}))
	startReq.Header.Set("Content-Type", "application/json")
	startResp, _ := f.app.Test(startReq)
	startResult := parseJSONResponse(t, startResp)
	gameID := startResult["data"].(map[string]interface{})["game"].(map[string]interface{})["gameId"].(string)

	// Try to open chest
	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/"+gameID+"/chest/open", jsonBody(t, map[string]string{
		"playerId": "player123",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

func TestHandler_OpenChest_404_NotFound(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/nonexistent/chest/open", jsonBody(t, map[string]string{
		"playerId": "player123",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 404. Body: %s", resp.StatusCode, string(body))
	}
}

// ========================================
// RetryChallenge Handler Tests
// ========================================

func TestHandler_RetryChallenge_201(t *testing.T) {
	f := setupHandlerFixture(t)

	// Create a completed game
	pid, _ := shared.NewUserID("player123")
	dailyQuizID := domainDaily.NewDailyQuizID()
	quizAgg := makeQuizAggregate(t, f.questions)
	streak := domainDaily.NewStreakSystem()

	game, _ := domainDaily.NewDailyGame(pid, dailyQuizID, f.date, quizAgg, streak, int64(1000000))
	game.Events()

	quizQuestions := quizAgg.Questions()
	for i, q := range quizQuestions {
		correct := q.Answers()[0]
		game.AnswerQuestion(q.ID(), correct.ID(), 2000, int64(1000000+(i+1)*2000))
	}
	game.Events()

	f.dailyGameRepo.Save(game)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/"+game.ID().String()+"/retry", jsonBody(t, map[string]string{
		"playerId":      "player123",
		"paymentMethod": "coins",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Status = %d, want 201. Body: %s", resp.StatusCode, string(body))
	}

	result := parseJSONResponse(t, resp)
	data := result["data"].(map[string]interface{})
	if data["newGameId"] == nil || data["newGameId"] == "" {
		t.Error("Expected newGameId in response")
	}
	if data["coinsDeducted"] != float64(100) {
		t.Errorf("coinsDeducted = %v, want 100", data["coinsDeducted"])
	}
}

func TestHandler_RetryChallenge_400_InvalidPayment(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/some-id/retry", jsonBody(t, map[string]string{
		"playerId":      "player123",
		"paymentMethod": "bitcoin",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

func TestHandler_RetryChallenge_400_MissingFields(t *testing.T) {
	f := setupHandlerFixture(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-challenge/some-id/retry", jsonBody(t, map[string]string{
		"playerId": "player123",
		// Missing paymentMethod
	}))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := f.app.Test(req)

	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Status = %d, want 400. Body: %s", resp.StatusCode, string(body))
	}
}

// ========================================
// Error Mapping Tests
// ========================================

func TestMapDailyChallengeError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{"DailyQuizNotFound -> 404", domainDaily.ErrDailyQuizNotFound, 404},
		{"GameNotFound -> 404", domainDaily.ErrGameNotFound, 404},
		{"AlreadyPlayedToday -> 409", domainDaily.ErrAlreadyPlayedToday, 409},
		{"GameAlreadyCompleted -> 400", domainDaily.ErrGameAlreadyCompleted, 400},
		{"GameNotActive -> 400", domainDaily.ErrGameNotActive, 400},
		{"QuestionNotFound -> 404", domainQuiz.ErrQuestionNotFound, 404},
		{"AnswerNotFound -> 404", domainQuiz.ErrAnswerNotFound, 404},
		{"Unknown error -> 500", fmt.Errorf("some unknown error"), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiberErr := mapDailyChallengeError(tt.err)

			fe, ok := fiberErr.(*fiber.Error)
			if !ok {
				t.Fatalf("Expected *fiber.Error, got %T", fiberErr)
			}
			if fe.Code != tt.expectedCode {
				t.Errorf("Code = %d, want %d", fe.Code, tt.expectedCode)
			}
		})
	}
}

// ========================================
// Helper: create quiz aggregate
// ========================================

func makeQuizAggregate(t *testing.T, questions []*domainQuiz.Question) *domainQuiz.Quiz {
	t.Helper()
	quizID := domainQuiz.NewQuizID()
	title, _ := domainQuiz.NewQuizTitle("Test Daily Quiz")
	timeLimit, _ := domainQuiz.NewTimeLimit(150)
	passingScore, _ := domainQuiz.NewPassingScore(0)

	q, err := domainQuiz.NewQuiz(quizID, title, "Desc", domainQuiz.CategoryID{}, timeLimit, passingScore, int64(1000000))
	if err != nil {
		t.Fatalf("Failed to create quiz: %v", err)
	}

	basePoints, _ := domainQuiz.NewPoints(100)
	maxTimeBonus, _ := domainQuiz.NewPoints(75)
	q.SetBasePoints(basePoints)
	q.SetTimeLimitPerQuestion(15)
	q.SetMaxTimeBonus(maxTimeBonus)

	for _, question := range questions {
		if err := q.AddQuestion(*question); err != nil {
			t.Fatalf("Failed to add question: %v", err)
		}
	}

	return q
}
