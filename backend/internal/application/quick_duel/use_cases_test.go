package quick_duel

import (
	"errors"
	"testing"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// ========================================
// GetDuelStatus Tests
// ========================================

func TestGetDuelStatus_NewPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetDuelStatusUC()

	output, err := uc.Execute(GetDuelStatusInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.HasActiveDuel {
		t.Error("new player should not have active duel")
	}
	if output.ActiveGameID != nil {
		t.Error("new player should not have active game ID")
	}
	if output.Player.MMR != quick_duel.InitialMMR {
		t.Errorf("new player MMR = %d, want %d", output.Player.MMR, quick_duel.InitialMMR)
	}
	if output.SeasonID != "2026-02" {
		t.Errorf("SeasonID = %s, want 2026-02", output.SeasonID)
	}
	if len(output.PendingChallenges) != 0 {
		t.Errorf("PendingChallenges = %d, want 0", len(output.PendingChallenges))
	}
}

func TestGetDuelStatus_WithActiveGame(t *testing.T) {
	f := setupFixture(t)

	// Start a game
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newGetDuelStatusUC()
	output, err := uc.Execute(GetDuelStatusInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.HasActiveDuel {
		t.Error("should have active duel")
	}
	if output.ActiveGameID == nil {
		t.Fatal("ActiveGameID should not be nil")
	}
	if *output.ActiveGameID != gameOutput.GameID {
		t.Errorf("ActiveGameID = %s, want %s", *output.ActiveGameID, gameOutput.GameID)
	}
}

func TestGetDuelStatus_WithPendingChallenge(t *testing.T) {
	f := setupFixture(t)

	// Send challenge to player1
	now := time.Now().UTC().Unix()
	challenge, _ := quick_duel.NewDirectChallenge(mustUserID(testPlayer2ID), mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newGetDuelStatusUC()
	output, err := uc.Execute(GetDuelStatusInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.PendingChallenges) != 1 {
		t.Fatalf("PendingChallenges = %d, want 1", len(output.PendingChallenges))
	}
	if output.PendingChallenges[0].ChallengerID != testPlayer2ID {
		t.Errorf("ChallengerID = %s, want %s", output.PendingChallenges[0].ChallengerID, testPlayer2ID)
	}
}

func TestGetDuelStatus_IncludesOutgoingChallenges(t *testing.T) {
	f := setupFixture(t)

	// Create a link challenge sent by player1
	challenge, err := quick_duel.NewLinkChallenge(
		mustUserID(testPlayer1ID),
		time.Now().UTC().Unix(),
	)
	if err != nil {
		t.Fatalf("failed to create link challenge: %v", err)
	}
	f.challengeRepo.Save(challenge)

	uc := f.newGetDuelStatusUC()
	output, err := uc.Execute(GetDuelStatusInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.OutgoingChallenges) != 1 {
		t.Fatalf("OutgoingChallenges = %d, want 1", len(output.OutgoingChallenges))
	}
	if output.OutgoingChallenges[0].Type != "link" {
		t.Errorf("Type = %s, want link", output.OutgoingChallenges[0].Type)
	}
	if output.OutgoingChallenges[0].Status != "pending" {
		t.Errorf("Status = %s, want pending", output.OutgoingChallenges[0].Status)
	}
}

func TestGetDuelStatus_InvalidPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetDuelStatusUC()

	_, err := uc.Execute(GetDuelStatusInput{PlayerID: ""})
	if err == nil {
		t.Error("expected error for empty player ID")
	}
}

func TestGetDuelStatus_WithAcceptedChallenge(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Create link challenge and have player1 accept it as invitee
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer1ID), "Player1", now+10)
	f.challengeRepo.Save(challenge)

	uc := f.newGetDuelStatusUC()
	output, err := uc.Execute(GetDuelStatusInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.AcceptedChallenges) != 1 {
		t.Fatalf("AcceptedChallenges = %d, want 1", len(output.AcceptedChallenges))
	}
	if output.AcceptedChallenges[0].ChallengerID != testPlayer2ID {
		t.Errorf("ChallengerID = %s, want %s", output.AcceptedChallenges[0].ChallengerID, testPlayer2ID)
	}
}

// ========================================
// JoinQueue Tests
// ========================================

