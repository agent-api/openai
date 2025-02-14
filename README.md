# OpenAI `agent-api` provider

The OpenAI provider for `agent-api`

---

👷🏗️ The OpenAI provider is a work in progress and the API may change unexpectedly.

# Usage

```go
// Create an openai provider
provider := openai.NewProvider(&openai.ProviderOpts{})
provider.UseModel(ctx, gpt4o.GPT4_O)

// Create a new agent-api agent with openai
myAgent := agent.NewAgent(&agent.NewAgentConfig{
	Provider:     provider,
	SystemPrompt: "You are a helpful assistant.",
})

// Send a message to the agent
response, err := myAgent.Run(ctx, "Why is the sky blue?", agent.DefaultStopCondition)
```
