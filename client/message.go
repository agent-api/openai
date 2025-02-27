package client

import (
	"encoding/json"
	"fmt"
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

		toolCalls := []openai.ChatCompletionMessageToolCallParam{}
		for _, t := range m.ToolCalls {
			toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
				ID:   openai.F(t.ID),
				Type: openai.F(openai.ChatCompletionMessageToolCallType("function")),
				Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
					Name:      openai.F(t.Name),
					Arguments: openai.F(string(t.Arguments)),
				}),
			})
		}
		message.ToolCalls = openai.F(toolCalls)

		return message

	case types.ToolMessageRole:
		var s strings.Builder

		s.WriteString(fmt.Sprintf("%v", m.ToolResult[0].Content))

		if m.ToolResult[0].Error != "" {
			s.WriteString(m.ToolResult[0].Error)
		}

		message := openai.ToolMessage(m.ToolResult[0].ToolCallID, s.String())
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
		t := []*types.ToolCall{}

		for _, tool := range m.ToolCalls {
			t = append(t, &types.ToolCall{
				ID:        tool.ID,
				Name:      tool.Function.Name,
				Arguments: json.RawMessage(tool.Function.Arguments),
			})
		}

		return types.Message{
			Role:      types.ToolMessageRole,
			Content:   m.Content,
			ToolCalls: t,
		}
	}

	return types.Message{}
}