func TestJoinQueue_Success(t *testing.T) {
	f := setupFixture(t)
	uc := f.newJoinQueueUC()

	output, err := uc.Execute(JoinQueueInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Status != "searching" {
		t.Errorf("Status = %s, want searching", output.Status)
	}
	if output.Position != 1 {
		t.Errorf("Position = %d, want 1", output.Position)
	}

	// Verify player is in queue
	inQueue, _ := f.matchmakingQueue.IsPlayerInQueue(mustUserID(testPlayer1ID))
	if !inQueue {
		t.Error("player should be in queue after joining")
	}
}

func TestJoinQueue_AlreadyInQueue(t *testing.T) {
	f := setupFixture(t)
	uc := f.newJoinQueueUC()

	// Join once
	_, err := uc.Execute(JoinQueueInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("first join failed: %v", err)
	}

	// Try to join again
	_, err = uc.Execute(JoinQueueInput{PlayerID: testPlayer1ID})
	if err != quick_duel.ErrAlreadyInQueue {
		t.Errorf("expected ErrAlreadyInQueue, got %v", err)
	}
}

func TestJoinQueue_AlreadyInGame(t *testing.T) {
	f := setupFixture(t)

	// Start a game first
	f.startGame(t, testPlayer1ID, testPlayer2ID)

	// Try to join queue
	uc := f.newJoinQueueUC()
	_, err := uc.Execute(JoinQueueInput{PlayerID: testPlayer1ID})
	if err != quick_duel.ErrAlreadyInGame {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}

func TestJoinQueue_InvalidPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newJoinQueueUC()

	_, err := uc.Execute(JoinQueueInput{PlayerID: ""})
	if err == nil {
		t.Error("expected error for empty player ID")
	}
}

// ========================================
// LeaveQueue Tests
// ========================================

func TestLeaveQueue_Success(t *testing.T) {
	f := setupFixture(t)

	// Join queue first
	joinUC := f.newJoinQueueUC()
	_, err := joinUC.Execute(JoinQueueInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}

	// Leave queue
	uc := f.newLeaveQueueUC()
	output, err := uc.Execute(LeaveQueueInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("Success should be true")
	}
	if !output.TicketRefunded {
		t.Error("TicketRefunded should be true when leaving queue")
	}

	// Verify player is no longer in queue
	inQueue, _ := f.matchmakingQueue.IsPlayerInQueue(mustUserID(testPlayer1ID))
	if inQueue {
		t.Error("player should not be in queue after leaving")
	}
}

func TestLeaveQueue_NotInQueue(t *testing.T) {
	f := setupFixture(t)
	uc := f.newLeaveQueueUC()

	output, err := uc.Execute(LeaveQueueInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("Success should be true even when not in queue")
	}
	if output.TicketRefunded {
		t.Error("TicketRefunded should be false when not in queue")
	}
}

// ========================================
// SendChallenge Tests
// ========================================

func TestSendChallenge_Success(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSendChallengeUC()

	output, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ChallengeID == "" {
		t.Error("ChallengeID should not be empty")
	}
	if output.Status != "pending" {
		t.Errorf("Status = %s, want pending", output.Status)
	}
	if output.ExpiresIn != quick_duel.DirectChallengeExpirySeconds {
		t.Errorf("ExpiresIn = %d, want %d", output.ExpiresIn, quick_duel.DirectChallengeExpirySeconds)
	}
	if !output.TicketConsumed {
		t.Error("TicketConsumed should be true")
	}

	// Verify events published
	if len(f.eventBus.events) == 0 {
		t.Error("expected at least one event to be published")
	}
}

func TestSendChallenge_FriendInGame(t *testing.T) {
	f := setupFixture(t)

	// Friend is in a game
	f.startGame(t, testPlayer2ID, testPlayer3ID)

	uc := f.newSendChallengeUC()
	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != quick_duel.ErrFriendBusy {
		t.Errorf("expected ErrFriendBusy, got %v", err)
	}
}

func TestSendChallenge_DuplicateChallenge(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSendChallengeUC()

	// First challenge — must succeed
	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != nil {
		t.Fatalf("first challenge failed: %v", err)
	}

	// Second challenge to same friend — must fail
	_, err = uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != quick_duel.ErrChallengeAlreadySent {
		t.Errorf("expected ErrChallengeAlreadySent, got %v", err)
	}
}

func TestSendChallenge_InvalidPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSendChallengeUC()

	_, err := uc.Execute(SendChallengeInput{
		PlayerID: "",
		FriendID: testPlayer2ID,
	})
	if err == nil {
		t.Error("expected error for empty player ID")
	}
}

func TestSendChallenge_InvalidFriend(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSendChallengeUC()

	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: "",
	})
	if err == nil {
		t.Error("expected error for empty friend ID")
	}
}

func TestSendChallenge_FailsIfChallengerInGame(t *testing.T) {
	f := setupFixture(t)
	f.startGame(t, testPlayer1ID, testPlayer2ID) // player1 is now in active game

	uc := f.newSendChallengeUC()
	_, err := uc.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer3ID,
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}

// ========================================
// RespondChallenge Tests
// ========================================

