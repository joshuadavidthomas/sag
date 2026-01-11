package cmd

import (
	"fmt"
	"os"
)

// ensureAPIKey loads the API key from flags or environment variables.
// Provider-specific env vars take precedence.
func ensureAPIKey() error {
	if cfg.APIKey != "" {
		return nil
	}

	// Try provider-specific environment variables first
	switch cfg.Provider {
	case "inworld":
		cfg.APIKey = os.Getenv("INWORLD_API_KEY")
	case "elevenlabs":
		cfg.APIKey = os.Getenv("ELEVENLABS_API_KEY")
	default:
		// Try provider name in uppercase as env var prefix
		cfg.APIKey = os.Getenv(fmt.Sprintf("%s_API_KEY", toEnvName(cfg.Provider)))
	}

	// Fallback to generic SAG_API_KEY
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("SAG_API_KEY")
	}

	// For backwards compatibility, also check ELEVENLABS_API_KEY for default provider
	if cfg.APIKey == "" && cfg.Provider == "elevenlabs" {
		cfg.APIKey = os.Getenv("ELEVENLABS_API_KEY")
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("missing API key for %s (set --api-key or %s_API_KEY)", cfg.Provider, toEnvName(cfg.Provider))
	}
	return nil
}

// toEnvName converts a provider name to an environment variable prefix.
func toEnvName(provider string) string {
	result := make([]byte, 0, len(provider))
	for i := 0; i < len(provider); i++ {
		c := provider[i]
		if c >= 'a' && c <= 'z' {
			c -= 32 // Convert to uppercase
		}
		if c == '-' {
			c = '_'
		}
		result = append(result, c)
	}
	return string(result)
}
