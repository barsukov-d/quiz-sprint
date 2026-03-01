package telegram_test

import (
	"context"
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/telegram"
)

func TestNoOpNotifier_DoesNotError(t *testing.T) {
	n := telegram.NewNoOpNotifier()
	err := n.NotifyChallengeAccepted(context.Background(), 123456, "@friend", "https://t.me/bot")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
