package quick_duel_test

import (
	"errors"
	"testing"

	quick_duel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

func TestDuelChallenge_TelegramMessageID(t *testing.T) {
	challengerID, _ := shared.NewUserID("111")
	challengedID, _ := shared.NewUserID("222")
	c, _ := quick_duel.NewDirectChallenge(challengerID, challengedID, 1000)

	if c.TelegramMessageID() != int64(0) {
		t.Errorf("expected 0, got %d", c.TelegramMessageID())
	}

	c.SetTelegramMessageID(42)
	if c.TelegramMessageID() != int64(42) {
		t.Errorf("expected 42, got %d", c.TelegramMessageID())
	}
}

func TestMarkStarted_TransitionsToAccepted(t *testing.T) {
	challengerID, _ := shared.NewUserID("user-challenger-001")
	now := int64(1700000000)

	challenge, err := quick_duel.NewLinkChallenge(challengerID, now)
	if err != nil {
		t.Fatal(err)
	}

	inviteeID, _ := shared.NewUserID("user-invitee-001")
	err = challenge.AcceptWaiting(inviteeID, "invitee_name", now+10)
	if err != nil {
		t.Fatal(err)
	}

	if challenge.Status() != quick_duel.ChallengeStatusAcceptedWaitingInviter {
		t.Fatalf("expected accepted_waiting_inviter, got %s", challenge.Status())
	}

	gameID := quick_duel.NewGameIDFromString("game-001")
	err = challenge.MarkStarted(gameID)
	if err != nil {
		t.Fatal(err)
	}

	if challenge.Status() != quick_duel.ChallengeStatusAccepted {
		t.Errorf("expected accepted, got %s", challenge.Status())
	}
	if challenge.MatchID() == nil || challenge.MatchID().String() != "game-001" {
		t.Errorf("expected matchID=game-001, got %v", challenge.MatchID())
	}
}

func TestMarkStarted_FailsIfNotWaitingInviter(t *testing.T) {
	challengerID, _ := shared.NewUserID("user-challenger-001")
	now := int64(1700000000)

	challenge, err := quick_duel.NewLinkChallenge(challengerID, now)
	if err != nil {
		t.Fatal(err)
	}

	gameID := quick_duel.NewGameIDFromString("game-001")
	err = challenge.MarkStarted(gameID)
	if !errors.Is(err, quick_duel.ErrChallengeNotPending) {
		t.Errorf("expected ErrChallengeNotPending, got %v", err)
	}
}
