package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shehanjayasinghe/sarathi_ai/internal/voice"
)

// Server exposes HTTP endpoints that proxy to the Python MCP voice stack.
type Server struct {
	Voice *voice.Client
	Root  string
}

type voiceRequest struct {
	Text      string `json:"text"`
	AudioPath string `json:"audio_path"`
	SkipTTS   bool   `json:"skip_tts"`
}

type voiceResponse struct {
	RawTranscript  string  `json:"raw_transcript"`
	NormalizedText string  `json:"normalized_text"`
	ReplyText      string  `json:"reply_text"`
	OutputAudio    *string `json:"output_audio,omitempty"`
	AudioURL       string  `json:"audio_url,omitempty"`
}

// Handler returns the root mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/voice", s.handleVoice)
	mux.HandleFunc("POST /api/voice/stream", s.handleVoiceStream)
	mux.HandleFunc("GET /api/voice/audio/", s.handleAudio)
	return withCORS(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVoice(w http.ResponseWriter, r *http.Request) {
	text, audioPath, err := s.parseVoiceInput(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	skipTTS := false
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		skipTTS = r.FormValue("skip_tts") == "true" || r.FormValue("skip_tts") == "1"
	}

	if (text == "") == (audioPath == "") {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "provide exactly one of text or audio upload/audio_path",
		})
		return
	}

	result, err := s.Voice.VoicePipeline(r.Context(), text, audioPath, skipTTS)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	resp := voiceResponse{
		RawTranscript:  result.RawTranscript,
		NormalizedText: result.NormalizedText,
		ReplyText:      result.ReplyText,
		OutputAudio:    result.OutputAudio,
	}
	if result.OutputAudio != nil && *result.OutputAudio != "" {
		resp.AudioURL = "/api/voice/audio/" + filepath.Base(*result.OutputAudio)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) parseVoiceInput(r *http.Request) (text, audioPath string, err error) {
	contentType := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return "", "", fmt.Errorf("invalid multipart form")
		}
		text = strings.TrimSpace(r.FormValue("text"))
		file, header, ferr := r.FormFile("audio")
		if ferr == nil {
			defer file.Close()
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".webm"
			}
			dir := filepath.Join(s.Root, "data", "voice")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return "", "", err
			}
			name := fmt.Sprintf("upload_%d%s", time.Now().UnixNano(), ext)
			dest := filepath.Join(dir, name)
			out, err := os.Create(dest)
			if err != nil {
				return "", "", err
			}
			if _, err := io.Copy(out, file); err != nil {
				out.Close()
				return "", "", err
			}
			out.Close()
			audioPath = dest
		}
		return text, audioPath, nil
	default:
		var req voiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return "", "", fmt.Errorf("invalid json")
		}
		return strings.TrimSpace(req.Text), strings.TrimSpace(req.AudioPath), nil
	}
}

func (s *Server) handleAudio(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/voice/audio/")
	name = filepath.Base(name)
	if name == "." || name == string(filepath.Separator) {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(s.Root, "data", "voice", name)
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	http.ServeFile(w, r, path)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
