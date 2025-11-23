package ai

import (
	"context"
	"fmt"
)

// ModelType represents the type of AI model
type ModelType string

const (
	ModelTypeLLM    ModelType = "llm"
	ModelTypeVision ModelType = "vision"
	ModelTypeSpeech ModelType = "speech"
)

// ProviderType represents the AI provider
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderLocal     ProviderType = "local"
)

// Request represents an AI inference request
type Request struct {
	ModelType ModelType
	Provider  ProviderType
	ModelName string
	Prompt    string
	Images    []string // Base64 or URL
	Options   map[string]interface{}
}

// Response represents an AI inference response
type Response struct {
	Text   string
	Usage  map[string]int
	Cached bool
}

// Provider interface for AI providers
type Provider interface {
	Generate(ctx context.Context, req Request) (*Response, error)
}

// Manager manages AI providers and routing
type Manager struct {
	providers map[ProviderType]Provider
}

// NewManager creates a new AI manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[ProviderType]Provider),
	}
}

// RegisterProvider registers an AI provider
func (m *Manager) RegisterProvider(pt ProviderType, p Provider) {
	m.providers[pt] = p
}

// Generate routes the request to the appropriate provider
func (m *Manager) Generate(ctx context.Context, req Request) (*Response, error) {
	provider, ok := m.providers[req.Provider]
	if !ok {
		// Fallback to local if not specified? Or error?
		// For now, error
		return nil, fmt.Errorf("provider %s not found", req.Provider)
	}

	return provider.Generate(ctx, req)
}
