package openai

import (
	"context"
	"fmt"

	"github.com/agent-api/core"
	"github.com/agent-api/openai/client"
	"github.com/go-logr/logr"
)

// Provider implements the LLMProvider interface for OpenAI
type Provider struct {
	host string
	port int

	model *core.Model

	// client is the internal Ollama HTTP client
	client *client.OpenAIClient

	logger logr.Logger
}

type ProviderOpts struct {
	BaseURL string
	Port    int
	APIKey  string

	Logger *logr.Logger
}

// NewProvider creates a new Ollama provider
func NewProvider(opts *ProviderOpts) *Provider {
	opts.Logger.Info("Creating new OpenAI provider")

	client := client.NewClient(
		opts.Logger,

		// TODO - need to enable local env variable, not just through opt
		//client.WithAPIKey(opts.APIKey),
	)

	return &Provider{
		client: client,
		logger: *opts.Logger,
	}
}

func (p *Provider) GetCapabilities(ctx context.Context) (*core.Capabilities, error) {
	p.logger.Info("Fetching capabilities")

	// Placeholder for future implementation
	p.logger.Info("GetCapabilities method is not implemented yet")

	return nil, nil
}

func (p *Provider) UseModel(ctx context.Context, model *core.Model) error {
	p.logger.Info("Setting model", "modelID", model.ID)

	p.model = model

	return nil
}

// Generate implements the LLMProvider interface for basic responses
func (p *Provider) Generate(ctx context.Context, opts *core.GenerateOptions) (*core.Message, error) {
	p.logger.Info("Generate request received", "modelID", p.model.ID)

	resp, err := p.client.Chat(ctx, &client.ChatRequest{
		Model:    p.model.ID,
		Messages: opts.Messages,
		Tools:    opts.Tools,
	})

	if err != nil {
		p.logger.V(0).Error(err, "Error calling client chat method", err)
		return nil, fmt.Errorf("error calling client chat method: %w", err)
	}

	return &core.Message{
		Role:      core.AssistantMessageRole,
		Content:   resp.Message.Content,
		ToolCalls: resp.Message.ToolCalls,
	}, nil
}

// GenerateStream streams the response token by token
func (p *Provider) GenerateStream(ctx context.Context, opts *core.GenerateOptions) (<-chan *core.Message, <-chan string, <-chan error) {
	p.logger.Info("Starting stream generation", "modelID", p.model.ID)

	return p.client.ChatStream(ctx, &client.ChatRequest{
		Model:    p.model.ID,
		Messages: opts.Messages,
		Tools:    opts.Tools,
	})
}
