package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	ChineseClass101URL = "https://www.chineseclass101.com/api/word-day/%s"
	HTTPTimeout        = 10 * time.Second
	MaxTokens          = 150
	GPTModel           = "gpt-4.1"
)

type WordService struct {
	client       *http.Client
	openaiClient *openai.Client
	prompt       string
}

type WordOfDayResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Payload struct {
		WordDay struct {
			Date         string `json:"date"`
			Text         string `json:"text"`
			English      string `json:"english"`
			Meaning      string `json:"meaning"`
			Romanization string `json:"romanization"`
			Traditional  string `json:"traditional"`
			AudioTarget  string `json:"audio_target"`
			ImageURL     string `json:"image_url"`
			Samples      []struct {
				Text         string `json:"text"`
				English      string `json:"english"`
				Romanization string `json:"romanization"`
				Traditional  string `json:"traditional"`
				AudioTarget  string `json:"audio_target"`
			} `json:"samples"`
		} `json:"word_day"`
	} `json:"payload"`
}

type ChineseWord struct {
	Chinese      string
	English      string
	Romanization string
	Meaning      string
	AudioURL     string
	ImageURL     string
	Samples      []string
}

func NewWordService(openaiClient *openai.Client, prompt string) *WordService {
	return &WordService{
		client: &http.Client{
			Timeout: HTTPTimeout,
		},
		openaiClient: openaiClient,
		prompt:       prompt,
	}
}

func (ws *WordService) GetTodaysWordWithTranslation(ctx context.Context) (string, error) {
	today := time.Now().Format("2006-01-02")

	wordResponse, err := ws.fetchWordOfDay(ctx, today)
	if err != nil {
		return "", fmt.Errorf("fetching word of day: %w", err)
	}

	translation, err := ws.translateWord(ctx, wordResponse)
	if err != nil {
		return "", fmt.Errorf("translating word: %w", err)
	}

	result := "## 今日の中国語\n" + fmt.Sprintf("単語: %s\n", wordResponse.Payload.WordDay.Text) + translation

	return result, nil
}

func (ws *WordService) fetchWordOfDay(ctx context.Context, date string) (*WordOfDayResponse, error) {
	url := fmt.Sprintf(ChineseClass101URL, date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var wordResponse WordOfDayResponse
	if err := json.NewDecoder(resp.Body).Decode(&wordResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if wordResponse.Status != "success" {
		return nil, fmt.Errorf("API returned error status: %s", wordResponse.Status)
	}

	return &wordResponse, nil
}

func (ws *WordService) translateWord(ctx context.Context, wordResponse *WordOfDayResponse) (string, error) {
	wordDay := wordResponse.Payload.WordDay
	prompt := fmt.Sprintf("%s\n\n%s %s", ws.prompt, wordDay.Text, wordDay.Romanization)

	req := openai.ChatCompletionRequest{
		Model: GPTModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: MaxTokens,
	}

	resp, err := ws.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("creating completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
