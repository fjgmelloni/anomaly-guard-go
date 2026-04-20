package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"anomaly-guard-go/internal/model"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string, httpClient *http.Client) *Client {
	if apiKey == "" {
		return nil
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout:   12 * time.Second,
			Transport: httpClient.Transport,
		},
	}
}

func (c *Client) Decide(ctx context.Context, input model.DecisionInput) (model.Decision, error) {
	payload := requestBody{
		Contents: []content{
			{
				Parts: []part{
					{Text: buildPrompt(input)},
				},
			},
		},
		GenerationConfig: generationConfig{
			Temperature:      0.1,
			ResponseMimeType: "application/json",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return model.Decision{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(baseURL, c.model, c.apiKey),
		bytes.NewReader(body),
	)
	if err != nil {
		return model.Decision{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.Decision{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		responseBody, _ := io.ReadAll(resp.Body)
		return model.Decision{}, fmt.Errorf("gemini API returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}

	var response geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return model.Decision{}, err
	}

	text, err := response.text()
	if err != nil {
		return model.Decision{}, err
	}

	var decision model.Decision
	if err := json.Unmarshal([]byte(text), &decision); err != nil {
		return model.Decision{}, err
	}

	return decision, nil
}

func buildPrompt(input model.DecisionInput) string {
	return fmt.Sprintf(`You are a fraud decision engine.
Analyze the transaction and return only JSON with this schema:
{"risk_level":"low|medium|high|critical","recommended_action":"approve|monitor|manual_review|block","priority":"P1|P2|P3","reason":"short explanation"}

Rules:
- Be strict and concise.
- Higher anomaly_score means higher risk.
- Unknown region, many failed attempts, high amount, and unusual hour increase risk.
- Never return markdown.

Transaction:
{
  "id": %q,
  "account_id": %q,
  "amount": %.2f,
  "hour": %d,
  "region": %q,
  "known_region": %t,
  "failed_attempts_10m": %d,
  "minutes_since_last_tx": %d,
  "anomaly_score": %.2f,
  "suspicious": %t
}`, input.Transaction.ID, input.Transaction.AccountID, input.Transaction.Amount, input.Transaction.Hour, input.Transaction.Region, input.Transaction.KnownRegion, input.Transaction.FailedAttempts10m, input.Transaction.MinutesSinceLastTx, input.AnomalyScore, input.Suspicious)
}

type requestBody struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType"`
}

type geminiResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content content `json:"content"`
}

func (r geminiResponse) text() (string, error) {
	if len(r.Candidates) == 0 || len(r.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini response did not contain any text")
	}
	return r.Candidates[0].Content.Parts[0].Text, nil
}