func TestRespondChallenge_Accept(t *testing.T) {
	f := setupFixture(t)

	// Create a direct challenge
	now := time.Now().UTC().Unix()
	challenge, _ := quick_duel.NewDirectChallenge(mustUserID(testPlayer1ID), mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newRespondChallengeUC()
	output, err := uc.Execute(RespondChallengeInput{
		PlayerID:    testPlayer2ID,
		ChallengeID: challenge.ID().String(),
		Action:      "accept",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("Success should be true")
	}
	if output.TicketConsumed != true {
		t.Error("TicketConsumed should be true on accept")
	}
	if output.StartsIn == nil {
		t.Error("StartsIn should not be nil after accept")
	}
}

func TestRespondChallenge_Decline(t *testing.T) {
	f := setupFixture(t)

	now := time.Now().UTC().Unix()
	challenge, _ := quick_duel.NewDirectChallenge(mustUserID(testPlayer1ID), mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newRespondChallengeUC()
	output, err := uc.Execute(RespondChallengeInput{
		PlayerID:    testPlayer2ID,
		ChallengeID: challenge.ID().String(),
		Action:      "decline",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("Success should be true")
	}
	if output.TicketConsumed {
		t.Error("TicketConsumed should be false on decline")
	}

	// Verify events published
	if len(f.eventBus.events) == 0 {
		t.Error("expected decline event to be published")
	}
}

func TestRespondChallenge_NotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newRespondChallengeUC()

	_, err := uc.Execute(RespondChallengeInput{
		PlayerID:    testPlayer2ID,
		ChallengeID: quick_duel.NewChallengeID().String(),
		Action:      "accept",
	})
	if err != quick_duel.ErrChallengeNotFound {
		t.Errorf("expected ErrChallengeNotFound, got %v", err)
	}
}

func TestRespondChallenge_WrongPlayer(t *testing.T) {
	f := setupFixture(t)

	now := time.Now().UTC().Unix()
	challenge, _ := quick_duel.NewDirectChallenge(mustUserID(testPlayer1ID), mustUserID(testPlayer2ID), now)
	f.challengeRepo.Save(challenge)

	// Player3 tries to accept a challenge meant for Player2
	uc := f.newRespondChallengeUC()
	_, err := uc.Execute(RespondChallengeInput{
		PlayerID:    testPlayer3ID,
		ChallengeID: challenge.ID().String(),
		Action:      "accept",
	})
	if err != quick_duel.ErrNotChallengedPlayer {
		t.Errorf("expected ErrNotChallengedPlayer, got %v", err)
	}
}

// ========================================
// CreateChallengeLink Tests
// ========================================

func TestCreateChallengeLink_Success(t *testing.T) {
	f := setupFixture(t)
	uc := f.newCreateChallengeLinkUC()

	output, err := uc.Execute(CreateChallengeLinkInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ChallengeLink == "" {
		t.Error("ChallengeLink should not be empty")
	}
	if output.ExpiresAt == 0 {
		t.Error("ExpiresAt should not be 0")
	}
	if output.ShareText == "" {
		t.Error("ShareText should not be empty")
	}

	// Verify challenge saved
	if len(f.challengeRepo.challenges) != 1 {
		t.Errorf("expected 1 challenge saved, got %d", len(f.challengeRepo.challenges))
	}
}

func TestCreateChallengeLink_InvalidPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newCreateChallengeLinkUC()

	_, err := uc.Execute(CreateChallengeLinkInput{PlayerID: ""})
	if err == nil {
		t.Error("expected error for empty player ID")
	}
}

// ========================================
// AcceptByLinkCode Tests
// ========================================

func TestAcceptByLinkCode_Success(t *testing.T) {
	f := setupFixture(t)

	// Create a link challenge
	createUC := f.newCreateChallengeLinkUC()
	createOutput, err := createUC.Execute(CreateChallengeLinkInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Accept by link code
	uc := f.newAcceptByLinkCodeUC()
	output, err := uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: createOutput.ChallengeLink,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("Success should be true")
	}
	if output.ChallengeID == "" {
		t.Error("ChallengeID should not be empty")
	}
	if output.Status != "accepted_waiting_inviter" {
		t.Errorf("Status = %s, want accepted_waiting_inviter", output.Status)
	}
}

func TestAcceptByLinkCode_SelfAccept(t *testing.T) {
	f := setupFixture(t)

	// Create a link challenge
	createUC := f.newCreateChallengeLinkUC()
	createOutput, err := createUC.Execute(CreateChallengeLinkInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Try to accept own link
	uc := f.newAcceptByLinkCodeUC()
	_, err = uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer1ID,
		LinkCode: createOutput.ChallengeLink,
	})
	if err != quick_duel.ErrCannotChallengeSelf {
		t.Errorf("expected ErrCannotChallengeSelf, got %v", err)
	}
}

func TestAcceptByLinkCode_NotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newAcceptByLinkCodeUC()

	_, err := uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: "nonexistent-code",
	})
	if err != quick_duel.ErrChallengeNotFound {
		t.Errorf("expected ErrChallengeNotFound, got %v", err)
	}
}

func TestAcceptByLinkCode_FailsIfInviteeInGame(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Create a link challenge from player1
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)

	// player2 is already in an active game
	f.startGame(t, testPlayer2ID, testPlayer3ID)

	uc := f.newAcceptByLinkCodeUC()
	_, err := uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: challenge.ChallengeLink(),
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}

func TestAcceptByLinkCode_ReturnsInviterName(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// player1 (username "Player1") creates link challenge
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)

	uc := f.newAcceptByLinkCodeUC()
	output, err := uc.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: challenge.ChallengeLink(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.InviterName != "Player1" {
		t.Errorf("InviterName = %q, want %q", output.InviterName, "Player1")
	}
}

// ========================================
// StartGame Tests
// ========================================

