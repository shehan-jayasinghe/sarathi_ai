package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shehanjayasinghe/sarathi_ai/internal/api"
	"github.com/shehanjayasinghe/sarathi_ai/internal/voice"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
	root := flag.String("root", "", "Repository root (default: cwd or parent of cmd)")
	flag.Parse()

	repoRoot := *root
	if repoRoot == "" {
		repoRoot = findRepoRoot()
	}

	voiceClient, err := voice.NewClient(voice.Config{RepoRoot: repoRoot})
	if err != nil {
		log.Fatalf("voice mcp client: %v", err)
	}
	defer voiceClient.Close()

	srv := &api.Server{Voice: voiceClient, Root: repoRoot}
	log.Printf("Sarathi API listening on http://%s (repo %s)", *addr, repoRoot)
	log.Printf("Frontend → Go → Python MCP voice pipeline ready")
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "tools", "voice", "mcp_server.py")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return wd
}
