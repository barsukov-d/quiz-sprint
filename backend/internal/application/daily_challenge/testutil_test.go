package daily_challenge

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// Mock Repositories
// ========================================

// MockDailyQuizRepository is an in-memory DailyQuizRepository
type MockDailyQuizRepository struct {
	quizzes map[string]*daily_challenge.DailyQuiz // keyed by ID
	byDate  map[string]*daily_challenge.DailyQuiz // keyed by date string
}

func NewMockDailyQuizRepository() *MockDailyQuizRepository {
	return &MockDailyQuizRepository{
		quizzes: make(map[string]*daily_challenge.DailyQuiz),
		byDate:  make(map[string]*daily_challenge.DailyQuiz),
	}
}

func (m *MockDailyQuizRepository) Save(dailyQuiz *daily_challenge.DailyQuiz) error {
	m.quizzes[dailyQuiz.ID().String()] = dailyQuiz
	m.byDate[dailyQuiz.Date().String()] = dailyQuiz
	return nil
}

func (m *MockDailyQuizRepository) FindByID(id daily_challenge.DailyQuizID) (*daily_challenge.DailyQuiz, error) {
	if q, ok := m.quizzes[id.String()]; ok {
		return q, nil
	}
	return nil, daily_challenge.ErrDailyQuizNotFound
}

func (m *MockDailyQuizRepository) FindByDate(date daily_challenge.Date) (*daily_challenge.DailyQuiz, error) {
	if q, ok := m.byDate[date.String()]; ok {
		return q, nil
	}
	return nil, daily_challenge.ErrDailyQuizNotFound
}

func (m *MockDailyQuizRepository) Delete(id daily_challenge.DailyQuizID) error {
	if q, ok := m.quizzes[id.String()]; ok {
		delete(m.byDate, q.Date().String())
		delete(m.quizzes, id.String())
		return nil
	}
	return daily_challenge.ErrDailyQuizNotFound
}

// MockDailyGameRepository is an in-memory DailyGameRepository
type MockDailyGameRepository struct {
	games map[string]*daily_challenge.DailyGame // keyed by game ID
}

func NewMockDailyGameRepository() *MockDailyGameRepository {
	return &MockDailyGameRepository{
		games: make(map[string]*daily_challenge.DailyGame),
	}
}

func (m *MockDailyGameRepository) Save(game *daily_challenge.DailyGame) error {
	m.games[game.ID().String()] = game
	return nil
}

func (m *MockDailyGameRepository) FindByID(id daily_challenge.GameID) (*daily_challenge.DailyGame, error) {
	if g, ok := m.games[id.String()]; ok {
		return g, nil
	}
	return nil, daily_challenge.ErrGameNotFound
}