func TestStartGame_Success(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartGameUC()

	output, err := uc.Execute(StartGameInput{
		Player1ID:       testPlayer1ID,
		Player2ID:       testPlayer2ID,
		Player1Username: "Player1",
		Player2Username: "Player2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.GameID == "" {
		t.Error("GameID should not be empty")
	}
	if output.Player1ID != testPlayer1ID {
		t.Errorf("Player1ID = %s, want %s", output.Player1ID, testPlayer1ID)
	}
	if output.Player2ID != testPlayer2ID {
		t.Errorf("Player2ID = %s, want %s", output.Player2ID, testPlayer2ID)
	}

	// Verify game saved
	if len(f.duelGameRepo.games) != 1 {
		t.Errorf("expected 1 game saved, got %d", len(f.duelGameRepo.games))
	}
}

func TestStartGame_InvalidPlayer1(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartGameUC()

	_, err := uc.Execute(StartGameInput{
		Player1ID:       "",
		Player2ID:       testPlayer2ID,
		Player1Username: "Player1",
		Player2Username: "Player2",
	})
	if err == nil {
		t.Error("expected error for empty player1 ID")
	}
}

func TestStartGame_InvalidPlayer2(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartGameUC()

	_, err := uc.Execute(StartGameInput{
		Player1ID:       testPlayer1ID,
		Player2ID:       "",
		Player1Username: "Player1",
		Player2Username: "Player2",
	})
	if err == nil {
		t.Error("expected error for empty player2 ID")
	}
}

// ========================================
// GetDomainPlayerOrder Tests
// ========================================

func TestGetDomainPlayerOrder_ReturnsCorrectOrder(t *testing.T) {
	f := setupFixture(t)

	// Start game: player1=testPlayer1ID (challenger), player2=testPlayer2ID (accepter)
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newStartGameUC()
	p1ID, p2ID, err := uc.GetDomainPlayerOrder(gameOutput.GameID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p1ID != testPlayer1ID {
		t.Errorf("Player1ID = %s, want %s", p1ID, testPlayer1ID)
	}
	if p2ID != testPlayer2ID {
		t.Errorf("Player2ID = %s, want %s", p2ID, testPlayer2ID)
	}
}

func TestGetDomainPlayerOrder_GameNotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newStartGameUC()

	_, _, err := uc.GetDomainPlayerOrder("nonexistent-game-id")
	if err == nil {
		t.Error("expected error for nonexistent game, got nil")
	}
}

// ========================================
// SubmitDuelAnswer Tests
// ========================================

func TestSubmitDuelAnswer_Success(t *testing.T) {
	f := setupFixture(t)

	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newSubmitDuelAnswerUC()
	output, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer1ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.correctAnswerID(0),
		TimeTaken: 2000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.IsCorrect {
		t.Error("IsCorrect should be true for correct answer")
	}
	if output.PointsEarned == 0 {
		t.Error("PointsEarned should not be 0 for correct answer")
	}
	if output.RoundComplete {
		t.Error("RoundComplete should be false - only one player answered")
	}
	if output.GameComplete {
		t.Error("GameComplete should be false")
	}
}

func TestSubmitDuelAnswer_RoundComplete(t *testing.T) {
	f := setupFixture(t)

	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newSubmitDuelAnswerUC()

	// Player1 answers correctly
	_, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer1ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.correctAnswerID(0),
		TimeTaken: 2000,
	})
	if err != nil {
		t.Fatalf("player1 answer error: %v", err)
	}

	// Player2 answers (wrong)
	output, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer2ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.wrongAnswerID(0),
		TimeTaken: 3000,
	})
	if err != nil {
		t.Fatalf("player2 answer error: %v", err)
	}

	if !output.RoundComplete {
		t.Error("RoundComplete should be true after both players answered")
	}
}

func TestSubmitDuelAnswer_GameNotFound(t *testing.T) {
	f := setupFixture(t)
	uc := f.newSubmitDuelAnswerUC()

	_, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:   testPlayer1ID,
		GameID:     quick_duel.NewGameID().String(),
		QuestionID: "q1",
		AnswerID:   "a1",
		TimeTaken:  2000,
	})
	if err != quick_duel.ErrGameNotFound {
		t.Errorf("expected ErrGameNotFound, got %v", err)
	}
}

func TestSubmitDuelAnswer_PlayerNotInGame(t *testing.T) {
	f := setupFixture(t)

	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newSubmitDuelAnswerUC()
	_, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer3ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.correctAnswerID(0),
		TimeTaken: 2000,
	})
	if err != quick_duel.ErrGameNotFound {
		t.Errorf("expected ErrGameNotFound, got %v", err)
	}
}

func TestSubmitDuelAnswer_BothPlayersAnswer(t *testing.T) {
	f := setupFixture(t)

	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newSubmitDuelAnswerUC()

	// Player1 answers correctly
	out1, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer1ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.correctAnswerID(0),
		TimeTaken: 2000,
	})
	if err != nil {
		t.Fatalf("player1 error: %v", err)
	}
	if out1.RoundComplete {
		t.Error("RoundComplete should be false after one player answers")
	}

	// Player2 answers (wrong)
	out2, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer2ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.wrongAnswerID(0),
		TimeTaken: 3000,
	})
	if err != nil {
		t.Fatalf("player2 error: %v", err)
	}
	if !out2.RoundComplete {
		t.Error("RoundComplete should be true after both players answer")
	}

	// Player1 answered correctly so score > 0; player2 answered wrong so 0
	if out2.Player1Score == 0 {
		t.Error("Player1Score should be > 0 after a correct answer")
	}
	if out2.Player2Score != 0 {
		t.Error("Player2Score should be 0 after a wrong answer")
	}
}

