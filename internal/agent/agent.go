package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/genai"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

// DefaultModel is the Gemini model used when neither Config.Model nor the
// ADK_MODEL environment variable names one.
const DefaultModel = "gemini-flash-latest"

const (
	appName   = "text-correction-improvement"
	agentName = "text_correction_agent"
	userID    = "local-user"
)

// GenerateFunc produces one raw model response for one correction request. It
// is the only seam between this package and the model provider.
type GenerateFunc func(ctx context.Context, req correction.Request) (string, error)

// Config configures an Agent.
type Config struct {
	// Model overrides the ADK model name. Empty means ADK_MODEL, then DefaultModel.
	Model string
	// APIKey authenticates Gemini. Empty means GOOGLE_API_KEY.
	APIKey string
	// Generate replaces the ADK Gemini call. Nil means the real call; tests
	// inject a fake, so no test needs a network or a credential.
	Generate GenerateFunc
}

// Agent corrects one document per request.
type Agent struct {
	generate GenerateFunc
	model    string
}

// New builds an agent. With an injected Generate it touches nothing outside the
// process; otherwise it constructs the ADK Gemini model and requires an API key.
func New(ctx context.Context, cfg Config) (*Agent, error) {
	modelName := resolveModel(cfg.Model)

	if cfg.Generate != nil {
		return &Agent{generate: cfg.Generate, model: modelName}, nil
	}

	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	}
	if apiKey == "" {
		return nil, &correction.Error{
			Code:   correction.CodeMissingAPIKey,
			Detail: "set GOOGLE_API_KEY (or pass Config.APIKey) to call Gemini, or inject Config.Generate",
		}
	}

	generate, err := newGeminiGenerator(ctx, modelName, apiKey)
	if err != nil {
		return nil, err
	}
	return &Agent{generate: generate, model: modelName}, nil
}

// Model reports the model name this agent uses.
func (a *Agent) Model() string { return a.model }

// Correct validates the request, asks the model, decodes the answer, and
// verifies it. An invalid request never reaches the model, and a draft that
// fails verification is blocked and discarded: an unverified revision is never
// handed back as if it had been reviewed.
func (a *Agent) Correct(ctx context.Context, req correction.Request) (correction.Result, error) {
	if err := correction.ValidateRequest(req); err != nil {
		return correction.Result{}, err
	}

	raw, err := a.generate(ctx, req)
	if err != nil {
		return correction.Result{}, &correction.Error{
			Code:   correction.CodeModelCallFailed,
			Detail: "the model call failed: " + err.Error(),
			Cause:  err,
		}
	}

	result, err := DecodeResult([]byte(raw))
	if err != nil {
		return correction.Result{}, err
	}

	if findings := correction.Verify(req, result); len(findings) > 0 {
		return correction.Result{}, &correction.BlockedError{Findings: findings}
	}
	return result, nil
}

func resolveModel(configured string) string {
	if name := strings.TrimSpace(configured); name != "" {
		return name
	}
	if name := strings.TrimSpace(os.Getenv("ADK_MODEL")); name != "" {
		return name
	}
	return DefaultModel
}

// newGeminiGenerator wires the ADK agent and returns the seam the rest of the
// package uses. The runner, the session service, and the model all stay inside
// this closure.
func newGeminiGenerator(ctx context.Context, modelName, apiKey string) (GenerateFunc, error) {
	llm, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, &correction.Error{
			Code:   correction.CodeModelCallFailed,
			Detail: "create the Gemini model: " + err.Error(),
			Cause:  err,
		}
	}

	root, err := llmagent.New(llmagent.Config{
		Name:        agentName,
		Model:       llm,
		Description: "Corrects one supplied text, reports every change with its reason, and asks about ambiguities instead of guessing.",
		Instruction: Instruction(),
		GenerateContentConfig: &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   ResultSchema(),
		},
	})
	if err != nil {
		return nil, &correction.Error{
			Code:   correction.CodeModelCallFailed,
			Detail: "create the ADK agent: " + err.Error(),
			Cause:  err,
		}
	}

	r, err := runner.NewInMemory(appName, root)
	if err != nil {
		return nil, &correction.Error{
			Code:   correction.CodeModelCallFailed,
			Detail: "create the ADK runner: " + err.Error(),
			Cause:  err,
		}
	}

	return func(ctx context.Context, req correction.Request) (string, error) {
		// One document per session: a fresh session id per request keeps one
		// document's state from reaching the next.
		sessionID, err := newSessionID()
		if err != nil {
			return "", err
		}

		var draft strings.Builder
		message := genai.NewContentFromText(Prompt(req), genai.RoleUser)
		for event, err := range r.Run(ctx, userID, sessionID, message, adkagent.RunConfig{}) {
			if err != nil {
				return "", err
			}
			if event == nil || !event.IsFinalResponse() {
				continue
			}
			draft.WriteString(textOf(event.Content))
		}

		if strings.TrimSpace(draft.String()) == "" {
			return "", errors.New("the model returned no text")
		}
		return draft.String(), nil
	}, nil
}

// textOf concatenates the text parts of one model response and skips thought
// parts, which are reasoning rather than the answer.
func textOf(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var out strings.Builder
	for _, part := range content.Parts {
		if part == nil || part.Thought {
			continue
		}
		out.WriteString(part.Text)
	}
	return out.String()
}

// newSessionID returns a fresh session id, so no two corrections share session
// state.
func newSessionID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate a session id: %w", err)
	}
	return "doc-" + hex.EncodeToString(buf), nil
}
