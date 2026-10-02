package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSmartLLMChatCompletionsURL(t *testing.T) {
	got, err := smartLLMChatCompletionsURL("https://api.openai.com/v1/")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://api.openai.com/v1/chat/completions"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got, err = smartLLMChatCompletionsURL("http://127.0.0.1:1234/v1")
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://127.0.0.1:1234/v1/chat/completions" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildSmartSystemPrompt_includesMarkers(t *testing.T) {
	p := buildSmartSystemPrompt(defaultSmartMarkers())
	if !strings.Contains(p, "enable coding mode") {
		t.Fatalf("expected code markers in prompt: %s", p)
	}
}

func TestOpenAICompatSmartTransformer_transform(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path %q", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer test-key") {
			t.Fatalf("auth header %q", r.Header.Get("Authorization"))
		}
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "test-model" || len(req.Messages) != 2 {
			t.Fatalf("req %+v", req)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "hello world"}},
			},
		})
	}))
	defer srv.Close()

	tr := &openAICompatSmartTransformer{
		httpClient: srv.Client(),
		chatURL:    srv.URL + "/v1/chat/completions",
		apiKey:     "test-key",
		model:      "test-model",
		markers:    defaultSmartMarkers(),
	}
	out, err := tr.Transform("raw transcript")
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello world" {
		t.Fatalf("got %q", out)
	}
}

func TestCfgSmartMarkers_defaultsWhenUnset(t *testing.T) {
	viper.Reset()
	viper.SetDefault("smart_markers", nil)
	m := cfgSmartMarkers()
	if len(m.Code) == 0 || len(m.Verbatim) == 0 {
		t.Fatalf("expected defaults %+v", m)
	}
}

func TestEffectiveSmartLLMAPIKey_fallback(t *testing.T) {
	viper.Reset()
	viper.Set("smart_llm_api_key", "")
	viper.Set("openai_api_key", "sk-fallback")
	if got := effectiveSmartLLMAPIKey(); got != "sk-fallback" {
		t.Fatalf("got %q", got)
	}
	viper.Set("smart_llm_api_key", "sk-dedicated")
	if got := effectiveSmartLLMAPIKey(); got != "sk-dedicated" {
		t.Fatalf("got %q", got)
	}
}
