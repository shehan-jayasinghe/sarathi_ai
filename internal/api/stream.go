package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shehanjayasinghe/sarathi_ai/internal/voice"
)

type sseEvent struct {
	Event string
	Data  any
}

func (s *Server) handleVoiceStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}

	text, audioPath, err := s.parseVoiceInput(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if (text == "") == (audioPath == "") {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "provide exactly one of text or audio upload/audio_path",
		})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	var writeMu sync.Mutex
	send := func(ev sseEvent) {
		writeMu.Lock()
		defer writeMu.Unlock()
		payload, _ := json.Marshal(ev.Data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, payload)
		flusher.Flush()
	}

	send(sseEvent{Event: "status", Data: map[string]string{"phase": "transcribing"}})

	userText := text
	rawTranscript := text
	if audioPath != "" {
		raw, err := s.Voice.SpeechToText(r.Context(), audioPath)
		if err != nil {
			send(sseEvent{Event: "error", Data: map[string]string{"error": err.Error()}})
			return
		}
		rawTranscript = raw
		normalized, err := s.Voice.NormalizeSpeech(r.Context(), raw)
		if err != nil {
			normalized = raw
		}
		userText = normalized
	} else {
		normalized, err := s.Voice.NormalizeSpeech(r.Context(), text)
		if err == nil && normalized != "" {
			userText = normalized
		}
	}

	send(sseEvent{Event: "transcript", Data: map[string]string{
		"raw_transcript":  rawTranscript,
		"normalized_text": userText,
	}})
	send(sseEvent{Event: "status", Data: map[string]string{"phase": "thinking"}})

	sentenceCh := make(chan string, 16)
	var ttsWG sync.WaitGroup
	ttsWG.Add(1)
	go func() {
		defer ttsWG.Done()
		for sentence := range sentenceCh {
			sentence = strings.TrimSpace(sentence)
			if sentence == "" {
				continue
			}
			outPath := filepath.Join(s.Root, "data", "voice", fmt.Sprintf("reply_%d.wav", time.Now().UnixNano()))
			path, err := s.Voice.TextToSpeech(r.Context(), sentence, outPath)
			if err != nil {
				send(sseEvent{Event: "error", Data: map[string]string{"error": "tts: " + err.Error()}})
				continue
			}
			send(sseEvent{Event: "sentence", Data: map[string]string{
				"text":      sentence,
				"audio_url": "/api/voice/audio/" + filepath.Base(path),
			}})
		}
	}()

	var pending string
	full, err := voice.StreamQwen(r.Context(), userText, func(token string) error {
		send(sseEvent{Event: "token", Data: map[string]string{"text": token}})
		pending += token
		sentences, rest := voice.PopCompleteSentences(pending)
		pending = rest
		for _, snt := range sentences {
			sentenceCh <- snt
		}
		return nil
	})
	if err != nil {
		close(sentenceCh)
		ttsWG.Wait()
		send(sseEvent{Event: "error", Data: map[string]string{"error": err.Error()}})
		return
	}

	if leftover := strings.TrimSpace(pending); leftover != "" {
		sentenceCh <- leftover
	}
	close(sentenceCh)
	ttsWG.Wait()

	send(sseEvent{Event: "done", Data: map[string]string{"reply_text": full}})
}
