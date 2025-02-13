package client

import (
	"context"
	"net/http"

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

func (c *OpenAIClient) Chat(ctx context.Context, req *ChatRequest) (ChatResponse, error) {
	openaiMessages := []openai.ChatCompletionMessageParamUnion{}

	for _, message := range req.Messages {
		openaiMessage := convertMessageToOpenAIMessage(message)
		openaiMessages = append(openaiMessages, openaiMessage)
	}

	res, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(openaiMessages),
		Model:    openai.F(c.model),
	})

	if err != nil {
		panic(err)
	}

	m := OpenAIChatCompletionMessageToAgentAPIMessage(&res.Choices[0].Message)

	return ChatResponse{
		Message: m,
		Model:   res.Model,
	}, nil
}
