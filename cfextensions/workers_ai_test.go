package cfextensions

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func float64Ptr(v float64) *float64 { return &v }

func TestResolveOption(t *testing.T) {
	messages := []Message{{Role: "user", Content: "hi"}}

	tests := []struct {
		name    string
		ai      WorkersAi
		opt     RequestOption
		want    RequestOption
		wantErr string
	}{
		{
			name: "fills default max tokens",
			ai:   WorkersAi{},
			opt:  RequestOption{Model: "@cf/meta/llama-3-8b-instruct", Messages: messages},
			want: RequestOption{
				Model:     "@cf/meta/llama-3-8b-instruct",
				Messages:  messages,
				MaxTokens: defaultMaxTokens,
			},
		},
		{
			name: "keeps explicit max tokens and temperature",
			ai:   WorkersAi{},
			opt: RequestOption{
				Model:       "@cf/meta/llama-3-8b-instruct",
				Messages:    messages,
				MaxTokens:   16,
				Temperature: float64Ptr(0.2),
			},
			want: RequestOption{
				Model:       "@cf/meta/llama-3-8b-instruct",
				Messages:    messages,
				MaxTokens:   16,
				Temperature: float64Ptr(0.2),
			},
		},
		{
			name: "falls back to the binding model",
			ai:   WorkersAi{Model: "@cf/openai/gpt-oss-120b"},
			opt:  RequestOption{Messages: messages},
			want: RequestOption{
				Model:     "@cf/openai/gpt-oss-120b",
				Messages:  messages,
				MaxTokens: defaultMaxTokens,
			},
		},
		{
			name: "option model wins over the binding model",
			ai:   WorkersAi{Model: "@cf/openai/gpt-oss-120b"},
			opt:  RequestOption{Model: "@cf/meta/llama-3-8b-instruct", Messages: messages},
			want: RequestOption{
				Model:     "@cf/meta/llama-3-8b-instruct",
				Messages:  messages,
				MaxTokens: defaultMaxTokens,
			},
		},
		{
			name:    "no model at all",
			ai:      WorkersAi{},
			opt:     RequestOption{Messages: messages},
			wantErr: "cfextensions: model is required",
		},
		{
			name:    "no message",
			ai:      WorkersAi{Model: "@cf/openai/gpt-oss-120b"},
			opt:     RequestOption{},
			wantErr: "cfextensions: messages must not be empty",
		},
		{
			name:    "negative max tokens",
			ai:      WorkersAi{Model: "@cf/openai/gpt-oss-120b"},
			opt:     RequestOption{Messages: messages, MaxTokens: -1},
			wantErr: "cfextensions: max_tokens must not be negative, got -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ai.resolveOption(tt.opt)

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Equal(t, RequestOption{}, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveOptionDoesNotMutateCaller(t *testing.T) {
	ai := WorkersAi{Model: "@cf/openai/gpt-oss-120b"}
	opt := RequestOption{Messages: []Message{{Role: "user", Content: "hi"}}}

	_, err := ai.resolveOption(opt)

	assert.NoError(t, err)
	assert.Empty(t, opt.Model)
	assert.Zero(t, opt.MaxTokens)
}

func TestParseAiResponse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    AiResponse
		wantErr string
	}{
		{
			name: "response shape",
			raw:  `{"response":"hello"}`,
			want: AiResponse{Response: "hello"},
		},
		{
			name: "unknown fields are ignored",
			raw:  `{"response":"hello","usage":{"total_tokens":3}}`,
			want: AiResponse{Response: "hello"},
		},
		{
			name:    "broken json",
			raw:     `{"response":`,
			wantErr: `cfextensions: failed to parse response {"response":: unexpected end of JSON input`,
		},
		{
			name:    "json of the wrong type",
			raw:     `{"response":42}`,
			wantErr: "cfextensions: failed to parse response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAiResponse([]byte(tt.raw))

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, AiResponse{}, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAiResponseText(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "response field",
			raw:  `{"response":"from response"}`,
			want: "from response",
		},
		{
			name: "output items",
			raw:  `{"output":[{"type":"reasoning","content":[]},{"type":"message","content":[{"type":"output_text","text":"from output"}]}]}`,
			want: "from output",
		},
		{
			name: "openai style message",
			raw:  `{"choices":[{"message":{"content":"from choices message"}}]}`,
			want: "from choices message",
		},
		{
			name: "openai style completion text",
			raw:  `{"choices":[{"text":"from choices text"}]}`,
			want: "from choices text",
		},
		{
			name: "response wins over the other shapes",
			raw:  `{"response":"from response","choices":[{"text":"from choices text"}]}`,
			want: "from response",
		},
		{
			name: "no text at all",
			raw:  `{"choices":[{"text":""}],"output":[]}`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseAiResponse([]byte(tt.raw))

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res.Text())
		})
	}
}

func TestRequestWithoutBinding(t *testing.T) {
	opt := RequestOption{
		Model:    "@cf/openai/gpt-oss-120b",
		Messages: []Message{{Role: "user", Content: "hi"}},
	}

	tests := []struct {
		name string
		ai   *WorkersAi
		opt  RequestOption
		want error
	}{
		{
			name: "nil binding",
			ai:   nil,
			opt:  opt,
			want: ErrAiUnavailable,
		},
		{
			name: "zero binding on native",
			ai:   &WorkersAi{},
			opt:  opt,
			want: ErrAiUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ai.Request(tt.opt)

			assert.True(t, errors.Is(err, tt.want), "got %v", err)
			assert.Equal(t, AiResponse{}, got)
		})
	}
}

// Request must reject an invalid option before reaching the runtime, so the
// same error surfaces on native and on wasm.
func TestRequestValidatesBeforeCallingTheRuntime(t *testing.T) {
	ai := &WorkersAi{}

	got, err := ai.Request(RequestOption{Messages: []Message{{Role: "user", Content: "hi"}}})

	assert.EqualError(t, err, "cfextensions: model is required")
	assert.Equal(t, AiResponse{}, got)
}
