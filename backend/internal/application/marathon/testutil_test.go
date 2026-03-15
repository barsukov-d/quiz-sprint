package marathon

import (
	"fmt"
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// Constants
// ========================================

const testPlayerID = "player123"
const testPlayerID2 = "player456"

// ========================================
// Mock Repositories
// ========================================

// mockMarathonRepo is an in-memory marathon game repository
type mockMarathonRepo struct {
	games map[string]*solo_marathon.MarathonGameV2
}

func newMockMarathonRepo() *mockMarathonRepo {
	return &mockMarathonRepo{games: make(map[string]*solo_marathon.MarathonGameV2)}
}

func (m *mockMarathonRepo) Save(game *solo_marathon.MarathonGameV2) error {
	m.games[game.ID().String()] = game
	return nil
}

func (m *mockMarathonRepo) FindByID(id solo_marathon.GameID) (*solo_marathon.MarathonGameV2, error) {
	if g, ok := m.games[id.String()]; ok {
		return g, nil
	}
	return nil, solo_marathon.ErrGameNotFound
}

func (m *mockMarathonRepo) FindActiveByPlayer(playerID solo_marathon.UserID) (*solo_marathon.MarathonGameV2, error) {
	for _, g := range m.games {
		if g.PlayerID().Equals(playerID) &&
			(g.Status() == solo_marathon.GameStatusInProgress || g.Status() == solo_marathon.GameStatusGameOver) {
			return g, nil
		}
	}
	return nil, solo_marathon.ErrGameNotFound
}

func (m *mockMarathonRepo) Delete(id solo_marathon.GameID) error {
	delete(m.games, id.String())
	return nil
}

// mockPersonalBestRepo is an in-memory personal best repository
type mockPersonalBestRepo struct {
	records map[string]*solo_marathon.PersonalBest // keyed by "playerID:categoryID"
}

func newMockPersonalBestRepo() *mockPersonalBestRepo {
	return &mockPersonalBestRepo{records: make(map[string]*solo_marathon.PersonalBest)}
}

func (m *mockPersonalBestRepo) key(playerID solo_marathon.UserID, category solo_marathon.MarathonCategory) string {
	return playerID.String() + ":" + category.CategoryID().String()
}

func (m *mockPersonalBestRepo) Save(pb *solo_marathon.PersonalBest) error {
	m.records[m.key(pb.PlayerID(), pb.Category())] = pb
	return nil
}

func (m *mockPersonalBestRepo) FindByPlayerAndCategory(playerID solo_marathon.UserID, category solo_marathon.MarathonCategory) (*solo_marathon.PersonalBest, error) {
	if pb, ok := m.records[m.key(playerID, category)]; ok {
		return pb, nil
	}
	return nil, solo_marathon.ErrPersonalBestNotFound
}

func (m *mockPersonalBestRepo) FindTopByCategory(category solo_marathon.MarathonCategory, limit int) ([]*solo_marathon.PersonalBest, error) {
	var result []*solo_marathon.PersonalBest
	catID := category.CategoryID().String()
	for k, pb := range m.records {
		// Match category (key format: "playerID:categoryID")
		if len(k) > len(catID) && k[len(k)-len(catID):] == catID {
			result = append(result, pb)
		}
	}
	// Sort by score descending
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].BestScore() > result[i].BestScore() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockPersonalBestRepo) FindAllByPlayer(playerID solo_marathon.UserID) ([]*solo_marathon.PersonalBest, error) {
	var result []*solo_marathon.PersonalBest
	prefix := playerID.String() + ":"
	for k, pb := range m.records {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result = append(result, pb)
		}
	}
	if len(result) == 0 {
		return nil, solo_marathon.ErrPersonalBestNotFound
	}
	return result, nil
}

// mockBonusWalletRepo is an in-memory bonus wallet repository
type mockBonusWalletRepo struct {
	wallets map[string]*solo_marathon.BonusWallet // keyed by playerID
}

func newMockBonusWalletRepo() *mockBonusWalletRepo {
	return &mockBonusWalletRepo{wallets: make(map[string]*solo_marathon.BonusWallet)}
}

func (m *mockBonusWalletRepo) FindByPlayer(playerID solo_marathon.UserID) (*solo_marathon.BonusWallet, error) {
	if w, ok := m.wallets[playerID.String()]; ok {
		return w, nil
	}
	return nil, nil
}

func (m *mockBonusWalletRepo) Save(wallet *solo_marathon.BonusWallet) error {
	m.wallets[wallet.PlayerID().String()] = wallet
	return nil
}

// mockCategoryRepo is an in-memory category repository
type mockCategoryRepo struct {
	categories map[string]*quiz.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[string]*quiz.Category)}
}

