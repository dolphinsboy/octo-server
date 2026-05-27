package i18n

import (
	"fmt"
	"os"
	"strings"
)

const (
	// EnvDefaultLanguage controls the runtime fallback language for requests
	// that carry no explicit language signal.
	EnvDefaultLanguage = "OCTO_DEFAULT_LANGUAGE"

	// DefaultLanguage preserves the legacy deployment behavior for clients that
	// do not send Accept-Language yet.
	DefaultLanguage = "zh-CN"
)

// DefaultLanguageFromEnv resolves OCTO_DEFAULT_LANGUAGE into a supported BCP-47
// language tag. Empty env uses DefaultLanguage; invalid values are rejected so
// rollout misconfiguration fails during startup instead of surfacing per request.
func DefaultLanguageFromEnv() (string, error) {
	return ResolveDefaultLanguage(os.Getenv(EnvDefaultLanguage))
}

// ResolveDefaultLanguage normalizes the configured default language.
func ResolveDefaultLanguage(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultLanguage, nil
	}
	if lang, ok := MatchSupportedLanguage(raw); ok {
		return lang, nil
	}
	return "", fmt.Errorf("%s must be one of %s, got %q", EnvDefaultLanguage, strings.Join(SupportedLanguages(), ", "), raw)
}

// SupportedLanguages returns the current runtime language matrix.
func SupportedLanguages() []string {
	out := make([]string, 0, len(supportedLanguageTags))
	for _, tag := range supportedLanguageTags {
		out = append(out, tag.String())
	}
	return out
}

// ValidateRuntimeLocales checks the locale files required for startup:
// source language and configured default language must both have active TOML
// files in the embedded runtime bundle.
func ValidateRuntimeLocales(defaultLang string) error {
	normalizedDefault, err := ResolveDefaultLanguage(defaultLang)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, lang := range []string{SourceLanguage, normalizedDefault} {
		if seen[lang] {
			continue
		}
		seen[lang] = true
		if !activeLocaleExists(lang) {
			return fmt.Errorf("i18n active locale for %s is missing", lang)
		}
	}
	if _, err := Bundle(); err != nil {
		return err
	}
	return nil
}

func activeLocaleExists(lang string) bool {
	_, err := localesFS.ReadFile("locales/active." + lang + ".toml")
	return err == nil
}
