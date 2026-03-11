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
	NotifyChallengeReceived(ctx context.Context, inviteeTelegramID int64, inviterName string, deepLink string) (int64, error)
	EditChallengeMessage(ctx context.Context, inviteeTelegramID int64, messageID int64, text string) error
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
func (n *NoOpNotifier) NotifyChallengeReceived(_ context.Context, _ int64, _ string, _ string) (int64, error) {
	return 0, nil
}
func (n *NoOpNotifier) EditChallengeMessage(_ context.Context, _ int64, _ int64, _ string) error {
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

type sendMessageWithButtonRequest struct {
	ChatID      int64          `json:"chat_id"`
	Text        string         `json:"text"`
	ParseMode   string         `json:"parse_mode"`
	ReplyMarkup inlineKeyboard `json:"reply_markup"`
}

type inlineKeyboard struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

type inlineButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type sendMessageResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

type editMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	MessageID int64  `json:"message_id"`
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

func (n *HTTPNotifier) sendMessageWithButton(ctx context.Context, chatID int64, text, buttonText, buttonURL string) (int64, error) {
	body, _ := json.Marshal(sendMessageWithButtonRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
		ReplyMarkup: inlineKeyboard{
			InlineKeyboard: [][]inlineButton{
				{{Text: buttonText, URL: buttonURL}},
			},
		},
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if !result.OK {
		return 0, fmt.Errorf("telegram API error")
	}
	return result.Result.MessageID, nil
}

func (n *HTTPNotifier) NotifyChallengeAccepted(ctx context.Context, inviterTelegramID int64, inviteeName string, lobbyURL string) error {
	text := fmt.Sprintf("⚔️ <b>%s</b> принял твой вызов и готов к дуэли!\n\n<a href=\"%s\">Зайти в лобби →</a>", inviteeName, lobbyURL)
	return n.sendMessage(ctx, inviterTelegramID, text)
}

func (n *HTTPNotifier) NotifyInviterWaiting(ctx context.Context, inviteeTelegramID int64, inviterName string, lobbyURL string) error {
	text := fmt.Sprintf("⚔️ <b>%s</b> ждёт тебя в лобби!\n\n<a href=\"%s\">Зайти →</a>", inviterName, lobbyURL)
	return n.sendMessage(ctx, inviteeTelegramID, text)
}

func (n *HTTPNotifier) NotifyChallengeReceived(ctx context.Context, inviteeTelegramID int64, inviterName string, deepLink string) (int64, error) {
	text := fmt.Sprintf(
		"⚔️ <b>Вызов на дуэль!</b>\n\n<b>%s</b> бросает тебе вызов в Quiz Sprint.\nУ тебя есть 1 час чтобы принять.",
		inviterName,
	)
	return n.sendMessageWithButton(ctx, inviteeTelegramID, text, "⚔️ Принять вызов", deepLink)
}

func (n *HTTPNotifier) EditChallengeMessage(ctx context.Context, inviteeTelegramID int64, messageID int64, text string) error {
	body, _ := json.Marshal(editMessageRequest{
		ChatID:    inviteeTelegramID,
		MessageID: messageID,
		Text:      text,
		ParseMode: "HTML",
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", n.botToken)
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