// ========================================
func TestSubmitDuelAnswer_WrongAnswer(t *testing.T) {
	f := setupFixture(t)

	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newSubmitDuelAnswerUC()
	output, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer1ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.wrongAnswerID(0),
		TimeTaken: 2000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.IsCorrect {
		t.Error("IsCorrect should be false for a wrong answer")
	}
	if output.PointsEarned != 0 {
		t.Errorf("PointsEarned should be 0 for a wrong answer, got %d", output.PointsEarned)
	}
	if output.CorrectAnswerID == "" {
		t.Error("CorrectAnswerID should be non-empty so client can show the correct answer")
	}
	if output.CorrectAnswerID == f.wrongAnswerID(0) {
		t.Error("CorrectAnswerID should differ from the submitted wrong answer ID")
	}
}

// ========================================
// Score Ordering Regression Tests
// ========================================

// TestScoreOrdering_AccepterAnswersFirst verifies that when the accepter (domain player2)
// answers correctly and the challenger (domain player1) answers wrong,
// player1Score reflects domain player1's score (not the accepter's).
// This catches the bug where hub WS-slot order could shadow domain player order.
func TestScoreOrdering_AccepterAnswersFirst(t *testing.T) {
	f := setupFixture(t)

	// Challenger = player1 (domain), Accepter = player2 (domain)
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)
	uc := f.newSubmitDuelAnswerUC()

	// Accepter (domain player2) answers CORRECTLY — simulates accepter connecting to WS first
	out, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer2ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.correctAnswerID(0),
		TimeTaken: 2000,
	})
	if err != nil {
		t.Fatalf("accepter answer error: %v", err)
	}

	// Player2Score (accepter = domain player2) should be > 0
	if out.Player2Score == 0 {
		t.Error("accepter answered correctly: Player2Score should be > 0")
	}
	// Player1Score (challenger = domain player1) should still be 0
	if out.Player1Score != 0 {
		t.Errorf("challenger hasn't answered: Player1Score should be 0, got %d", out.Player1Score)
	}

	// Now challenger (domain player1) answers WRONG
	out2, err := uc.Execute(SubmitDuelAnswerInput{
		PlayerID:  testPlayer1ID,
		GameID:    gameOutput.GameID,
		AnswerID:  f.wrongAnswerID(0),
		TimeTaken: 3000,
	})
	if err != nil {
		t.Fatalf("challenger answer error: %v", err)
	}

	// After both answered: Player1Score = challenger (0, wrong), Player2Score = accepter (> 0, correct)
	if out2.Player1Score != 0 {
		t.Errorf("challenger answered wrong: Player1Score should be 0, got %d", out2.Player1Score)
	}
	if out2.Player2Score == 0 {
		t.Error("accepter answered correctly: Player2Score should be > 0")
	}
}

// ========================================
// RequestRematch Tests
// ========================================