func (m *MockDailyGameRepository) FindByPlayerAndDate(playerID daily_challenge.UserID, date daily_challenge.Date) (*daily_challenge.DailyGame, error) {
	// Return the best (highest score) completed game, or the first in-progress game
	var best *daily_challenge.DailyGame
	for _, g := range m.games {
		if g.PlayerID() == playerID && g.Date().Equals(date) {
			if best == nil {
				best = g
			} else if g.GetFinalScore() > best.GetFinalScore() {
				best = g
			}
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, daily_challenge.ErrGameNotFound
}

func (m *MockDailyGameRepository) FindAllAttemptsByPlayerAndDate(playerID daily_challenge.UserID, date daily_challenge.Date) ([]*daily_challenge.DailyGame, error) {
	var result []*daily_challenge.DailyGame
	for _, g := range m.games {
		if g.PlayerID() == playerID && g.Date().Equals(date) {
			result = append(result, g)
		}
	}
	return result, nil
}

func (m *MockDailyGameRepository) CountAttemptsByPlayerAndDate(playerID daily_challenge.UserID, date daily_challenge.Date) (int, error) {
	count := 0
	for _, g := range m.games {
		if g.PlayerID() == playerID && g.Date().Equals(date) {
			count++
		}
	}
	return count, nil
}

func (m *MockDailyGameRepository) FindTopByDate(date daily_challenge.Date, limit int) ([]*daily_challenge.DailyGame, error) {
	var games []*daily_challenge.DailyGame
	for _, g := range m.games {
		if g.Date().Equals(date) && g.IsCompleted() {
			games = append(games, g)
		}
	}
	// Simple sort by score descending
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

func (m *MockDailyGameRepository) FindTopByDateAndFriends(date daily_challenge.Date, playerID daily_challenge.UserID, limit int) ([]*daily_challenge.DailyGame, error) {
	return m.FindTopByDate(date, limit)
}

func (m *MockDailyGameRepository) FindTopByDateAndCountry(date daily_challenge.Date, playerID daily_challenge.UserID, limit int) ([]*daily_challenge.DailyGame, error) {
	return m.FindTopByDate(date, limit)
}

func (m *MockDailyGameRepository) GetPlayerRankByDate(playerID daily_challenge.UserID, date daily_challenge.Date) (int, error) {
	topGames, _ := m.FindTopByDate(date, 0)
	for i, g := range topGames {
		if g.PlayerID() == playerID {
			return i + 1, nil
		}
	}
	return 0, nil
}

func (m *MockDailyGameRepository) GetTotalPlayersByDate(date daily_challenge.Date) (int, error) {
	seen := make(map[string]bool)
	for _, g := range m.games {
		if g.Date().Equals(date) && g.IsCompleted() {
			seen[g.PlayerID().String()] = true
		}
	}
	return len(seen), nil
}

func (m *MockDailyGameRepository) Delete(id daily_challenge.GameID) error {
	if _, ok := m.games[id.String()]; ok {
		delete(m.games, id.String())
		return nil
	}
	return daily_challenge.ErrGameNotFound
}

func (m *MockDailyGameRepository) MarkAbandonedGames() (int, error) {
	return 0, nil
}

// MockQuestionRepository is an in-memory QuestionRepository
type MockQuestionRepository struct {
	questions map[string]*quiz.Question // keyed by question ID
}

func NewMockQuestionRepository() *MockQuestionRepository {
	return &MockQuestionRepository{
		questions: make(map[string]*quiz.Question),
	}
}

func (m *MockQuestionRepository) AddQuestion(q *quiz.Question) {
	m.questions[q.ID().String()] = q
}

func (m *MockQuestionRepository) FindByID(id quiz.QuestionID) (*quiz.Question, error) {
	if q, ok := m.questions[id.String()]; ok {
		return q, nil
	}
	return nil, quiz.ErrQuestionNotFound
}

func (m *MockQuestionRepository) FindByIDs(ids []quiz.QuestionID) ([]*quiz.Question, error) {
	result := make([]*quiz.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := m.questions[id.String()]; ok {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *MockQuestionRepository) FindByFilter(_ quiz.QuestionFilter) ([]*quiz.Question, error) {
	var result []*quiz.Question
	for _, q := range m.questions {
		result = append(result, q)
	}
	return result, nil
}

func (m *MockQuestionRepository) FindRandomQuestions(_ quiz.QuestionFilter, limit int) ([]*quiz.Question, error) {
	var result []*quiz.Question
	for _, q := range m.questions {
		result = append(result, q)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *MockQuestionRepository) FindQuestionsBySeed(filter quiz.QuestionFilter, limit int, _ int64) ([]*quiz.Question, error) {
	return m.FindRandomQuestions(filter, limit)
}

func (m *MockQuestionRepository) FindQuestionsByQuizSeed(questionsPerQuiz int, _ int64, _ *quiz.CategoryID) ([]*quiz.Question, error) {
	var result []*quiz.Question
	for _, q := range m.questions {
		result = append(result, q)
		if len(result) >= questionsPerQuiz {
			break
		}
	}
	if len(result) < questionsPerQuiz {
		return nil, fmt.Errorf("not enough questions: need %d, have %d", questionsPerQuiz, len(result))
	}
	return result, nil
}

func (m *MockQuestionRepository) CountByFilter(_ quiz.QuestionFilter) (int, error) {
	return len(m.questions), nil
}

func (m *MockQuestionRepository) Save(question *quiz.Question) error {
	m.questions[question.ID().String()] = question
	return nil
}

func (m *MockQuestionRepository) SaveAll(questions []*quiz.Question) error {
	for _, q := range questions {
		m.questions[q.ID().String()] = q
	}
	return nil
}

func (m *MockQuestionRepository) Delete(id quiz.QuestionID) error {
	delete(m.questions, id.String())
	return nil
}

// MockQuizRepository is an in-memory QuizRepository (minimal, used by StartDailyChallenge)
type MockQuizRepository struct {
	quizzes map[string]*quiz.Quiz
}

func NewMockQuizRepository() *MockQuizRepository {
	return &MockQuizRepository{
		quizzes: make(map[string]*quiz.Quiz),
	}
}

func (m *MockQuizRepository) FindByID(id quiz.QuizID) (*quiz.Quiz, error) {
	if q, ok := m.quizzes[id.String()]; ok {
		return q, nil
	}
	return nil, quiz.ErrQuizNotFound
}

func (m *MockQuizRepository) FindAll() ([]quiz.Quiz, error) {
	return nil, nil
}

func (m *MockQuizRepository) FindAllSummaries() ([]*quiz.QuizSummary, error) {
	return nil, nil
}

func (m *MockQuizRepository) FindSummariesByCategory(_ quiz.CategoryID) ([]*quiz.QuizSummary, error) {
	return nil, nil
}

func (m *MockQuizRepository) Save(q *quiz.Quiz) error {
	m.quizzes[q.ID().String()] = q
	return nil
}

func (m *MockQuizRepository) Delete(id quiz.QuizID) error {
	delete(m.quizzes, id.String())
	return nil
}

// MockUserRepository is an in-memory UserRepository
type MockUserRepository struct {
	users map[string]*domainUser.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domainUser.User),
	}
}

func (m *MockUserRepository) AddUser(u *domainUser.User) {
	m.users[u.ID().String()] = u
}

func (m *MockUserRepository) FindByID(id domainUser.UserID) (*domainUser.User, error) {
	if u, ok := m.users[id.String()]; ok {
		return u, nil
	}
	return nil, domainUser.ErrUserNotFound
}

func (m *MockUserRepository) FindByTelegramUsername(_ domainUser.TelegramUsername) (*domainUser.User, error) {
	return nil, domainUser.ErrUserNotFound
}

func (m *MockUserRepository) FindAll(_, _ int) ([]domainUser.User, error) {
	return nil, nil
}

func (m *MockUserRepository) Save(u *domainUser.User) error {
	m.users[u.ID().String()] = u
	return nil
}

func (m *MockUserRepository) Delete(_ domainUser.UserID) error {
	return nil
}

func (m *MockUserRepository) Exists(id domainUser.UserID) (bool, error) {
	_, ok := m.users[id.String()]
	return ok, nil
}

// MockEventBus collects published events
type MockEventBus struct {
	Events []daily_challenge.Event
}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		Events: make([]daily_challenge.Event, 0),
	}
}

func (m *MockEventBus) Publish(event daily_challenge.Event) {
	m.Events = append(m.Events, event)
}

// ========================================
// Test Helpers
// ========================================

const testPlayerID = "player123"
const testPlayerID2 = "player456"

func testDate() daily_challenge.Date {
	return daily_challenge.NewDate(2026, time.January, 25)
}

// newTestQuestion creates a test question with 4 answers (first is correct)
func newTestQuestion(t *testing.T, position int) *quiz.Question {
	t.Helper()

	questionText, _ := quiz.NewQuestionText(fmt.Sprintf("Test Question %d", position))
	points, _ := quiz.NewPoints(100)

	q, err := quiz.NewQuestion(
		quiz.NewQuestionID(),
		questionText,
		points,
		position,
	)
	if err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	correctText, _ := quiz.NewAnswerText("Correct Answer")
	correct, _ := quiz.NewAnswer(quiz.NewAnswerID(), correctText, true, 1)
	q.AddAnswer(*correct)

	for i := 2; i <= 4; i++ {
		wrongText, _ := quiz.NewAnswerText(fmt.Sprintf("Wrong Answer %d", i))
		wrong, _ := quiz.NewAnswer(quiz.NewAnswerID(), wrongText, false, i)
		q.AddAnswer(*wrong)
	}

	return q
}

// newTestQuestions creates N test questions
func newTestQuestions(t *testing.T, count int) []*quiz.Question {
	t.Helper()
	questions := make([]*quiz.Question, count)
	for i := 0; i < count; i++ {
		questions[i] = newTestQuestion(t, i+1)
	}
	return questions
}

// newTestQuizAggregate creates a quiz.Quiz with 10 questions
func newTestQuizAggregate(t *testing.T, questions []*quiz.Question) *quiz.Quiz {
	t.Helper()

	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("Daily Test Quiz")
	timeLimit, _ := quiz.NewTimeLimit(150)
	passingScore, _ := quiz.NewPassingScore(0)

	q, err := quiz.NewQuiz(quizID, title, "Test description", quiz.CategoryID{}, timeLimit, passingScore, int64(1000000))
	if err != nil {
		t.Fatalf("Failed to create quiz: %v", err)
	}

	basePoints, _ := quiz.NewPoints(100)
	maxTimeBonus, _ := quiz.NewPoints(75)
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

// newTestDailyQuiz creates a DailyQuiz with given questions for a date
func newTestDailyQuiz(t *testing.T, date daily_challenge.Date, questions []*quiz.Question) *daily_challenge.DailyQuiz {
	t.Helper()

	questionIDs := make([]daily_challenge.QuestionID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID()
	}

	dateTime, _ := time.Parse("2006-01-02", date.String())
	expiresAt := dateTime.AddDate(0, 0, 1).Unix()

	dq, err := daily_challenge.NewDailyQuiz(date, questionIDs, expiresAt, int64(1000000))
	if err != nil {
		t.Fatalf("Failed to create daily quiz: %v", err)
	}

	return dq
}

// newTestUser creates a domain user
func newTestUser(t *testing.T, id string, username string) *domainUser.User {
	t.Helper()
	userID, _ := shared.NewUserID(id)
	uname, _ := domainUser.NewUsername(username)
	u, err := domainUser.NewUser(userID, uname, int64(1000000))
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return u
}

// newTestChestRewardCalc creates a deterministic ChestRewardCalculator
func newTestChestRewardCalc() *daily_challenge.ChestRewardCalculator {
	rng := rand.New(rand.NewSource(42))
	return daily_challenge.NewChestRewardCalculator(rng)
}

// newCompletedGame creates a completed DailyGame by answering all questions correctly
func newCompletedGame(t *testing.T, playerID string, date daily_challenge.Date, questions []*quiz.Question, streak daily_challenge.StreakSystem) *daily_challenge.DailyGame {
	t.Helper()

	pid, _ := shared.NewUserID(playerID)
	dailyQuizID := daily_challenge.NewDailyQuizID()
	quizAgg := newTestQuizAggregate(t, questions)
	now := int64(1000000)

	game, err := daily_challenge.NewDailyGame(pid, dailyQuizID, date, quizAgg, streak, now)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	game.Events() // clear startup events

	// Answer all questions correctly
	quizQuestions := quizAgg.Questions()
	for i, q := range quizQuestions {
		correctAnswer := q.Answers()[0]
		_, err := game.AnswerQuestion(q.ID(), correctAnswer.ID(), 2000, now+int64((i+1)*2000))
		if err != nil {
			t.Fatalf("Failed to answer question %d: %v", i, err)
		}
	}
	game.Events() // clear answer/completion events

	// Set chest reward
	calc := newTestChestRewardCalc()
	chestType := daily_challenge.CalculateChestType(game.GetCorrectAnswersCount(), quizAgg.QuestionsCount())
	reward := calc.CalculateRewards(chestType, game.Streak().GetBonus())
	game.SetChestReward(reward)

	return game
}

// ========================================
// Setup Helpers for Use Cases
// ========================================

type testFixture struct {
	dailyQuizRepo *MockDailyQuizRepository
	dailyGameRepo *MockDailyGameRepository
	questionRepo  *MockQuestionRepository
	quizRepo      *MockQuizRepository
	userRepo      *MockUserRepository
	eventBus      *MockEventBus
	questions     []*quiz.Question
	dailyQuiz     *daily_challenge.DailyQuiz
	date          daily_challenge.Date
}

// setupFixture creates a standard test fixture with 10 questions and a daily quiz
func setupFixture(t *testing.T) *testFixture {
	t.Helper()

	date := testDate()
	questions := newTestQuestions(t, 10)
	dailyQuiz := newTestDailyQuiz(t, date, questions)

	dailyQuizRepo := NewMockDailyQuizRepository()
	dailyQuizRepo.Save(dailyQuiz)

	dailyGameRepo := NewMockDailyGameRepository()

	questionRepo := NewMockQuestionRepository()
	for _, q := range questions {
		questionRepo.AddQuestion(q)
	}

	quizRepo := NewMockQuizRepository()
	userRepo := NewMockUserRepository()
	userRepo.AddUser(newTestUser(t, testPlayerID, "TestPlayer"))
	userRepo.AddUser(newTestUser(t, testPlayerID2, "TestPlayer2"))

	eventBus := NewMockEventBus()

	return &testFixture{
		dailyQuizRepo: dailyQuizRepo,
		dailyGameRepo: dailyGameRepo,
		questionRepo:  questionRepo,
		quizRepo:      quizRepo,
		userRepo:      userRepo,
		eventBus:      eventBus,
		questions:     questions,
		dailyQuiz:     dailyQuiz,
		date:          date,
	}
}

func (f *testFixture) newGetOrCreateQuizUC() *GetOrCreateDailyQuizUseCase {
	return NewGetOrCreateDailyQuizUseCase(f.dailyQuizRepo, f.dailyGameRepo, f.questionRepo, f.eventBus)
}

func (f *testFixture) newStartUC() *StartDailyChallengeUseCase {
	return NewStartDailyChallengeUseCase(
		f.dailyQuizRepo, f.dailyGameRepo, f.questionRepo, f.quizRepo,
		f.eventBus, f.newGetOrCreateQuizUC(),
	)
}

func (f *testFixture) newLeaderboardUC() *GetDailyLeaderboardUseCase {
	return NewGetDailyLeaderboardUseCase(f.dailyGameRepo, f.userRepo)
}

func (f *testFixture) newSubmitAnswerUC() *SubmitDailyAnswerUseCase {
	return NewSubmitDailyAnswerUseCase(f.dailyGameRepo, f.eventBus, f.newLeaderboardUC(), newTestChestRewardCalc())
}

func (f *testFixture) newGetStatusUC() *GetDailyGameStatusUseCase {
	return NewGetDailyGameStatusUseCase(f.dailyQuizRepo, f.dailyGameRepo, f.newLeaderboardUC())
}

func (f *testFixture) newGetStreakUC() *GetPlayerStreakUseCase {
	return NewGetPlayerStreakUseCase(f.dailyGameRepo)
}

func (f *testFixture) newOpenChestUC() *OpenChestUseCase {
	return NewOpenChestUseCase(f.dailyGameRepo, nil)
}

func (f *testFixture) newRetryUC() *RetryChallengeUseCase {
	return NewRetryChallengeUseCase(f.dailyGameRepo, f.dailyQuizRepo, f.questionRepo, f.eventBus, nil, nil)
}
