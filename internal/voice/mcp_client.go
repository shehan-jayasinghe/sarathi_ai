package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// Client talks to the Sarathi Python MCP voice server over stdio.
type Client struct {
	mu       sync.Mutex
	mcp      *client.Client
	python   string
	server   string
	repoRoot string
}

// Config for connecting to the Python MCP server.
type Config struct {
	RepoRoot     string
	PythonPath   string
	MCPServerPath string
}

// PipelineResult mirrors the Python voice_pipeline tool response.
type PipelineResult struct {
	RawTranscript   string  `json:"raw_transcript"`
	NormalizedText  string  `json:"normalized_text"`
	ReplyText       string  `json:"reply_text"`
	OutputAudio     *string `json:"output_audio"`
}

// NewClient creates a voice MCP client. Call Close when done.
func NewClient(cfg Config) (*Client, error) {
	root := cfg.RepoRoot
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		root = wd
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	python := cfg.PythonPath
	if python == "" {
		python = filepath.Join(root, "tools", "voice", ".venv", "bin", "python")
	}
	server := cfg.MCPServerPath
	if server == "" {
		server = filepath.Join(root, "tools", "voice", "mcp_server.py")
	}

	if _, err := os.Stat(python); err != nil {
		return nil, fmt.Errorf("python venv not found at %s: %w", python, err)
	}
	if _, err := os.Stat(server); err != nil {
		return nil, fmt.Errorf("mcp server not found at %s: %w", server, err)
	}

	c := &Client{
		python:   python,
		server:   server,
		repoRoot: root,
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	mcpClient, err := client.NewStdioMCPClient(c.python, nil, c.server)
	if err != nil {
		return fmt.Errorf("start mcp stdio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "sarathi-go",
		Version: "0.1.0",
	}
	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		_ = mcpClient.Close()
		return fmt.Errorf("mcp initialize: %w", err)
	}

	c.mcp = mcpClient
	return nil
}

// Close shuts down the MCP subprocess.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mcp == nil {
		return nil
	}
	err := c.mcp.Close()
	c.mcp = nil
	return err
}

// VoicePipeline calls the MCP voice_pipeline tool.
// Provide exactly one of text or audioPath.
func (c *Client) VoicePipeline(ctx context.Context, text, audioPath string, skipTTS bool) (*PipelineResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.mcp == nil {
		return nil, fmt.Errorf("mcp client not connected")
	}

	args := map[string]any{
		"skip_tts": skipTTS,
	}
	text = strings.TrimSpace(text)
	audioPath = strings.TrimSpace(audioPath)
	switch {
	case text != "" && audioPath == "":
		args["text"] = text
	case audioPath != "" && text == "":
		args["audio_path"] = audioPath
	default:
		return nil, fmt.Errorf("provide exactly one of text or audio_path")
	}

	outDefault := filepath.Join(
		c.repoRoot,
		"data",
		"voice",
		fmt.Sprintf("reply_%d.wav", time.Now().UnixNano()),
	)
	args["output_path"] = outDefault

	req := mcp.CallToolRequest{}
	req.Params.Name = "voice_pipeline"
	req.Params.Arguments = args

	callCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	res, err := c.mcp.CallTool(callCtx, req)
	if err != nil {
		return nil, fmt.Errorf("voice_pipeline: %w", err)
	}
	if res.IsError {
		return nil, fmt.Errorf("voice_pipeline tool error: %s", toolText(res))
	}

	raw := toolText(res)
	var out PipelineResult
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		// Some servers wrap JSON in text; try to extract object
		if start := strings.Index(raw, "{"); start >= 0 {
			if end := strings.LastIndex(raw, "}"); end > start {
				if err2 := json.Unmarshal([]byte(raw[start:end+1]), &out); err2 == nil {
					return &out, nil
				}
			}
		}
		return nil, fmt.Errorf("decode pipeline result: %w; raw=%q", err, raw)
	}
	return &out, nil
}

// SpeechToText calls speech_to_text_tool on a local audio file.
func (c *Client) SpeechToText(ctx context.Context, audioPath string) (string, error) {
	return c.callTextTool(ctx, "speech_to_text_tool", map[string]any{"audio_path": audioPath})
}

// TextToSpeech calls text_to_speech_tool and returns the output WAV path.
func (c *Client) TextToSpeech(ctx context.Context, text, outputPath string) (string, error) {
	args := map[string]any{"text": text}
	if outputPath != "" {
		args["output_path"] = outputPath
	}
	return c.callTextTool(ctx, "text_to_speech_tool", args)
}

// RepoRoot returns the configured repository root.
func (c *Client) RepoRoot() string {
	return c.repoRoot
}

// NormalizeSpeech calls normalize_speech_tool.
func (c *Client) NormalizeSpeech(ctx context.Context, text string) (string, error) {
	return c.callTextTool(ctx, "normalize_speech_tool", map[string]any{"text": text})
}

// ChatQwen calls chat_qwen_tool.
func (c *Client) ChatQwen(ctx context.Context, text string) (string, error) {
	return c.callTextTool(ctx, "chat_qwen_tool", map[string]any{"text": text})
}

func (c *Client) callTextTool(ctx context.Context, name string, args map[string]any) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mcp == nil {
		return "", fmt.Errorf("mcp client not connected")
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	res, err := c.mcp.CallTool(callCtx, req)
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s error: %s", name, toolText(res))
	}
	return strings.TrimSpace(toolText(res)), nil
}

func toolText(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	var b strings.Builder
	for _, item := range res.Content {
		if tc, ok := mcp.AsTextContent(item); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}