func TestRequestRematch_Success(t *testing.T) {
	f := setupFixture(t)

	// Create a finished game
	p1 := quick_duel.NewDuelPlayer(mustUserID(testPlayer1ID), "Player1", quick_duel.NewEloRating())
	p2 := quick_duel.NewDuelPlayer(mustUserID(testPlayer2ID), "Player2", quick_duel.NewEloRating())
	now := time.Now().UTC().Unix()

	qIDs := f.questionIDs()[:quick_duel.QuestionsPerDuel]

	game := quick_duel.ReconstructDuelGame(
		quick_duel.NewGameID(), p1, p2, qIDs,
		quick_duel.QuestionsPerDuel, quick_duel.GameStatusFinished,
		nil, now-60, now-10,
	)
	f.duelGameRepo.Save(game)

	uc := f.newRequestRematchUC()
	output, err := uc.Execute(RequestRematchInput{
		PlayerID: testPlayer1ID,
		GameID:   game.ID().String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.RematchID == "" {
		t.Error("RematchID should not be empty")
	}
	if output.Status != "pending" {
		t.Errorf("Status = %s, want pending", output.Status)
	}
	if output.ExpiresIn != quick_duel.DirectChallengeExpirySeconds {
		t.Errorf("ExpiresIn = %d, want %d", output.ExpiresIn, quick_duel.DirectChallengeExpirySeconds)
	}
}

func TestRequestRematch_GameNotFinished(t *testing.T) {
	f := setupFixture(t)

	// Create an in-progress game
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	uc := f.newRequestRematchUC()
	_, err := uc.Execute(RequestRematchInput{
		PlayerID: testPlayer1ID,
		GameID:   gameOutput.GameID,
	})
	if err != quick_duel.ErrGameNotActive {
		t.Errorf("expected ErrGameNotActive, got %v", err)
	}
}

func TestRequestRematch_PlayerNotInGame(t *testing.T) {
	f := setupFixture(t)

	// Create a finished game between player1 and player2
	p1 := quick_duel.NewDuelPlayer(mustUserID(testPlayer1ID), "Player1", quick_duel.NewEloRating())
	p2 := quick_duel.NewDuelPlayer(mustUserID(testPlayer2ID), "Player2", quick_duel.NewEloRating())
	now := time.Now().UTC().Unix()

	qIDs := f.questionIDs()[:quick_duel.QuestionsPerDuel]

	game := quick_duel.ReconstructDuelGame(
		quick_duel.NewGameID(), p1, p2, qIDs,
		quick_duel.QuestionsPerDuel, quick_duel.GameStatusFinished,
		nil, now-60, now-10,
	)
	f.duelGameRepo.Save(game)

	// Player3 tries to request rematch
	uc := f.newRequestRematchUC()
	_, err := uc.Execute(RequestRematchInput{
		PlayerID: testPlayer3ID,
		GameID:   game.ID().String(),
	})
	if err != quick_duel.ErrGameNotFound {
		t.Errorf("expected ErrGameNotFound, got %v", err)
	}
}

// ========================================
// GetGameHistory Tests
// ========================================

func TestGetGameHistory_Empty(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetGameHistoryUC()

	output, err := uc.Execute(GetGameHistoryInput{
		PlayerID: testPlayer1ID,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Games) != 0 {
		t.Errorf("expected 0 games, got %d", len(output.Games))
	}
	if output.Total != 0 {
		t.Errorf("Total = %d, want 0", output.Total)
	}
	if output.HasMore {
		t.Error("HasMore should be false")
	}
}

func TestGetGameHistory_WithGames(t *testing.T) {
	f := setupFixture(t)

	// Create a finished game
	p1 := quick_duel.NewDuelPlayer(mustUserID(testPlayer1ID), "Player1", quick_duel.NewEloRating())
	p2 := quick_duel.NewDuelPlayer(mustUserID(testPlayer2ID), "Player2", quick_duel.NewEloRating())
	now := time.Now().UTC().Unix()

	qIDs := f.questionIDs()[:quick_duel.QuestionsPerDuel]

	game := quick_duel.ReconstructDuelGame(
		quick_duel.NewGameID(), p1, p2, qIDs,
		quick_duel.QuestionsPerDuel, quick_duel.GameStatusFinished,
		nil, now-60, now-10,
	)
	f.duelGameRepo.Save(game)

	uc := f.newGetGameHistoryUC()
	output, err := uc.Execute(GetGameHistoryInput{
		PlayerID: testPlayer1ID,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(output.Games))
	}
	if output.Games[0].Opponent != "Player2" {
		t.Errorf("Opponent = %s, want Player2", output.Games[0].Opponent)
	}
}

func TestGetGameHistory_LimitClamping(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetGameHistoryUC()

	// Default limit (0 → 20)
	output, err := uc.Execute(GetGameHistoryInput{
		PlayerID: testPlayer1ID,
		Limit:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = output // Just verify no error

	// Over-limit (200 → 100)
	output, err = uc.Execute(GetGameHistoryInput{
		PlayerID: testPlayer1ID,
		Limit:    200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = output
}

func TestGetGameHistory_InvalidPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetGameHistoryUC()

	_, err := uc.Execute(GetGameHistoryInput{
		PlayerID: "",
		Limit:    20,
	})
	if err == nil {
		t.Error("expected error for empty player ID")
	}
}

// ========================================
// GetLeaderboard Tests
// ========================================

func TestGetLeaderboard_Seasonal_Empty(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	output, err := uc.Execute(GetLeaderboardInput{
		PlayerID: testPlayer1ID,
		Type:     "seasonal",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Type != "seasonal" {
		t.Errorf("Type = %s, want seasonal", output.Type)
	}
	if output.SeasonID != "2026-02" {
		t.Errorf("SeasonID = %s, want 2026-02", output.SeasonID)
	}
	if len(output.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(output.Entries))
	}
}

func TestGetLeaderboard_Seasonal_WithPlayers(t *testing.T) {
	f := setupFixture(t)

	// Create some ratings
	now := time.Now().UTC().Unix()
	f.playerRatingRepo.FindOrCreate(mustUserID(testPlayer1ID), "2026-02", now)
	f.playerRatingRepo.FindOrCreate(mustUserID(testPlayer2ID), "2026-02", now)

	uc := f.newGetLeaderboardUC()
	output, err := uc.Execute(GetLeaderboardInput{
		PlayerID: testPlayer1ID,
		Type:     "seasonal",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(output.Entries))
	}
}

func TestGetLeaderboard_LimitClamping(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	// Default limit (0 → 10)
	output, err := uc.Execute(GetLeaderboardInput{
		PlayerID: testPlayer1ID,
		Type:     "seasonal",
		Limit:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = output

	// Over-limit (200 → 100)
	output, err = uc.Execute(GetLeaderboardInput{
		PlayerID: testPlayer1ID,
		Type:     "seasonal",
		Limit:    200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = output
}

func TestGetLeaderboard_Friends_Empty(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	output, err := uc.Execute(GetLeaderboardInput{
		PlayerID: testPlayer1ID,
		Type:     "friends",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Entries) != 0 {
		t.Errorf("expected 0 entries for friends, got %d", len(output.Entries))
	}
}

func TestGetLeaderboard_Referrals(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	output, err := uc.Execute(GetLeaderboardInput{
		PlayerID: testPlayer1ID,
		Type:     "referrals",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Type != "referrals" {
		t.Errorf("Type = %s, want referrals", output.Type)
	}
}

func TestGetLeaderboard_InvalidPlayer(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetLeaderboardUC()

	_, err := uc.Execute(GetLeaderboardInput{
		PlayerID: "",
		Type:     "seasonal",
		Limit:    10,
	})
	if err == nil {
		t.Error("expected error for empty player ID")
	}
}

// ========================================
// GetOnlineFriends Tests
// ========================================

func TestGetOnlineFriends_NoFriends(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetOnlineFriendsUC()

	output, err := uc.Execute(GetOnlineFriendsInput{
		PlayerID:  testPlayer1ID,
		FriendIDs: []string{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.OnlineFriends) != 0 {
		t.Errorf("expected 0 friends, got %d", len(output.OnlineFriends))
	}
}

func TestGetOnlineFriends_SomeOnline(t *testing.T) {
	f := setupFixture(t)

	// Set player2 online, player3 offline
	f.onlineTracker.SetOnline(testPlayer2ID, 300)

	uc := f.newGetOnlineFriendsUC()
	output, err := uc.Execute(GetOnlineFriendsInput{
		PlayerID:  testPlayer1ID,
		FriendIDs: []string{testPlayer2ID, testPlayer3ID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.OnlineFriends) != 1 {
		t.Fatalf("expected 1 online friend, got %d", len(output.OnlineFriends))
	}
	if output.OnlineFriends[0].ID != testPlayer2ID {
		t.Errorf("OnlineFriend ID = %s, want %s", output.OnlineFriends[0].ID, testPlayer2ID)
	}
	if !output.OnlineFriends[0].IsOnline {
		t.Error("friend should be marked online")
	}
}

func TestGetOnlineFriends_InGame(t *testing.T) {
	f := setupFixture(t)

	// Player2 is online and in a game
	f.onlineTracker.SetOnline(testPlayer2ID, 300)
	f.onlineTracker.SetInGame(testPlayer2ID, "some-game-id")

	uc := f.newGetOnlineFriendsUC()
	output, err := uc.Execute(GetOnlineFriendsInput{
		PlayerID:  testPlayer1ID,
		FriendIDs: []string{testPlayer2ID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.OnlineFriends) != 1 {
		t.Fatalf("expected 1 online friend, got %d", len(output.OnlineFriends))
	}
	if !output.OnlineFriends[0].InGame {
		t.Error("friend should be marked as in game")
	}
}

// ========================================
// Integration / Full Flow Tests
// ========================================

func TestFullFlow_ChallengeAcceptAndPlay(t *testing.T) {
	f := setupFixture(t)

	// 1. Player1 sends challenge to Player2
	sendUC := f.newSendChallengeUC()
	sendOutput, err := sendUC.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != nil {
		t.Fatalf("send challenge error: %v", err)
	}

	// 2. Player2 sees the pending challenge in status
	statusUC := f.newGetDuelStatusUC()
	statusOutput, err := statusUC.Execute(GetDuelStatusInput{PlayerID: testPlayer2ID})
	if err != nil {
		t.Fatalf("get status error: %v", err)
	}
	if len(statusOutput.PendingChallenges) != 1 {
		t.Fatalf("expected 1 pending challenge, got %d", len(statusOutput.PendingChallenges))
	}

	// 3. Player2 accepts the challenge
	respondUC := f.newRespondChallengeUC()
	respondOutput, err := respondUC.Execute(RespondChallengeInput{
		PlayerID:    testPlayer2ID,
		ChallengeID: sendOutput.ChallengeID,
		Action:      "accept",
	})
	if err != nil {
		t.Fatalf("respond challenge error: %v", err)
	}
	if !respondOutput.Success {
		t.Error("respond should succeed")
	}
	if respondOutput.GameID == nil {
		t.Fatalf("respond should return GameID")
	}

	// 4. Play all rounds using the game created by RespondChallenge
	gameID := *respondOutput.GameID
	answerUC := f.newSubmitDuelAnswerUC()

	for round := 0; round < quick_duel.QuestionsPerDuel; round++ {
		_, err := answerUC.Execute(SubmitDuelAnswerInput{
			PlayerID:  testPlayer1ID,
			GameID:    gameID,
			AnswerID:  f.correctAnswerID(round),
			TimeTaken: 2000,
		})
		if err != nil {
			t.Fatalf("round %d p1 error: %v", round+1, err)
		}

		_, err = answerUC.Execute(SubmitDuelAnswerInput{
			PlayerID:  testPlayer2ID,
			GameID:    gameID,
			AnswerID:  f.wrongAnswerID(round),
			TimeTaken: 3000,
		})
		if err != nil {
			t.Fatalf("round %d p2 error: %v", round+1, err)
		}
	}

	// 6. Check game history
	historyUC := f.newGetGameHistoryUC()
	historyOutput, err := historyUC.Execute(GetGameHistoryInput{
		PlayerID: testPlayer1ID,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("get history error: %v", err)
	}
	if len(historyOutput.Games) != 1 {
		t.Errorf("expected 1 game in history, got %d", len(historyOutput.Games))
	}
}

func TestFullFlow_LinkChallengeAccept(t *testing.T) {
	f := setupFixture(t)

	// 1. Player1 creates a challenge link
	createLinkUC := f.newCreateChallengeLinkUC()
	linkOutput, err := createLinkUC.Execute(CreateChallengeLinkInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("create link error: %v", err)
	}

	// 2. Player2 accepts via link code
	acceptUC := f.newAcceptByLinkCodeUC()
	acceptOutput, err := acceptUC.Execute(AcceptByLinkCodeInput{
		PlayerID: testPlayer2ID,
		LinkCode: linkOutput.ChallengeLink,
	})
	if err != nil {
		t.Fatalf("accept by link error: %v", err)
	}

	if !acceptOutput.Success {
		t.Error("accept should succeed")
	}
	if acceptOutput.ChallengeID == "" {
		t.Error("ChallengeID should not be empty after accept")
	}
}

// ========================================
// GetRivals Tests
// ========================================

func TestGetRivals_EmptyWhenNoGames(t *testing.T) {
	f := setupFixture(t)
	uc := f.newGetRivalsUC()

	output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Rivals) != 0 {
		t.Errorf("Rivals = %d, want 0", len(output.Rivals))
	}
}

func TestGetRivals_ReturnsOpponentsFromCompletedGames(t *testing.T) {
	f := setupFixture(t)

	// Play a game between player1 and player2
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)

	// Force-complete the game by submitting all answers
	submitUC := f.newSubmitDuelAnswerUC()
	for i := 0; i < 7; i++ {
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer1ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  3000,
		})
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer2ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  4000,
		})
	}

	uc := f.newGetRivalsUC()
	output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Rivals) != 1 {
		t.Fatalf("Rivals = %d, want 1", len(output.Rivals))
	}
	if output.Rivals[0].ID != testPlayer2ID {
		t.Errorf("Rivals[0].ID = %s, want %s", output.Rivals[0].ID, testPlayer2ID)
	}
	if output.Rivals[0].GamesCount != 1 {
		t.Errorf("Rivals[0].GamesCount = %d, want 1", output.Rivals[0].GamesCount)
	}
}

// ========================================
// GetRivals Tests
// ========================================

func TestGetRivals_HasPendingChallenge(t *testing.T) {
	f := setupFixture(t)

	// Complete a game between player1 and player2 so player2 appears as rival
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)
	submitUC := f.newSubmitDuelAnswerUC()
	for i := 0; i < 7; i++ {
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer1ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  3000,
		})
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer2ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  4000,
		})
	}

	// Player1 sends challenge to player2
	sendUC := f.newSendChallengeUC()
	_, err := sendUC.Execute(SendChallengeInput{
		PlayerID: testPlayer1ID,
		FriendID: testPlayer2ID,
	})
	if err != nil {
		t.Fatalf("sendChallenge failed: %v", err)
	}

	uc := f.newGetRivalsUC()
	output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Rivals) == 0 {
		t.Fatal("expected at least one rival")
	}

	rival := output.Rivals[0]
	if !rival.HasPendingChallenge {
		t.Error("expected HasPendingChallenge=true for rival with pending challenge")
	}
}

func TestGetRivals_NoPendingChallenge(t *testing.T) {
	f := setupFixture(t)

	// Complete a game between player1 and player2 so player2 appears as rival
	gameOutput := f.startGame(t, testPlayer1ID, testPlayer2ID)
	submitUC := f.newSubmitDuelAnswerUC()
	for i := 0; i < 7; i++ {
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer1ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  3000,
		})
		submitUC.Execute(SubmitDuelAnswerInput{
			PlayerID:   testPlayer2ID,
			GameID:     gameOutput.GameID,
			QuestionID: f.questionRepo.questions[i].ID,
			AnswerID:   f.correctAnswerID(i),
			TimeTaken:  4000,
		})
	}

	uc := f.newGetRivalsUC()
	output, err := uc.Execute(GetRivalsInput{PlayerID: testPlayer1ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Rivals) == 0 {
		t.Fatal("expected at least one rival")
	}

	if output.Rivals[0].HasPendingChallenge {
		t.Error("expected HasPendingChallenge=false when no challenge sent")
	}
}

// ========================================
// StartChallenge Tests
// ========================================

func TestStartChallenge_FailsIfInviterInGame(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Create link challenge and have invitee accept
	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer2ID), "Player2", now+10)
	f.challengeRepo.Save(challenge)

	// inviter (player1) joins another game
	f.startGame(t, testPlayer1ID, testPlayer3ID)

	uc := f.newStartChallengeUC()
	_, err := uc.Execute(StartChallengeInput{
		PlayerID:    testPlayer1ID,
		ChallengeID: challenge.ID().String(),
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}

func TestStartChallenge_FailsIfInviteeInGame(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	challenge, _ := quick_duel.NewLinkChallenge(mustUserID(testPlayer1ID), now)
	f.challengeRepo.Save(challenge)
	_ = challenge.AcceptWaiting(mustUserID(testPlayer2ID), "Player2", now+10)
	f.challengeRepo.Save(challenge)

	// invitee (player2) joins another game
	f.startGame(t, testPlayer2ID, testPlayer3ID)

	uc := f.newStartChallengeUC()
	_, err := uc.Execute(StartChallengeInput{
		PlayerID:    testPlayer1ID,
		ChallengeID: challenge.ID().String(),
	})
	if !errors.Is(err, quick_duel.ErrAlreadyInGame) {
		t.Errorf("expected ErrAlreadyInGame, got %v", err)
	}
}
