// Package cfextensions wraps Cloudflare bindings that syumai/workers does not
// expose yet, so a Bindings struct can declare them with a `cfbinding` tag.
package cfextensions

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrAiUnavailable is returned when the AI binding is missing — on a native
// build, or when wrangler.jsonc declares no binding under the tagged name.
var ErrAiUnavailable = errors.New("cfextensions: AI binding is not available")

// Message is one chat turn sent to Workers AI.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RequestOption is the payload of a Workers AI inference call. Model is sent
// as the first argument of `AI.run`, not as part of the JSON body.
type RequestOption struct {
	Model       string    `json:"-"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
}

// AiContent is one content part of an AiOutput.
type AiContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AiOutput is one output item of a responses-style model.
type AiOutput struct {
	Type    string      `json:"type"`
	Content []AiContent `json:"content"`
}

// AiChoice is one completion of an OpenAI-compatible model.
type AiChoice struct {
	Text    string `json:"text"`
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// AiResponse covers the three response shapes Workers AI returns depending on
// the model: a bare `response`, `output` items, or OpenAI-style `choices`.
type AiResponse struct {
	Response string     `json:"response"`
	Output   []AiOutput `json:"output"`
	Choices  []AiChoice `json:"choices"`
}

// Text returns the first non-empty completion text, whichever shape the model
// answered in, or "" when the response carries no text.
func (r AiResponse) Text() string {
	if r.Response != "" {
		return r.Response
	}

	for _, output := range r.Output {
		for _, content := range output.Content {
			if content.Text != "" {
				return content.Text
			}
		}
	}

	for _, choice := range r.Choices {
		if choice.Message.Content != "" {
			return choice.Message.Content
		}
		if choice.Text != "" {
			return choice.Text
		}
	}

	return ""
}

const defaultMaxTokens = 1024

// Request runs an inference on the bound Workers AI account.
//
//	func Handler(ctx MyContext) error {
//		res, err := ctx.Env().AI.Request(cfextensions.RequestOption{
//			Model: "@cf/openai/gpt-oss-120b",
//			Messages: []cfextensions.Message{
//				{Role: "system", Content: systemPrompt},
//				{Role: "user", Content: prompt},
//			},
//		})
//		if err != nil {
//			return err
//		}
//		return ctx.Text(res.Text())
//	}
//
// It returns ErrAiUnavailable on a native build, so the same handler compiles
// for both targets.
func (w *WorkersAi) Request(opt RequestOption) (AiResponse, error) {
	if w == nil {
		return AiResponse{}, ErrAiUnavailable
	}

	opt, err := w.resolveOption(opt)
	if err != nil {
		return AiResponse{}, err
	}

	raw, err := w.run(opt)
	if err != nil {
		return AiResponse{}, err
	}

	return parseAiResponse(raw)
}

// resolveOption validates opt and fills in the defaults: the model falls back
// to WorkersAi.Model, and an unset MaxTokens to defaultMaxTokens.
func (w *WorkersAi) resolveOption(opt RequestOption) (RequestOption, error) {
	if opt.Model == "" {
		opt.Model = w.Model
	}
	if opt.Model == "" {
		return RequestOption{}, errors.New("cfextensions: model is required")
	}
	if len(opt.Messages) == 0 {
		return RequestOption{}, errors.New("cfextensions: messages must not be empty")
	}
	if opt.MaxTokens < 0 {
		return RequestOption{}, fmt.Errorf("cfextensions: max_tokens must not be negative, got %d", opt.MaxTokens)
	}
	if opt.MaxTokens == 0 {
		opt.MaxTokens = defaultMaxTokens
	}

	return opt, nil
}

func parseAiResponse(raw []byte) (AiResponse, error) {
	var parsed AiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return AiResponse{}, fmt.Errorf("cfextensions: failed to parse response %s: %w", raw, err)
	}

	return parsed, nil
}
