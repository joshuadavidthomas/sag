package inworld

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client talks to the Inworld AI TTS HTTP API.
type Client struct {
	baseURL    string
	apiKey     string // Base64-encoded for Basic auth
	httpClient *http.Client
}

// NewClient returns a Client configured with the given API key and base URL.
// The API key should be the raw key; it will be used for Basic auth.
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.inworld.ai"
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Voice represents a voice available from Inworld.
type Voice struct {
	ID   string `json:"voiceId"`
	Name string `json:"name"`
}

// TTSRequest configures a text-to-speech request payload.
type TTSRequest struct {
	Text        string       `json:"text"`
	VoiceID     string       `json:"voiceId"`
	ModelID     string       `json:"modelId"`
	AudioConfig *AudioConfig `json:"audioConfig,omitempty"`
	Temperature *float64     `json:"temperature,omitempty"`
	Speed       *float64     `json:"speed,omitempty"`
}

// AudioConfig specifies the audio encoding configuration.
type AudioConfig struct {
	AudioEncoding   string `json:"audioEncoding,omitempty"`   // LINEAR16 or MP3
	SampleRateHertz int    `json:"sampleRateHertz,omitempty"` // e.g., 48000
}

// ttsResponse represents the non-streaming API response.
type ttsResponse struct {
	AudioContent string `json:"audioContent"` // Base64-encoded audio
}

// streamResponse represents a single chunk in the streaming response.
type streamResponse struct {
	Result struct {
		AudioContent string `json:"audioContent"`
	} `json:"result"`
}

// StreamTTS requests streaming audio for text-to-speech.
func (c *Client) StreamTTS(ctx context.Context, req TTSRequest) (io.ReadCloser, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/tts/v1/voice:stream")

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer func() {
			_ = resp.Body.Close()
		}()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stream TTS failed: %s: %s", resp.Status, string(b))
	}

	// Inworld returns newline-delimited JSON with base64 audio chunks.
	// We need to decode and stream the audio.
	return &streamDecoder{
		resp:    resp,
		scanner: bufio.NewScanner(resp.Body),
	}, nil
}

// streamDecoder decodes Inworld's newline-delimited JSON streaming response.
type streamDecoder struct {
	resp    *http.Response
	scanner *bufio.Scanner
	buf     []byte // decoded audio buffer
	pos     int    // current position in buffer
}

func (s *streamDecoder) Read(p []byte) (int, error) {
	// If we have buffered data, return it first
	if s.pos < len(s.buf) {
		n := copy(p, s.buf[s.pos:])
		s.pos += n
		return n, nil
	}

	// Read next line from the stream
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}

	line := strings.TrimSpace(s.scanner.Text())
	if line == "" {
		// Empty line, try next
		return s.Read(p)
	}

	var chunk streamResponse
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		// Skip malformed lines
		return s.Read(p)
	}

	if chunk.Result.AudioContent == "" {
		return s.Read(p)
	}

	// Decode base64 audio content
	audioData, err := base64.StdEncoding.DecodeString(chunk.Result.AudioContent)
	if err != nil {
		return 0, fmt.Errorf("failed to decode audio chunk: %w", err)
	}

	s.buf = audioData
	s.pos = 0
	n := copy(p, s.buf)
	s.pos = n
	return n, nil
}

func (s *streamDecoder) Close() error {
	return s.resp.Body.Close()
}

// ConvertTTS downloads the full audio before returning.
func (c *Client) ConvertTTS(ctx context.Context, req TTSRequest) ([]byte, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/tts/v1/voice")

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("convert TTS failed: %s: %s", resp.Status, string(b))
	}

	var ttsResp ttsResponse
	if err := json.NewDecoder(resp.Body).Decode(&ttsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Decode base64 audio content
	audioData, err := base64.StdEncoding.DecodeString(ttsResp.AudioContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio: %w", err)
	}

	return audioData, nil
}

// ListVoices is not directly supported by Inworld's public API.
// This returns a static list of known voices.
func (c *Client) ListVoices(_ context.Context, search string) ([]Voice, error) {
	// Inworld doesn't have a public voices endpoint.
	// Return common preset voices that are known to work.
	voices := []Voice{
		{ID: "Ashley", Name: "Ashley"},
		{ID: "Brian", Name: "Brian"},
		{ID: "Cora", Name: "Cora"},
		{ID: "David", Name: "David"},
		{ID: "Emma", Name: "Emma"},
		{ID: "James", Name: "James"},
		{ID: "Kate", Name: "Kate"},
		{ID: "Lily", Name: "Lily"},
		{ID: "Oliver", Name: "Oliver"},
		{ID: "Zoe", Name: "Zoe"},
	}

	if search == "" {
		return voices, nil
	}

	// Filter by search term
	search = strings.ToLower(search)
	var filtered []Voice
	for _, v := range voices {
		if strings.Contains(strings.ToLower(v.Name), search) {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}