func (m *mockCategoryRepo) FindByID(id quiz.CategoryID) (*quiz.Category, error) {
	if c, ok := m.categories[id.String()]; ok {
		return c, nil
	}
	return nil, quiz.ErrCategoryNotFound
}

func (m *mockCategoryRepo) FindAll() ([]*quiz.Category, error) {
	var result []*quiz.Category
	for _, c := range m.categories {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockCategoryRepo) Save(c *quiz.Category) error {
	m.categories[c.ID().String()] = c
	return nil
}

func (m *mockCategoryRepo) Delete(id quiz.CategoryID) error {
	delete(m.categories, id.String())
	return nil
}

// mockQuestionRepo is an in-memory question repository for marathon tests
type mockQuestionRepo struct {
	questions map[string]*quiz.Question
}

func newMockQuestionRepo() *mockQuestionRepo {
	return &mockQuestionRepo{questions: make(map[string]*quiz.Question)}
}

func (m *mockQuestionRepo) FindByID(id quiz.QuestionID) (*quiz.Question, error) {
	if q, ok := m.questions[id.String()]; ok {
		return q, nil
	}
	return nil, quiz.ErrQuestionNotFound
}

func (m *mockQuestionRepo) FindByIDs(ids []quiz.QuestionID) ([]*quiz.Question, error) {
	result := make([]*quiz.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := m.questions[id.String()]; ok {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *mockQuestionRepo) FindByFilter(_ quiz.QuestionFilter) ([]*quiz.Question, error) {
	var result []*quiz.Question
	for _, q := range m.questions {
		result = append(result, q)
	}
	return result, nil
}

func (m *mockQuestionRepo) FindRandomQuestions(f quiz.QuestionFilter, limit int) ([]*quiz.Question, error) {
	excluded := make(map[string]bool, len(f.ExcludeIDs))
	for _, id := range f.ExcludeIDs {
		excluded[id.String()] = true
	}
	var result []*quiz.Question
	for _, q := range m.questions {
		if excluded[q.ID().String()] {
			continue
		}
		result = append(result, q)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockQuestionRepo) FindQuestionsBySeed(f quiz.QuestionFilter, limit int, _ int64) ([]*quiz.Question, error) {
	return m.FindRandomQuestions(f, limit)
}

func (m *mockQuestionRepo) FindQuestionsByQuizSeed(n int, _ int64, _ *quiz.CategoryID) ([]*quiz.Question, error) {
	var result []*quiz.Question
	for _, q := range m.questions {
		result = append(result, q)
		if len(result) >= n {
			break
		}
	}
	if len(result) < n {
		return nil, fmt.Errorf("not enough questions: need %d, have %d", n, len(result))
	}
	return result, nil
}

func (m *mockQuestionRepo) CountByFilter(_ quiz.QuestionFilter) (int, error) {
	return len(m.questions), nil
}

func (m *mockQuestionRepo) Save(q *quiz.Question) error {
	m.questions[q.ID().String()] = q
	return nil
}

func (m *mockQuestionRepo) SaveAll(qs []*quiz.Question) error {
	for _, q := range qs {
		m.questions[q.ID().String()] = q
	}
	return nil
}

func (m *mockQuestionRepo) Delete(id quiz.QuestionID) error {
	delete(m.questions, id.String())
	return nil
}

// mockUserRepo for leaderboard tests
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

func (m *mockUserRepo) Save(u *domainUser.User) error {
	m.users[u.ID().String()] = u
	return nil
}

func (m *mockUserRepo) Delete(_ domainUser.UserID) error { return nil }

func (m *mockUserRepo) Exists(id domainUser.UserID) (bool, error) {
	_, ok := m.users[id.String()]
	return ok, nil
}

// mockEventBus collects published marathon events
type mockEventBus struct {
	events []solo_marathon.Event
}

func (m *mockEventBus) Publish(event solo_marathon.Event) {
	m.events = append(m.events, event)
}

// ========================================
// Test Helpers
// ========================================

func mustUserID(s string) shared.UserID {
	id, err := shared.NewUserID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func createTestQuestion(t *testing.T, position int) *quiz.Question {
	t.Helper()
	text, _ := quiz.NewQuestionText(fmt.Sprintf("Marathon Question %d", position))
	points, _ := quiz.NewPoints(100)
	q, err := quiz.NewQuestion(quiz.NewQuestionID(), text, points, position)
	if err != nil {
		t.Fatalf("Failed to create question: %v", err)
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

func createTestQuestions(t *testing.T, count int) []*quiz.Question {
	t.Helper()
	qs := make([]*quiz.Question, count)
	for i := 0; i < count; i++ {
		qs[i] = createTestQuestion(t, i+1)
	}
	return qs
}

func createTestUser(t *testing.T, id string, username string) *domainUser.User {
	t.Helper()
	userID, _ := shared.NewUserID(id)
	uname, _ := domainUser.NewUsername(username)
	u, err := domainUser.NewUser(userID, uname, int64(1000000))
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return u
}

// ========================================
// Test Fixture
// ========================================

type marathonFixture struct {
	marathonRepo     *mockMarathonRepo
	personalBestRepo *mockPersonalBestRepo
	bonusWalletRepo  *mockBonusWalletRepo
	questionRepo     *mockQuestionRepo
	categoryRepo     *mockCategoryRepo
	userRepo         *mockUserRepo
	eventBus         *mockEventBus
	questions        []*quiz.Question
}

func setupFixture(t *testing.T) *marathonFixture {
	t.Helper()

	questions := createTestQuestions(t, 20) // enough for several answers

	questionRepo := newMockQuestionRepo()
	for _, q := range questions {
		questionRepo.questions[q.ID().String()] = q
	}

	userRepo := newMockUserRepo()
	userRepo.users[testPlayerID] = createTestUser(t, testPlayerID, "TestPlayer")
	userRepo.users[testPlayerID2] = createTestUser(t, testPlayerID2, "TestPlayer2")

	return &marathonFixture{
		marathonRepo:     newMockMarathonRepo(),
		personalBestRepo: newMockPersonalBestRepo(),
		bonusWalletRepo:  newMockBonusWalletRepo(),
		questionRepo:     questionRepo,
		categoryRepo:     newMockCategoryRepo(),
		userRepo:         userRepo,
		eventBus:         &mockEventBus{events: make([]solo_marathon.Event, 0)},
		questions:        questions,
	}
}

// Use case constructors

func (f *marathonFixture) newStartUC() *StartMarathonUseCase {
	return NewStartMarathonUseCase(
		f.marathonRepo, f.personalBestRepo, f.questionRepo,
		f.categoryRepo, f.eventBus, f.bonusWalletRepo,
	)
}

func (f *marathonFixture) newSubmitAnswerUC() *SubmitMarathonAnswerUseCase {
	return NewSubmitMarathonAnswerUseCase(
		f.marathonRepo, f.personalBestRepo, f.questionRepo, f.eventBus,
	)
}

func (f *marathonFixture) newUseBonusUC() *UseMarathonBonusUseCase {
	return NewUseMarathonBonusUseCase(f.marathonRepo, f.questionRepo, f.eventBus)
}

func (f *marathonFixture) newContinueUC() *ContinueMarathonUseCase {
	return NewContinueMarathonUseCase(f.marathonRepo, f.questionRepo, f.eventBus, nil)
}

func (f *marathonFixture) newAbandonUC() *AbandonMarathonUseCase {
	return NewAbandonMarathonUseCase(f.marathonRepo, f.personalBestRepo, f.eventBus)
}

func (f *marathonFixture) newGetStatusUC() *GetMarathonStatusUseCase {
	return NewGetMarathonStatusUseCase(f.marathonRepo, f.bonusWalletRepo)
}

func (f *marathonFixture) newGetPersonalBestsUC() *GetPersonalBestsUseCase {
	return NewGetPersonalBestsUseCase(f.personalBestRepo)
}

func (f *marathonFixture) newGetLeaderboardUC() *GetMarathonLeaderboardUseCase {
	return NewGetMarathonLeaderboardUseCase(f.personalBestRepo, f.categoryRepo, f.userRepo)
}

// startGameForPlayer creates and starts a marathon game, returning the start output
func (f *marathonFixture) startGameForPlayer(t *testing.T, playerID string) StartMarathonOutput {
	t.Helper()
	uc := f.newStartUC()
	output, err := uc.Execute(StartMarathonInput{PlayerID: playerID})
	if err != nil {
		t.Fatalf("Failed to start marathon: %v", err)
	}
	return output
}

// answerCurrentQuestion answers the current question (correct or wrong)
func (f *marathonFixture) answerCurrentQuestion(t *testing.T, gameID string, playerID string, correct bool) SubmitMarathonAnswerOutput {
	t.Helper()

	game, err := f.marathonRepo.FindByID(solo_marathon.NewGameIDFromString(gameID))
	if err != nil {
		t.Fatalf("Game not found: %v", err)
	}

	currentQ, err := game.GetCurrentQuestion()
	if err != nil {
		t.Fatalf("No current question: %v", err)
	}

	var answerID string
	if correct {
		answerID = currentQ.Answers()[0].ID().String() // first answer is correct
	} else {
		answerID = currentQ.Answers()[1].ID().String() // second is wrong
	}

	uc := f.newSubmitAnswerUC()
	output, err := uc.Execute(SubmitMarathonAnswerInput{
		GameID:     gameID,
		QuestionID: currentQ.ID().String(),
		AnswerID:   answerID,
		PlayerID:   playerID,
		TimeTaken:  2000,
	})
	if err != nil {
		t.Fatalf("Failed to submit answer: %v", err)
	}
	return output
}
