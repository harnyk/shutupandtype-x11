package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type smartMarkers struct {
	Code     []string `mapstructure:"code"`
	Verbatim []string `mapstructure:"verbatim"`
	Prose    []string `mapstructure:"prose"`
}

func defaultSmartMarkers() smartMarkers {
	return smartMarkers{
		Code:     []string{"enable coding mode"},
		Verbatim: []string{"дословно", "verbatim"},
		Prose:    []string{"обычный текст", "prose mode"},
	}
}

func cfgSmartMode() bool {
	viperMu.RLock()
	defer viperMu.RUnlock()
	return viper.GetBool("smart_mode")
}

func cfgSmartLLMHTTPTimeout() time.Duration {
	d := cfgTimeout()
	if d < 60*time.Second {
		return 60 * time.Second
	}
	return d
}

func cfgSmartMarkers() smartMarkers {
	def := defaultSmartMarkers()
	viperMu.RLock()
	defer viperMu.RUnlock()
	if !viper.IsSet("smart_markers") {
		return def
	}
	var m smartMarkers
	if err := viper.UnmarshalKey("smart_markers", &m); err != nil {
		eventLogf("smart_markers config: %v; using defaults", err)
		return def
	}
	if len(m.Code) == 0 {
		m.Code = def.Code
	}
	if len(m.Verbatim) == 0 {
		m.Verbatim = def.Verbatim
	}
	if len(m.Prose) == 0 {
		m.Prose = def.Prose
	}
	return m
}

func effectiveSmartLLMAPIKey() string {
	viperMu.RLock()
	defer viperMu.RUnlock()
	if k := strings.TrimSpace(viper.GetString("smart_llm_api_key")); k != "" {
		return k
	}
	return strings.TrimSpace(viper.GetString("openai_api_key"))
}

func cfgSmartLLMModel() string {
	viperMu.RLock()
	defer viperMu.RUnlock()
	if m := strings.TrimSpace(viper.GetString("smart_llm_model")); m != "" {
		return m
	}
	return "gpt-4o-mini"
}

func cfgSmartLLMBaseURL() string {
	viperMu.RLock()
	defer viperMu.RUnlock()
	b := strings.TrimSpace(viper.GetString("smart_llm_base_url"))
	if b == "" {
		return "https://api.openai.com/v1"
	}
	return strings.TrimRight(b, "/")
}

func smartLLMChatCompletionsURL(base string) (string, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("smart_llm_base_url: %w", err)
	}
	ref := u.JoinPath("chat", "completions")
	return ref.String(), nil
}

func normalizeSmartConfig() {
	viperMu.Lock()
	defer viperMu.Unlock()
	if !viper.GetBool("smart_mode") {
		return
	}
	if effectiveSmartLLMAPIKeyUnlocked() == "" {
		eventLogf("smart_mode is enabled but no smart_llm_api_key or openai_api_key; disabling smart mode")
		viper.Set("smart_mode", false)
		return
	}
	if _, err := smartLLMChatCompletionsURL(cfgSmartLLMBaseURLUnlocked()); err != nil {
		eventLogf("smart_llm_base_url invalid: %v; disabling smart mode", err)
		viper.Set("smart_mode", false)
	}
}

func effectiveSmartLLMAPIKeyUnlocked() string {
	if k := strings.TrimSpace(viper.GetString("smart_llm_api_key")); k != "" {
		return k
	}
	return strings.TrimSpace(viper.GetString("openai_api_key"))
}

func cfgSmartLLMBaseURLUnlocked() string {
	b := strings.TrimSpace(viper.GetString("smart_llm_base_url"))
	if b == "" {
		return "https://api.openai.com/v1"
	}
	return strings.TrimRight(b, "/")
}

func smartModeCanEnable() bool {
	return effectiveSmartLLMAPIKey() != ""
}
