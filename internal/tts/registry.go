package tts

import (
	"fmt"
	"sort"
)

// ProviderFactory creates a provider instance with the given API key and base URL.
type ProviderFactory func(apiKey, baseURL string) Provider

var providers = make(map[string]ProviderFactory)

// Register adds a provider factory to the registry.
// This should be called from provider package init() functions.
func Register(name string, factory ProviderFactory) {
	providers[name] = factory
}

// Get returns a provider instance by name.
func Get(name, apiKey, baseURL string) (Provider, error) {
	factory, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s (available: %v)", name, Available())
	}
	return factory(apiKey, baseURL), nil
}

// Available returns a sorted list of registered provider names.
func Available() []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsRegistered checks if a provider is registered.
func IsRegistered(name string) bool {
	_, ok := providers[name]
	return ok
}
