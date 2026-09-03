package voice

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	DefaultOllamaURL   = "http://localhost:11434"
	DefaultOllamaModel = "qwen2.5:3b"
)

// SystemPrompt matches the Python voice assistant prompt.
const SystemPrompt = "You are Sarathi, a local desktop voice assistant. " +
	"Reply in clear spoken-friendly English, 1-3 short sentences unless code is required. " +
	"Do not mention being an AI model unless asked."

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Stream   bool            `json:"stream"`
	Messages []ollamaMessage `json:"messages"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaStreamChunk struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// StreamQwen streams plain-text tokens from local Ollama into onToken.
func StreamQwen(ctx context.Context, userText string, onToken func(string) error) (string, error) {
	payload := ollamaChatRequest{
		Model:  DefaultOllamaModel,
		Stream: true,
		Messages: []ollamaMessage{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: userText},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DefaultOllamaURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, string(b))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk ollamaStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			full.WriteString(chunk.Message.Content)
			if err := onToken(chunk.Message.Content); err != nil {
				return full.String(), err
			}
		}
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), err
	}
	out := strings.TrimSpace(full.String())
	if out == "" {
		return "", fmt.Errorf("empty reply from ollama")
	}
	return out, nil
}

// PopCompleteSentences pulls finished sentences from buf; returns leftover buffer.
func PopCompleteSentences(buf string) (sentences []string, rest string) {
	start := 0
	for i := 0; i < len(buf); {
		r, size := utf8.DecodeRuneInString(buf[i:])
		if size <= 0 {
			break
		}
		if r != '.' && r != '!' && r != '?' {
			i += size
			continue
		}
		end := i + size
		if end < len(buf) {
			nr, _ := utf8.DecodeRuneInString(buf[end:])
			if !unicode.IsSpace(nr) && nr != '"' && nr != '\'' {
				i += size
				continue
			}
		}
		piece := strings.TrimSpace(buf[start:end])
		if piece != "" {
			sentences = append(sentences, piece)
		}
		for end < len(buf) {
			nr, nsz := utf8.DecodeRuneInString(buf[end:])
			if !unicode.IsSpace(nr) {
				break
			}
			end += nsz
		}
		start = end
		i = end
	}
	return sentences, buf[start:]
}
