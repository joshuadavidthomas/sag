package elevenlabs

import (
	"context"
	"io"

	"github.com/steipete/sag/internal/tts"
)

const (
	ProviderName  = "elevenlabs"
	DefaultModel  = "eleven_v3"
	DefaultFormat = "mp3_44100_128"
)

// Provider wraps the ElevenLabs Client to implement tts.Provider.
type Provider struct {
	client *Client
}

// NewProvider creates a new ElevenLabs TTS provider.
func NewProvider(apiKey, baseURL string) *Provider {
	return &Provider{
		client: NewClient(apiKey, baseURL),
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return ProviderName
}

// DefaultModel returns the default ElevenLabs model.
func (p *Provider) DefaultModel() string {
	return DefaultModel
}

// DefaultFormat returns the default audio format.
func (p *Provider) DefaultFormat() string {
	return DefaultFormat
}

// StreamTTS requests streaming audio for text-to-speech.
func (p *Provider) StreamTTS(ctx context.Context, voiceID string, req tts.Request) (io.ReadCloser, error) {
	payload := mapRequest(req)
	return p.client.StreamTTS(ctx, voiceID, payload, req.LatencyTier)
}

// ConvertTTS downloads the full audio before returning.
func (p *Provider) ConvertTTS(ctx context.Context, voiceID string, req tts.Request) ([]byte, error) {
	payload := mapRequest(req)
	return p.client.ConvertTTS(ctx, voiceID, payload)
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
			ID:       v.VoiceID,
			Name:     v.Name,
			Category: v.Category,
			Provider: ProviderName,
			Labels:   v.Labels,
		}
	}
	return result, nil
}

// Client returns the underlying ElevenLabs client for direct access if needed.
func (p *Provider) Client() *Client {
	return p.client
}

// mapRequest converts a tts.Request to an ElevenLabs TTSRequest.
func mapRequest(req tts.Request) TTSRequest {
	payload := TTSRequest{
		Text:                   req.Text,
		ModelID:                req.Model,
		OutputFormat:           req.OutputFormat,
		Seed:                   req.Seed,
		ApplyTextNormalization: req.Normalize,
		LanguageCode:           req.LanguageCode,
	}

	// Only set VoiceSettings if at least one parameter is provided
	if req.Speed != nil || req.Stability != nil || req.SimilarityBoost != nil ||
		req.Style != nil || req.UseSpeakerBoost != nil {
		payload.VoiceSettings = &VoiceSettings{
			Speed:           req.Speed,
			Stability:       req.Stability,
			SimilarityBoost: req.SimilarityBoost,
			Style:           req.Style,
			UseSpeakerBoost: req.UseSpeakerBoost,
		}
	}

	return payload
}

func init() {
	tts.Register(ProviderName, func(apiKey, baseURL string) tts.Provider {
		return NewProvider(apiKey, baseURL)
	})
}
