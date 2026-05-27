package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Update struct {
	UpdateID      int64          `json:"update_id"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type CallbackQuery struct {
	ID      string `json:"id"`
	Data    string `json:"data"`
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		MessageID int `json:"message_id"`
	} `json:"message"`
}

func (t *Telegram) DeleteWebhook(ctx context.Context) error {
	_, err := t.apiGet(ctx, "deleteWebhook", "")
	return err
}

func (t *Telegram) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	q := fmt.Sprintf("offset=%d&timeout=%d&allowed_updates=%s", offset, timeoutSec, `["callback_query"]`)
	raw, err := t.apiGet(ctx, "getUpdates", q)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if !parsed.OK {
		return nil, fmt.Errorf("telegram getUpdates not ok")
	}
	return parsed.Result, nil
}

func (t *Telegram) AnswerCallback(ctx context.Context, callbackID, text string) error {
	body, err := json.Marshal(map[string]string{
		"callback_query_id": callbackID,
		"text":              text,
		"show_alert":        "false",
	})
	if err != nil {
		return err
	}
	_, err = t.apiPost(ctx, "answerCallbackQuery", body)
	return err
}

func (t *Telegram) EditMessageAfterFix(ctx context.Context, chatID string, messageID int, htmlText string) error {
	body, err := json.Marshal(map[string]any{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         htmlText,
		"parse_mode":   "HTML",
		"reply_markup": inlineKeyboard{InlineKeyboard: [][]inlineButton{}},
	})
	if err != nil {
		return err
	}
	_, err = t.apiPost(ctx, "editMessageText", body)
	return err
}

func (t *Telegram) apiGet(ctx context.Context, method, query string) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, method)
	if query != "" {
		url += "?" + query
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram %s: %s %s", method, resp.Status, string(raw))
	}
	return raw, nil
}

func (t *Telegram) apiPost(ctx context.Context, method string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, method)
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram %s: %s %s", method, resp.Status, string(raw))
	}
	return raw, nil
}
