package client

import (
	"strings"

	"github.com/agent-api/core/types"
	"github.com/openai/openai-go"
)

func convertMessageToOpenAIMessage(m *types.Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case types.UserMessageRole:
		message := openai.UserMessage(m.Content)
		return message

	case types.AssistantMessageRole:
		message := openai.AssistantMessage(m.Content)
		return message

	case types.ToolMessageRole:
		message := openai.ToolMessage("", m.Content)
		return message
	}

	return nil
}

func convertOpenAIMessageToMessage(m *openai.Message) types.Message {
	content := strings.Builder{}

	for _, c := range m.Content {
		_, err := content.WriteString(c.Text.Value)
		if err != nil {
			panic(err)
		}
	}

	switch m.Role {
	case "user":
		return types.Message{
			Role:    types.UserMessageRole,
			Content: content.String(),
		}

	case "assistant":
		return types.Message{
			Role:    types.AssistantMessageRole,
			Content: content.String(),
		}

	case "tool":
		return types.Message{
			Role:    types.ToolMessageRole,
			Content: content.String(),
		}
	}

	return types.Message{}
}

func OpenAIChatCompletionMessageToAgentAPIMessage(m *openai.ChatCompletionMessage) types.Message {
	switch m.Role {
	case "user":
		return types.Message{
			Role:    types.UserMessageRole,
			Content: m.Content,
		}

	case "assistant":
		return types.Message{
			Role:    types.AssistantMessageRole,
			Content: m.Content,
		}

	case "tool":
		return types.Message{
			Role:    types.ToolMessageRole,
			Content: m.Content,
		}
	}

	return types.Message{}
}
