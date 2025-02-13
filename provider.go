package openai

import (
	"context"
	"fmt"

	"github.com/agent-api/core/types"
	"github.com/agent-api/openai/client"
	"github.com/go-logr/logr"
)

// Provider implements the LLMProvider interface for OpenAI
type Provider struct {
	host string
	port int

	model *types.Model

	// client is the internal Ollama HTTP client
	client *client.OpenAIClient

	logger logr.Logger
}

type ProviderOpts struct {
	BaseURL string
	Port    int
	APIKey  string

	Logger logr.Logger
}

// NewProvider creates a new Ollama provider
func NewProvider(opts *ProviderOpts) *Provider {
	opts.Logger.V(0).Info("Creating new OpenAI provider")

	client := client.NewClient(
		client.WithAPIKey(opts.APIKey),
	)

	return &Provider{
		client: client,
		logger: opts.Logger,
	}
}

func (p *Provider) GetCapabilities(ctx context.Context) (*types.Capabilities, error) {
	p.logger.V(1).Info("Fetching capabilities")

	// Placeholder for future implementation
	p.logger.V(1).Info("GetCapabilities method is not implemented yet")

	return nil, nil
}

func (p *Provider) UseModel(ctx context.Context, model *types.Model) error {
	p.logger.V(1).Info("Setting model", "modelID", model.ID)

	p.model = model

	return nil
}

// Generate implements the LLMProvider interface for basic responses
func (p *Provider) Generate(ctx context.Context, opts *types.GenerateOptions) (*types.Message, error) {
	p.logger.V(1).Info("Generate request received", "modelID", p.model.ID)

	mPtrs := []*types.Message{}
	for _, m := range opts.Messages {
		mPtrs = append(mPtrs, &m)
	}

	resp, err := p.client.Chat(ctx, &client.ChatRequest{
		Model:    p.model.ID,
		Messages: mPtrs,
	})

	if err != nil {
		p.logger.Error(err, "Error calling client chat method")
		return nil, fmt.Errorf("error calling client chat method: %w", err)
	}

	//toolCalls := []types.ToolCall{}
	//for _, toolCall := range resp.Message.ToolCalls {
	//toolCalls = append(toolCalls, types.ToolCall{
	//Name:      toolCall.Function.Name,
	//Arguments: toolCall.Function.Arguments,
	//})
	//}

	return &types.Message{
		Role:    types.AssistantMessageRole,
		Content: resp.Message.Content,
		//ToolCalls: toolCalls,
	}, nil
}

// GenerateStream streams the response token by token
func (p *Provider) GenerateStream(ctx context.Context, opts *types.GenerateOptions) (<-chan *types.Message, <-chan error) {
	p.logger.V(1).Info("Stream generation not implemented yet")
	return nil, nil
}
