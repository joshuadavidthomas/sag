// Package tts defines the interface for text-to-speech providers.
package tts

import (
	"context"
	"io"
)

// Provider defines the interface that all TTS providers must implement.
type Provider interface {
	// StreamTTS requests streaming audio for text-to-speech.
	// Returns an io.ReadCloser that streams audio data.
	StreamTTS(ctx context.Context, voiceID string, req Request) (io.ReadCloser, error)

	// ConvertTTS downloads the full audio before returning.
	ConvertTTS(ctx context.Context, voiceID string, req Request) ([]byte, error)

	// ListVoices returns available voices, optionally filtered by search term.
	ListVoices(ctx context.Context, search string) ([]Voice, error)

	// Name returns the provider's identifier (e.g., "elevenlabs", "inworld").
	Name() string

	// DefaultModel returns the provider's default model ID.
	DefaultModel() string

	// DefaultFormat returns the provider's default output format.
	DefaultFormat() string
}

// Request contains common TTS request parameters.
// Providers map these to their API-specific formats.
type Request struct {
	Text         string
	Model        string
	OutputFormat string

	// Voice tuning parameters (providers ignore unsupported params)
	Speed       *float64 // Speech speed multiplier (typically 0.5-2.0)
	Temperature *float64 // Randomness/expressiveness (Inworld: 0-2, ElevenLabs: use Stability)

	// ElevenLabs-specific (ignored by other providers)
	Stability       *float64
	SimilarityBoost *float64
	Style           *float64
	UseSpeakerBoost *bool
	Seed            *uint32
	Normalize       string // auto|on|off
	LanguageCode    string

	// Streaming optimization
	LatencyTier int // ElevenLabs latency optimization tier (0-4)
}

// Voice represents a voice available from a provider.
type Voice struct {
	ID       string
	Name     string
	Category string            // e.g., "premade", "cloned", "generated"
	Provider string            // Provider name for disambiguation
	Labels   map[string]string // Provider-specific metadata
}
