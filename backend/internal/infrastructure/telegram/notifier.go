package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TelegramNotifier sends notifications via Telegram Bot API.
type TelegramNotifier interface {
	NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error
	NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error
}

// NoOpNotifier does nothing (used in tests / when bot token is absent).
type NoOpNotifier struct{}

func NewNoOpNotifier() TelegramNotifier { return &NoOpNotifier{} }

func (n *NoOpNotifier) NotifyChallengeAccepted(_ context.Context, _ int64, _ string, _ string) error {
	return nil
}
func (n *NoOpNotifier) NotifyInviterWaiting(_ context.Context, _ int64, _ string, _ string) error {
	return nil
}

// HTTPNotifier sends real Telegram messages.
type HTTPNotifier struct {
	botToken string
	client   *http.Client
}

func NewHTTPNotifier(botToken string) TelegramNotifier {
	return &HTTPNotifier{botToken: botToken, client: &http.Client{}}
}

type sendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func (n *HTTPNotifier) sendMessage(ctx context.Context, chatID int64, text string) error {
	body, _ := json.Marshal(sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (n *HTTPNotifier) NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error {
	text := fmt.Sprintf("⚔️ <b>%s</b> принял твой вызов и готов к дуэли!\n\n<a href=\"%s\">Зайти в лобби →</a>", inviteeName, lobbyURL)
	return n.sendMessage(ctx, inviterTelegramID, text)
}

func (n *HTTPNotifier) NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error {
	text := fmt.Sprintf("⚔️ <b>%s</b> ждёт тебя в лобби!\n\n<a href=\"%s\">Зайти →</a>", inviterName, lobbyURL)
	return n.sendMessage(ctx, inviteeTelegramID, text)
}
