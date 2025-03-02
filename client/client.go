package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/agent-api/core/types"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIClient struct {
	opts []option.RequestOption

	client *openai.Client

	model string

	logger *slog.Logger
}

// ClientOption is a function that modifies the client
type ClientOption option.RequestOption

func NewClient(logger *slog.Logger, options ...ClientOption) *OpenAIClient {
	o := &OpenAIClient{}

	for _, option := range options {
		o.opts = append(o.opts, option)
	}

	client := openai.NewClient(o.opts...)

	return &OpenAIClient{
		client: client,
		model:  "gpt-4o",
		logger: logger,
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

	chatParams := openai.ChatCompletionNewParams{
		Messages: openai.F(openaiMessages),
		Model:    openai.F(c.model),
	}

	if len(openaiTools) != 0 {
		chatParams.Tools = openai.F(openaiTools)
	}

	res, err := c.client.Chat.Completions.New(ctx, chatParams)
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

func (c *OpenAIClient) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *types.Message, <-chan string, <-chan error) {
	c.logger.Debug("received chat stream message request")

	msgChan := make(chan *types.Message)
	deltaChan := make(chan string)
	errChan := make(chan error, 1)

	openaiMessages := []openai.ChatCompletionMessageParamUnion{}
	for _, message := range req.Messages {
		openaiMessages = append(openaiMessages, convertMessageToOpenAIMessage(message))
	}

	c.logger.Debug("creating openai tools")
	openaiTools := []openai.ChatCompletionToolParam{}
	for _, tool := range req.Tools {
		t, err := ToOpenAIToolParam(tool)
		if err != nil {
			errChan <- fmt.Errorf("error converting tool: %w", err)
			close(msgChan)
			close(deltaChan)
			close(errChan)
			return msgChan, deltaChan, errChan
		}
		openaiTools = append(openaiTools, *t)
	}

	chatParams := openai.ChatCompletionNewParams{
		Messages: openai.F(openaiMessages),
		Model:    openai.F(req.Model),
	}

	if len(openaiTools) > 0 {
		chatParams.Tools = openai.F(openaiTools)
	}

	c.logger.Debug("kicking async go func for chat stream")

	go func() {
		stream := c.client.Chat.Completions.NewStreaming(ctx, chatParams)

		defer close(msgChan)
		defer close(deltaChan)
		defer close(errChan)
		defer stream.Close()

		// Create accumulator for building the final message
		acc := openai.ChatCompletionAccumulator{}

		// holds the message to be processed and sent down through the Go chan
		//currentMessage := &ChatResponse{}

		// iterate SSE from OpenAI stream
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			delta := ""

			if content, ok := acc.JustFinishedContent(); ok {
				// send the final message.
				// Blocks on reader grabbing message off channel
				msgChan <- &types.Message{
					Content: content,
				}
			}

			// if using tool calls
			if tool, ok := acc.JustFinishedToolCall(); ok {
				// send message with tool call to msg chan.
				// blocks on message being consumed by consumer.
				msgChan <- &types.Message{
					ToolCalls: []*types.ToolCall{
						{
							ID:        "woof_stream_tool",
							Name:      tool.Name,
							Arguments: json.RawMessage(tool.Arguments),
						},
					},
				}
			}

			if refusal, ok := acc.JustFinishedRefusal(); ok {
				c.logger.Error("unhandled refusal stream finished", "refusal", refusal)
			}

			if len(chunk.Choices) > 0 {
				delta = chunk.Choices[0].Delta.Content
			}

			select {
			case deltaChan <- delta:
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}

		if err := stream.Err(); err != nil {
			errChan <- fmt.Errorf("stream error: %w", err)
		}
	}()

	return msgChan, deltaChan, errChan
}
