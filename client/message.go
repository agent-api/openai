package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agent-api/core"
	"github.com/openai/openai-go"
)

func convertMessageToOpenAIMessage(m *core.Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case core.UserMessageRole:
		message := openai.UserMessage(m.Content)
		return message

	case core.AssistantMessageRole:
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

	case core.ToolMessageRole:
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

func convertOpenAIMessageToMessage(m *openai.Message) core.Message {
	content := strings.Builder{}

	for _, c := range m.Content {
		_, err := content.WriteString(c.Text.Value)
		if err != nil {
			panic(err)
		}
	}

	switch m.Role {
	case "user":
		return core.Message{
			Role:    core.UserMessageRole,
			Content: content.String(),
		}

	case "assistant":
		return core.Message{
			Role:    core.AssistantMessageRole,
			Content: content.String(),
		}
	}

	return core.Message{}
}

func OpenAIChatCompletionMessageToAgentAPIMessage(m *openai.ChatCompletionMessage) core.Message {
	switch m.Role {
	case "user":
		return core.Message{
			Role:    core.UserMessageRole,
			Content: m.Content,
		}

	case "assistant":
		t := []*core.ToolCall{}

		for _, tool := range m.ToolCalls {
			t = append(t, &core.ToolCall{
				ID:        tool.ID,
				Name:      tool.Function.Name,
				Arguments: json.RawMessage(tool.Function.Arguments),
			})
		}

		return core.Message{
			Role:      core.ToolMessageRole,
			Content:   m.Content,
			ToolCalls: t,
		}
	}

	return core.Message{}
}
