package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/agent-api/core/types"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIClient struct {
	opts []option.RequestOption

	client *openai.Client

	model string
}

// ClientOption is a function that modifies the client
type ClientOption option.RequestOption

func NewClient(options ...ClientOption) *OpenAIClient {
	o := &OpenAIClient{}

	for _, option := range options {
		o.opts = append(o.opts, option)
	}

	client := openai.NewClient(o.opts...)

	return &OpenAIClient{
		client: client,
		model:  "gpt-4o",
	}
}

// WithBaseURL sets the base URL for the client
func WithBaseURL(url string) ClientOption {
	return option.WithBaseURL(url)
}

// WithAPIKey
func WithAPIKey(key string) ClientOption {
	return option.WithAPIKey(key)
}

// WithHTTPClient
func WithHTTPClient(client *http.Client) ClientOption {
	return option.WithHTTPClient(client)
}

// Convert your Tool to OpenAI's ChatCompletionToolParam
func ToOpenAIToolParam(t *types.Tool) (*openai.ChatCompletionToolParam, error) {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(t.JSONSchema, &schemaMap); err != nil {
		return nil, err
	}

	return &openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(t.Name),
			Description: openai.String(t.Description),
			Parameters:  openai.F(openai.FunctionParameters(schemaMap)),
		}),
	}, nil
}

func (c *OpenAIClient) Chat(ctx context.Context, req *ChatRequest) (ChatResponse, error) {
	openaiMessages := []openai.ChatCompletionMessageParamUnion{}

	for i, message := range req.Messages {
		openaiMessage := convertMessageToOpenAIMessage(message)
		fmt.Printf("message %d - %v\n", i, openaiMessage)
		openaiMessages = append(openaiMessages, openaiMessage)
	}

	openaiTools := []openai.ChatCompletionToolParam{}

	for _, tool := range req.Tools {
		t, _ := ToOpenAIToolParam(tool)

		openaiTools = append(openaiTools, *t)
	}

	res, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(openaiMessages),
		Model:    openai.F(c.model),
		Tools:    openai.F(openaiTools),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("res: %v\n", res.Choices[0].Message.Role)

	m := OpenAIChatCompletionMessageToAgentAPIMessage(&res.Choices[0].Message)

	return ChatResponse{
		Message: m,
		Model:   res.Model,
	}, nil
}
