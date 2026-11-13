package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type Manager struct {
	clients []LLMClient
	primary string
}

func NewManager(provider, model, endpoint string) *Manager {
	var clients []LLMClient
	primary := strings.ToLower(provider)

	switch primary {
	case "nvidia":
		clients = append(clients, NewNVIDIAClient(
			getEnv("NVIDIA_API_KEY", ""),
			model,
			endpoint,
		))
		// Fallback to Ollama
		clients = append(clients, NewOllamaClient(
			getEnv("OLLAMA_HOST", ""),
			getEnv("OLLAMA_MODEL", "nemotron3-ultra"),
		))

	case "ollama":
		clients = append(clients, NewOllamaClient(
			getEnv("OLLAMA_HOST", ""),
			model,
		))

	case "openai":
		clients = append(clients, NewOpenAIClient(
			getEnv("OPENAI_API_KEY", ""),
			model,
			endpoint,
		))
		clients = append(clients, NewNVIDIAClient(
			getEnv("NVIDIA_API_KEY", ""),
			getEnv("NVIDIA_MODEL", "nvidia/nemotron-3-ultra"),
			"",
		))

	case "anthropic":
		clients = append(clients, NewAnthropicClient(
			getEnv("ANTHROPIC_API_KEY", ""),
			model,
			endpoint,
		))
		clients = append(clients, NewNVIDIAClient(
			getEnv("NVIDIA_API_KEY", ""),
			getEnv("NVIDIA_MODEL", "nvidia/nemotron-3-ultra"),
			"",
		))

	case "auto":
		// Try NVIDIA first (free tier)
		if key := getEnv("NVIDIA_API_KEY", ""); key != "" {
			clients = append(clients, NewNVIDIAClient(key, getEnv("NVIDIA_MODEL", "nvidia/nemotron-3-ultra"), ""))
			primary = "nvidia"
		}
		// Try Ollama local
		if host := getEnv("OLLAMA_HOST", ""); host != "" {
			clients = append(clients, NewOllamaClient(host, getEnv("OLLAMA_MODEL", "nemotron3-ultra")))
			if primary == "" {
				primary = "ollama"
			}
		}
		// Try OpenAI
		if key := getEnv("OPENAI_API_KEY", ""); key != "" {
			clients = append(clients, NewOpenAIClient(key, getEnv("OPENAI_MODEL", "gpt-4o"), ""))
			if primary == "" {
				primary = "openai"
			}
		}
		// Try Anthropic
		if key := getEnv("ANTHROPIC_API_KEY", ""); key != "" {
			clients = append(clients, NewAnthropicClient(key, getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"), ""))
			if primary == "" {
				primary = "anthropic"
			}
		}

	case "none":
		// No AI clients
		clients = []LLMClient{}

	default:
		// Try to parse as custom endpoint
		clients = append(clients, NewNVIDIAClient(
			getEnv("NVIDIA_API_KEY", ""),
			model,
			endpoint,
		))
	}

	return &Manager{
		clients: clients,
		primary: primary,
	}
}

func (m *Manager) Analyze(ctx context.Context, prompt string) (string, error) {
	if len(m.clients) == 0 {
		return "", fmt.Errorf("no AI clients configured")
	}

	var lastErr error
	for _, client := range m.clients {
		result, err := client.Analyze(ctx, prompt)
		if err == nil && result != "" {
			return result, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("all AI clients failed, last error: %w", lastErr)
}

func (m *Manager) GetPrimaryProvider() string {
	return m.primary
}

func (m *Manager) GetAvailableProviders() []string {
	providers := make([]string, len(m.clients))
	for i, c := range m.clients {
		providers[i] = c.Name()
	}
	return providers
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

type NoopClient struct{}

func (n *NoopClient) Name() string { return "none" }
func (n *NoopClient) Analyze(ctx context.Context, prompt string) (string, error) {
	return "", fmt.Errorf("AI analysis disabled (provider=none)")
}