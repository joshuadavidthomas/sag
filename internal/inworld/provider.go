package inworld

import (
	"context"
	"io"

	"github.com/steipete/sag/internal/tts"
)

const (
	ProviderName  = "inworld"
	DefaultModel  = "inworld-tts-1"
	DefaultFormat = "mp3"
)

// Provider wraps the Inworld Client to implement tts.Provider.
type Provider struct {
	client *Client
}

// NewProvider creates a new Inworld TTS provider.
func NewProvider(apiKey, baseURL string) *Provider {
	return &Provider{
		client: NewClient(apiKey, baseURL),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return ProviderName
}

// DefaultModel returns the default Inworld model.
func (p *Provider) DefaultModel() string {
	return DefaultModel
}

// DefaultFormat returns the default audio format.
func (p *Provider) DefaultFormat() string {
	return DefaultFormat
}

// StreamTTS requests streaming audio for text-to-speech.
func (p *Provider) StreamTTS(ctx context.Context, voiceID string, req tts.Request) (io.ReadCloser, error) {
	payload := mapRequest(voiceID, req)
	return p.client.StreamTTS(ctx, payload)
}

// ConvertTTS downloads the full audio before returning.
func (p *Provider) ConvertTTS(ctx context.Context, voiceID string, req tts.Request) ([]byte, error) {
	payload := mapRequest(voiceID, req)
	return p.client.ConvertTTS(ctx, payload)
}

// ListVoices returns available voices.
func (p *Provider) ListVoices(ctx context.Context, search string) ([]tts.Voice, error) {
	voices, err := p.client.ListVoices(ctx, search)
	if err != nil {
		return nil, err
	}

	result := make([]tts.Voice, len(voices))
	for i, v := range voices {
		result[i] = tts.Voice{
			ID:       v.ID,
			Name:     v.Name,
			Category: "preset",
			Provider: ProviderName,
		}
	}
	return result, nil
}

// Client returns the underlying Inworld client for direct access if needed.
func (p *Provider) Client() *Client {
	return p.client
}

// mapRequest converts a tts.Request to an Inworld TTSRequest.
func mapRequest(voiceID string, req tts.Request) TTSRequest {
	payload := TTSRequest{
		Text:    req.Text,
		VoiceID: voiceID,
		ModelID: req.Model,
	}

	// Map temperature if provided
	if req.Temperature != nil {
		payload.Temperature = req.Temperature
	}

	// Map speed if provided
	if req.Speed != nil {
		payload.Speed = req.Speed
	}

	// Set audio config based on output format
	if req.OutputFormat != "" {
		payload.AudioConfig = mapAudioConfig(req.OutputFormat)
	}

	return payload
}

// mapAudioConfig maps a format string to Inworld's AudioConfig.
func mapAudioConfig(format string) *AudioConfig {
	switch format {
	case "mp3":
		return &AudioConfig{AudioEncoding: "MP3", SampleRateHertz: 48000}
	case "linear16", "pcm", "wav":
		return &AudioConfig{AudioEncoding: "LINEAR16", SampleRateHertz: 48000}
	default:
		// Default to MP3
		return &AudioConfig{AudioEncoding: "MP3", SampleRateHertz: 48000}
	}
}

func init() {
	tts.Register(ProviderName, func(apiKey, baseURL string) tts.Provider {
		return NewProvider(apiKey, baseURL)
	})
}
