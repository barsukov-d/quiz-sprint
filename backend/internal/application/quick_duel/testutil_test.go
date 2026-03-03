package quick_duel

import (
	"context"
	"fmt"
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// Constants
// ========================================

const testPlayer1ID = "player111"
const testPlayer2ID = "player222"
const testPlayer3ID = "player333"

// ========================================
// Mock Repositories
// ========================================

// mockDuelGameRepo is an in-memory duel game repository
type mockDuelGameRepo struct {
	games map[string]*quick_duel.DuelGame
}

func newMockDuelGameRepo() *mockDuelGameRepo {
	return &mockDuelGameRepo{games: make(map[string]*quick_duel.DuelGame)}
}

func (m *mockDuelGameRepo) Save(game *quick_duel.DuelGame) error {
	m.games[game.ID().String()] = game
	return nil
}

func (m *mockDuelGameRepo) FindByID(id quick_duel.GameID) (*quick_duel.DuelGame, error) {
	if g, ok := m.games[id.String()]; ok {
		return g, nil
	}
	return nil, quick_duel.ErrGameNotFound
}

func (m *mockDuelGameRepo) FindActiveByPlayer(playerID quick_duel.UserID) (*quick_duel.DuelGame, error) {
	for _, g := range m.games {
		if g.Status() == quick_duel.GameStatusInProgress || g.Status() == quick_duel.GameStatusWaitingStart {
			if g.Player1().UserID().Equals(playerID) || g.Player2().UserID().Equals(playerID) {
				return g, nil
			}
		}
	}
	return nil, quick_duel.ErrGameNotFound
}

func (m *mockDuelGameRepo) FindByPlayerPaginated(playerID quick_duel.UserID, limit int, offset int, _ string) ([]*quick_duel.DuelGame, int, error) {
	var result []*quick_duel.DuelGame
	for _, g := range m.games {
		if g.Player1().UserID().Equals(playerID) || g.Player2().UserID().Equals(playerID) {
			result = append(result, g)
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []*quick_duel.DuelGame{}, total, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, total, nil
}

func (m *mockDuelGameRepo) Delete(id quick_duel.GameID) error {
	delete(m.games, id.String())
	return nil
}

func (m *mockDuelGameRepo) FindRecentOpponents(playerID quick_duel.UserID, limit int) ([]quick_duel.RecentOpponentEntry, error) {
	seen := make(map[string]int)
	lastPlayed := make(map[string]int64)

	for _, g := range m.games {
		if g.Status() != quick_duel.GameStatusFinished {
			continue
		}
		var opponentID string
		if g.Player1().UserID().Equals(playerID) {
			opponentID = g.Player2().UserID().String()
		} else if g.Player2().UserID().Equals(playerID) {
			opponentID = g.Player1().UserID().String()
		} else {
			continue
		}
		seen[opponentID]++
		if g.StartedAt() > lastPlayed[opponentID] {
			lastPlayed[opponentID] = g.StartedAt()
		}
	}

	var result []quick_duel.RecentOpponentEntry
	for idStr, count := range seen {
		id, err := shared.NewUserID(idStr)
		if err != nil {
			continue
		}
		result = append(result, quick_duel.RecentOpponentEntry{
			OpponentID:   id,
			GamesCount:   count,
			LastPlayedAt: lastPlayed[idStr],
		})
	}

	// Sort by LastPlayedAt desc
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].LastPlayedAt > result[i].LastPlayedAt {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// mockChallengeRepo is an in-memory challenge repository
type mockChallengeRepo struct {
	challenges map[string]*quick_duel.DuelChallenge
}

func newMockChallengeRepo() *mockChallengeRepo {
	return &mockChallengeRepo{challenges: make(map[string]*quick_duel.DuelChallenge)}
}

func (m *mockChallengeRepo) Save(c *quick_duel.DuelChallenge) error {
	m.challenges[c.ID().String()] = c
	return nil
}

func (m *mockChallengeRepo) FindByID(id quick_duel.ChallengeID) (*quick_duel.DuelChallenge, error) {
	if c, ok := m.challenges[id.String()]; ok {
		return c, nil
	}
	return nil, quick_duel.ErrChallengeNotFound
}

func (m *mockChallengeRepo) FindByLink(link string) (*quick_duel.DuelChallenge, error) {
	for _, c := range m.challenges {
		if c.ChallengeLink() == link {
			return c, nil
		}
	}
	return nil, quick_duel.ErrChallengeNotFound
}

func (m *mockChallengeRepo) FindByLinkCode(code string) (*quick_duel.DuelChallenge, error) {
	for _, c := range m.challenges {
		if c.ChallengeLink() == code {
			return c, nil
		}
	}
	return nil, quick_duel.ErrChallengeNotFound
}

func (m *mockChallengeRepo) FindPendingForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	var result []*quick_duel.DuelChallenge
	for _, c := range m.challenges {
		if c.Status() == quick_duel.ChallengeStatusPending {
			if c.ChallengedID() != nil && c.ChallengedID().Equals(playerID) {
				result = append(result, c)
			}
		}
	}
	return result, nil
}

func (m *mockChallengeRepo) FindPendingByChallenger(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	var result []*quick_duel.DuelChallenge
	for _, c := range m.challenges {
		if c.Status() == quick_duel.ChallengeStatusPending && c.ChallengerID().Equals(playerID) {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockChallengeRepo) Delete(id quick_duel.ChallengeID) error {
	delete(m.challenges, id.String())
	return nil
}

func (m *mockChallengeRepo) DeleteExpired(_ int64) error {
	return nil
}

// mockPlayerRatingRepo is an in-memory player rating repository
type mockPlayerRatingRepo struct {
	ratings map[string]*quick_duel.PlayerRating // keyed by "playerID:seasonID"
}

func newMockPlayerRatingRepo() *mockPlayerRatingRepo {
	return &mockPlayerRatingRepo{ratings: make(map[string]*quick_duel.PlayerRating)}
}

func (m *mockPlayerRatingRepo) key(playerID quick_duel.UserID, seasonID string) string {
	return playerID.String() + ":" + seasonID
}

func (m *mockPlayerRatingRepo) Save(rating *quick_duel.PlayerRating) error {
	m.ratings[m.key(rating.PlayerID(), rating.SeasonID())] = rating
	return nil
}

func (m *mockPlayerRatingRepo) FindByPlayerID(playerID quick_duel.UserID) (*quick_duel.PlayerRating, error) {
	for _, r := range m.ratings {
		if r.PlayerID().Equals(playerID) {
			return r, nil
		}
	}
	return nil, fmt.Errorf("player rating not found")
}

func (m *mockPlayerRatingRepo) FindOrCreate(playerID quick_duel.UserID, seasonID string, createdAt int64) (*quick_duel.PlayerRating, error) {
	k := m.key(playerID, seasonID)
	if r, ok := m.ratings[k]; ok {
		return r, nil
	}
	r := quick_duel.NewPlayerRating(playerID, seasonID, createdAt)
	m.ratings[k] = r
	return r, nil
}

func (m *mockPlayerRatingRepo) GetLeaderboard(seasonID string, limit int, _ int) ([]*quick_duel.PlayerRating, error) {
	var result []*quick_duel.PlayerRating
	for _, r := range m.ratings {
		if r.SeasonID() == seasonID {
			result = append(result, r)
		}
	}
	// Sort by MMR descending
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].MMR() > result[i].MMR() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockPlayerRatingRepo) GetFriendsLeaderboard(_ quick_duel.UserID, _ []quick_duel.UserID, _ int) ([]*quick_duel.PlayerRating, error) {
	return nil, nil
}

func (m *mockPlayerRatingRepo) GetPlayerRank(playerID quick_duel.UserID, seasonID string) (int, error) {
	var ratings []*quick_duel.PlayerRating
	for _, r := range m.ratings {
		if r.SeasonID() == seasonID {
			ratings = append(ratings, r)
		}
	}
	// Sort by MMR descending
	for i := 0; i < len(ratings); i++ {
		for j := i + 1; j < len(ratings); j++ {
			if ratings[j].MMR() > ratings[i].MMR() {
				ratings[i], ratings[j] = ratings[j], ratings[i]
			}
		}
	}
	for i, r := range ratings {
		if r.PlayerID().Equals(playerID) {
			return i + 1, nil
		}
	}
	return 0, nil
}

func (m *mockPlayerRatingRepo) GetTotalPlayers(seasonID string) (int, error) {
	count := 0
	for _, r := range m.ratings {
		if r.SeasonID() == seasonID {
			count++
		}
	}
	return count, nil
}

// mockMatchmakingQueue is an in-memory matchmaking queue
type mockMatchmakingQueue struct {
	queue map[string]struct {
		mmr      int
		joinedAt int64
	}
}

func newMockMatchmakingQueue() *mockMatchmakingQueue {
	return &mockMatchmakingQueue{
		queue: make(map[string]struct {
			mmr      int
			joinedAt int64
		}),
	}
}

func (m *mockMatchmakingQueue) AddToQueue(playerID quick_duel.UserID, mmr int, joinedAt int64) error {
	m.queue[playerID.String()] = struct {
		mmr      int
		joinedAt int64
	}{mmr: mmr, joinedAt: joinedAt}
	return nil
}

func (m *mockMatchmakingQueue) RemoveFromQueue(playerID quick_duel.UserID) error {
	delete(m.queue, playerID.String())
	return nil
}

func (m *mockMatchmakingQueue) FindMatch(_ quick_duel.UserID, _ int, _ int) (*quick_duel.UserID, *int, error) {
	return nil, nil, nil
}

func (m *mockMatchmakingQueue) GetQueueLength() (int, error) {
	return len(m.queue), nil
}

func (m *mockMatchmakingQueue) IsPlayerInQueue(playerID quick_duel.UserID) (bool, error) {
	_, ok := m.queue[playerID.String()]
	return ok, nil
}

func (m *mockMatchmakingQueue) GetPlayerQueueInfo(playerID quick_duel.UserID) (int64, int, error) {
	if info, ok := m.queue[playerID.String()]; ok {
		return info.joinedAt, info.mmr, nil
	}
	return 0, 0, fmt.Errorf("player not in queue")
}

// mockSeasonRepo is an in-memory season repository
type mockSeasonRepo struct {
	currentSeason string
	seasons       map[string]struct {
		startsAt int64
		endsAt   int64
	}
}

func newMockSeasonRepo() *mockSeasonRepo {
	return &mockSeasonRepo{
		currentSeason: "2026-02",
		seasons: map[string]struct {
			startsAt int64
			endsAt   int64
		}{
			"2026-02": {startsAt: 1738368000, endsAt: 1740787200},
		},
	}
}

func (m *mockSeasonRepo) GetCurrentSeason() (string, error) {
	return m.currentSeason, nil
}

func (m *mockSeasonRepo) CreateSeason(seasonID string, startsAt int64, endsAt int64) error {
	m.seasons[seasonID] = struct {
		startsAt int64
		endsAt   int64
	}{startsAt: startsAt, endsAt: endsAt}
	return nil
}

func (m *mockSeasonRepo) GetSeasonInfo(seasonID string) (int64, int64, error) {
	if s, ok := m.seasons[seasonID]; ok {
		return s.startsAt, s.endsAt, nil
	}
	return 0, 0, fmt.Errorf("season not found")
}

// mockReferralRepo is an in-memory referral repository
type mockReferralRepo struct {
	referrals map[string]*quick_duel.Referral // keyed by ID
}

func newMockReferralRepo() *mockReferralRepo {
	return &mockReferralRepo{referrals: make(map[string]*quick_duel.Referral)}
}

func (m *mockReferralRepo) Save(r *quick_duel.Referral) error {
	m.referrals[r.ID().String()] = r
	return nil
}

func (m *mockReferralRepo) FindByID(id quick_duel.ReferralID) (*quick_duel.Referral, error) {
	if r, ok := m.referrals[id.String()]; ok {
		return r, nil
	}
	return nil, quick_duel.ErrReferralNotFound
}

func (m *mockReferralRepo) FindByInviterAndInvitee(inviterID quick_duel.UserID, inviteeID quick_duel.UserID) (*quick_duel.Referral, error) {
	for _, r := range m.referrals {
		if r.InviterID().Equals(inviterID) && r.InviteeID().Equals(inviteeID) {
			return r, nil
		}
	}
	return nil, quick_duel.ErrReferralNotFound
}

func (m *mockReferralRepo) FindByInvitee(inviteeID quick_duel.UserID) (*quick_duel.Referral, error) {
	for _, r := range m.referrals {
		if r.InviteeID().Equals(inviteeID) {
			return r, nil
		}
	}
	return nil, quick_duel.ErrReferralNotFound
}

func (m *mockReferralRepo) FindByInviter(inviterID quick_duel.UserID) ([]*quick_duel.Referral, error) {
	var result []*quick_duel.Referral
	for _, r := range m.referrals {
		if r.InviterID().Equals(inviterID) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockReferralRepo) CountByInviter(inviterID quick_duel.UserID) (int, error) {
	count := 0
	for _, r := range m.referrals {
		if r.InviterID().Equals(inviterID) {
			count++
		}
	}
	return count, nil
}

func (m *mockReferralRepo) CountActiveByInviter(inviterID quick_duel.UserID) (int, error) {
	count := 0
	for _, r := range m.referrals {
		if r.InviterID().Equals(inviterID) && r.MilestonePlayedFive() {
			count++
		}
	}
	return count, nil
}

func (m *mockReferralRepo) GetReferralLeaderboard(limit int) ([]quick_duel.ReferralLeaderboardEntry, error) {
	// Count referrals by inviter
	counts := make(map[string]int)
	for _, r := range m.referrals {
		counts[r.InviterID().String()]++
	}

	var result []quick_duel.ReferralLeaderboardEntry
	for idStr, count := range counts {
		id, _ := shared.NewUserID(idStr)
		result = append(result, quick_duel.ReferralLeaderboardEntry{
			PlayerID:       id,
			Username:       "Player_" + idStr,
			TotalReferrals: count,
		})
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockReferralRepo) GetPlayerReferralRank(inviterID quick_duel.UserID) (int, error) {
	return 1, nil
}

// mockOnlineTracker is an in-memory online tracker
type mockOnlineTracker struct {
	online  map[string]bool
	inGame  map[string]string // playerID -> gameID
}

func newMockOnlineTracker() *mockOnlineTracker {
	return &mockOnlineTracker{
		online: make(map[string]bool),
		inGame: make(map[string]string),
	}
}

func (m *mockOnlineTracker) SetOnline(playerID string, _ int) error {
	m.online[playerID] = true
	return nil
}

func (m *mockOnlineTracker) IsOnline(playerID string) (bool, error) {
	return m.online[playerID], nil
}

func (m *mockOnlineTracker) GetOnlineFriends(_ string, friendIDs []string) ([]string, error) {
	var result []string
	for _, id := range friendIDs {
		if m.online[id] {
			result = append(result, id)
		}
	}
	return result, nil
}

func (m *mockOnlineTracker) SetInGame(playerID string, gameID string) error {
	m.inGame[playerID] = gameID
	return nil
}

func (m *mockOnlineTracker) ClearInGame(playerID string) error {
	delete(m.inGame, playerID)
	return nil
}

func (m *mockOnlineTracker) GetGameID(playerID string) (string, error) {
	return m.inGame[playerID], nil
}

// mockQuestionRepo is an in-memory question repository for duels
type mockQuestionRepo struct {
	questions []QuestionData
}

func newMockQuestionRepo() *mockQuestionRepo {
	qs := make([]QuestionData, 10)
	for i := 0; i < 10; i++ {
		qID := quiz.NewQuestionID().String()
		aIDs := [4]string{
			quiz.NewAnswerID().String(),
			quiz.NewAnswerID().String(),
			quiz.NewAnswerID().String(),
			quiz.NewAnswerID().String(),
		}
		qs[i] = QuestionData{
			ID:   qID,
			Text: fmt.Sprintf("Question %d", i+1),
			Answers: []AnswerData{
				{ID: aIDs[0], Text: "Correct", IsCorrect: true},
				{ID: aIDs[1], Text: "Wrong A", IsCorrect: false},
				{ID: aIDs[2], Text: "Wrong B", IsCorrect: false},
				{ID: aIDs[3], Text: "Wrong C", IsCorrect: false},
			},
		}
	}
	return &mockQuestionRepo{questions: qs}
}

func (m *mockQuestionRepo) FindRandomByDifficulty(count int, _ string) ([]QuestionData, error) {
	if count > len(m.questions) {
		return nil, fmt.Errorf("not enough questions: need %d, have %d", count, len(m.questions))
	}
	return m.questions[:count], nil
}

func (m *mockQuestionRepo) FindByID(questionID quiz.QuestionID) (*quiz.Question, error) {
	for _, qd := range m.questions {
		if qd.ID == questionID.String() {
			qText, _ := quiz.NewQuestionText(qd.Text)
			pts, _ := quiz.NewPoints(100)
			q, err := quiz.NewQuestion(questionID, qText, pts, 0)
			if err != nil {
				return nil, err
			}
			for i, ad := range qd.Answers {
				aID, err := quiz.NewAnswerIDFromString(ad.ID)
				if err != nil {
					return nil, err
				}
				aText, _ := quiz.NewAnswerText(ad.Text)
				ans, err := quiz.NewAnswer(aID, aText, ad.IsCorrect, i)
				if err != nil {
					return nil, err
				}
				if err := q.AddAnswer(*ans); err != nil {
					return nil, err
				}
			}
			return q, nil
		}
	}
	return nil, fmt.Errorf("question not found: %s", questionID.String())
}

// mockUserRepo for user lookups
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

// mockEventBus collects published events
type mockEventBus struct {
	events []quick_duel.Event
}

func (m *mockEventBus) Publish(event quick_duel.Event) {
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

type duelFixture struct {
	duelGameRepo     *mockDuelGameRepo
	challengeRepo    *mockChallengeRepo
	playerRatingRepo *mockPlayerRatingRepo
	matchmakingQueue *mockMatchmakingQueue
	seasonRepo       *mockSeasonRepo
	referralRepo     *mockReferralRepo
	onlineTracker    *mockOnlineTracker
	questionRepo     *mockQuestionRepo
	userRepo         *mockUserRepo
	eventBus         *mockEventBus
}

func setupFixture(t *testing.T) *duelFixture {
	t.Helper()

	userRepo := newMockUserRepo()
	userRepo.users[testPlayer1ID] = createTestUser(t, testPlayer1ID, "Player1")
	userRepo.users[testPlayer2ID] = createTestUser(t, testPlayer2ID, "Player2")
	userRepo.users[testPlayer3ID] = createTestUser(t, testPlayer3ID, "Player3")

	return &duelFixture{
		duelGameRepo:     newMockDuelGameRepo(),
		challengeRepo:    newMockChallengeRepo(),
		playerRatingRepo: newMockPlayerRatingRepo(),
		matchmakingQueue: newMockMatchmakingQueue(),
		seasonRepo:       newMockSeasonRepo(),
		referralRepo:     newMockReferralRepo(),
		onlineTracker:    newMockOnlineTracker(),
		questionRepo:     newMockQuestionRepo(),
		userRepo:         userRepo,
		eventBus:         &mockEventBus{events: make([]quick_duel.Event, 0)},
	}
}

// Use case constructors

func (f *duelFixture) newGetDuelStatusUC() *GetDuelStatusUseCase {
	return NewGetDuelStatusUseCase(
		f.playerRatingRepo, f.duelGameRepo, f.challengeRepo,
		f.seasonRepo, f.userRepo,
	)
}

func (f *duelFixture) newJoinQueueUC() *JoinQueueUseCase {
	return NewJoinQueueUseCase(
		f.matchmakingQueue, f.playerRatingRepo, f.duelGameRepo, f.seasonRepo,
	)
}

func (f *duelFixture) newLeaveQueueUC() *LeaveQueueUseCase {
	return NewLeaveQueueUseCase(f.matchmakingQueue)
}

func (f *duelFixture) newSendChallengeUC() *SendChallengeUseCase {
	return NewSendChallengeUseCase(f.challengeRepo, f.duelGameRepo, f.eventBus)
}

func (f *duelFixture) newRespondChallengeUC() *RespondChallengeUseCase {
	return NewRespondChallengeUseCase(
		f.challengeRepo, f.duelGameRepo, f.playerRatingRepo,
		f.seasonRepo, f.eventBus,
	)
}

func (f *duelFixture) newAcceptByLinkCodeUC() *AcceptByLinkCodeUseCase {
	return NewAcceptByLinkCodeUseCase(
		f.challengeRepo, f.userRepo, &noOpNotifier{}, f.eventBus,
	)
}

// noOpNotifier satisfies TelegramNotifier in tests
type noOpNotifier struct{}

func (n *noOpNotifier) NotifyChallengeAccepted(_ context.Context, _ int64, _ string, _ string) error {
	return nil
}
func (n *noOpNotifier) NotifyInviterWaiting(_ context.Context, _ int64, _ string, _ string) error {
	return nil
}

func (f *duelFixture) newCreateChallengeLinkUC() *CreateChallengeLinkUseCase {
	return NewCreateChallengeLinkUseCase(f.challengeRepo, f.eventBus)
}

func (f *duelFixture) newGetGameHistoryUC() *GetGameHistoryUseCase {
	return NewGetGameHistoryUseCase(f.duelGameRepo, f.userRepo)
}

func (f *duelFixture) newGetLeaderboardUC() *GetLeaderboardUseCase {
	return NewGetLeaderboardUseCase(
		f.playerRatingRepo, f.referralRepo, f.seasonRepo, f.userRepo,
	)
}

func (f *duelFixture) newStartGameUC() *StartGameUseCase {
	return NewStartGameUseCase(
		f.duelGameRepo, f.playerRatingRepo, f.questionRepo,
		f.seasonRepo, f.eventBus,
	)
}

func (f *duelFixture) newSubmitDuelAnswerUC() *SubmitDuelAnswerUseCase {
	return NewSubmitDuelAnswerUseCase(
		f.duelGameRepo, f.playerRatingRepo, f.questionRepo,
		f.seasonRepo, f.eventBus,
		NewMemoryRoundCache(),
	)
}

func (f *duelFixture) newRequestRematchUC() *RequestRematchUseCase {
	return NewRequestRematchUseCase(f.duelGameRepo, f.challengeRepo, f.eventBus)
}

func (f *duelFixture) newGetOnlineFriendsUC() *GetOnlineFriendsUseCase {
	return NewGetOnlineFriendsUseCase(f.onlineTracker, f.userRepo)
}

func (f *duelFixture) newGetRivalsUC() *GetRivalsUseCase {
	return NewGetRivalsUseCase(
		f.duelGameRepo, f.playerRatingRepo, f.userRepo, f.onlineTracker,
	)
}

// correctAnswerID returns the correct answer ID for the question at the given index.
func (f *duelFixture) correctAnswerID(questionIdx int) string {
	for _, a := range f.questionRepo.questions[questionIdx].Answers {
		if a.IsCorrect {
			return a.ID
		}
	}
	panic("no correct answer found")
}

// wrongAnswerID returns a wrong (incorrect) answer ID for the question at the given index.
func (f *duelFixture) wrongAnswerID(questionIdx int) string {
	for _, a := range f.questionRepo.questions[questionIdx].Answers {
		if !a.IsCorrect {
			return a.ID
		}
	}
	panic("no wrong answer found")
}

// questionIDs returns QuestionIDs from the mock question repo for building test games
func (f *duelFixture) questionIDs() []quick_duel.QuestionID {
	qIDs := make([]quick_duel.QuestionID, len(f.questionRepo.questions))
	for i, q := range f.questionRepo.questions {
		qid, _ := quiz.NewQuestionIDFromString(q.ID)
		qIDs[i] = qid
	}
	return qIDs
}

// startGame creates and starts a duel game between two players
func (f *duelFixture) startGame(t *testing.T, player1ID, player2ID string) StartGameOutput {
	t.Helper()
	uc := f.newStartGameUC()
	output, err := uc.Execute(StartGameInput{
		Player1ID:       player1ID,
		Player2ID:       player2ID,
		Player1Username: "Player1",
		Player2Username: "Player2",
	})
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}
	return output
}
