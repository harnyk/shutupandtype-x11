package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SmartTransformer interface {
	Transform(transcript string) (string, error)
}

type openAICompatSmartTransformer struct {
	httpClient *http.Client
	chatURL    string
	apiKey     string
	model      string
	markers    smartMarkers
}

func newSmartTransformerFromConfig() (SmartTransformer, error) {
	chatURL, err := smartLLMChatCompletionsURL(cfgSmartLLMBaseURL())
	if err != nil {
		return nil, err
	}
	key := effectiveSmartLLMAPIKey()
	if key == "" {
		return nil, fmt.Errorf("smart LLM API key not configured")
	}
	return &openAICompatSmartTransformer{
		httpClient: http.DefaultClient,
		chatURL:    chatURL,
		apiKey:     key,
		model:      cfgSmartLLMModel(),
		markers:    cfgSmartMarkers(),
	}, nil
}

func smartTransform(transcript string) (string, error) {
	t, err := newSmartTransformerFromConfig()
	if err != nil {
		return "", err
	}
	return t.Transform(transcript)
}

func (t *openAICompatSmartTransformer) Transform(transcript string) (string, error) {
	system := buildSmartSystemPrompt(t.markers)
	body, err := json.Marshal(map[string]any{
		"model": t.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": transcript},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, t.chatURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API error %d: %s", resp.StatusCode, respBody)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("LLM returned empty content")
	}
	return out, nil
}

func buildSmartSystemPrompt(m smartMarkers) string {
	var b strings.Builder
	b.WriteString(`You post-process speech-to-text for paste into another app.

Output ONLY the final text to paste. No explanations, no markdown code fences, no preamble.

Default mode is PROSE until a mode marker switches behavior:
- PROSE: fix punctuation and capitalization; light grammar; preserve language (Russian/English); do not change meaning or add facts.
- CODE (after a code marker): recover operators, brackets, quotes, dots, indices, nullish coalescing, etc.; do not refactor or invent identifiers; prefer faithful symbols over natural language.
- VERBATIM (after a verbatim marker): keep words as spoken except obvious STT glitches; do not polish prose.

Mode markers in speech switch modes. They may be slightly corrupted by STT or spoken as synonyms. Never include marker phrases in the output.

Configured marker phrases (non-exhaustive; infer similar variants):
`)
	fmt.Fprintf(&b, "- CODE: %s\n", strings.Join(m.Code, ", "))
	fmt.Fprintf(&b, "- VERBATIM: %s\n", strings.Join(m.Verbatim, ", "))
	fmt.Fprintf(&b, "- PROSE (return to default): %s\n", strings.Join(m.Prose, ", "))
	b.WriteString(`
Text before the first marker uses PROSE.
If uncertain in CODE segments, prefer literal STT tokens over guessing new code.
`)
	return b.String()
}
