package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LLMClient interface {
	Analyze(ctx context.Context, prompt string) (string, error)
	Name() string
}

type NVIDIAClient struct {
	APIKey     string
	Model      string
	Endpoint   string
	HTTPClient *http.Client
}

func NewNVIDIAClient(apiKey, model, endpoint string) *NVIDIAClient {
	if endpoint == "" {
		endpoint = "https://integrate.api.nvidia.com/v1/chat/completions"
	}
	if model == "" {
		model = "nvidia/nemotron-3-ultra"
	}
	return &NVIDIAClient{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *NVIDIAClient) Name() string {
	return "nvidia"
}

type nvidiaRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type nvidiaResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (c *NVIDIAClient) Analyze(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("NVIDIA_API_KEY not set")
	}

	reqBody := nvidiaRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "system", Content: "You are kube-triage, an expert Kubernetes SRE. Provide concise, actionable root cause analysis and remediation commands."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   4096,
		TopP:        0.95,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var nvidiaResp nvidiaResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvidiaResp); err != nil {
		return "", err
	}

	if nvidiaResp.Error != nil {
		return "", fmt.Errorf("NVIDIA API error: %s", nvidiaResp.Error.Message)
	}

	if len(nvidiaResp.Choices) == 0 {
		return "", fmt.Errorf("no response from NVIDIA API")
	}

	return nvidiaResp.Choices[0].Message.Content, nil
}