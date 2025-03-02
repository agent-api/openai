package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models"
)

func main() {
	ctx := context.Background()

	// create a new std library logger
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	// Create an openai provider
	provider := openai.NewProvider(&openai.ProviderOpts{
		Logger: logger,
	})
	provider.UseModel(ctx, models.GPT4_O)

	// Create a new agent
	myAgent := agent.NewAgent(&agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a helpful assistant.",
	})

	result := myAgent.RunStream(
		ctx,
		agent.WithInput("Why is the sky blue?"),
	)

	for delta := range result.DeltaChan {
		print(delta)
	}
}
