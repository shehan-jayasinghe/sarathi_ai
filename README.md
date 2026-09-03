# Sarathi AI

A human-first, offline-first desktop AI assistant that uses small specialist models and tools only when needed.

## Vision

Sarathi is not one giant AI model. It is a desktop assistant that understands the user's task, selects the right specialist model or deterministic tool, remembers relevant context, explains what it plans to do, and asks before making meaningful changes.

**Understand → Explain → Ask → Execute → Report**

## Initial Target

- Apple Mac mini M4
- 16 GB unified memory
- 256 GB SSD
- Native desktop application
- Go + Wails backend/shell
- React + TypeScript UI
- Local-first model execution

## Planned Capabilities

- General conversational assistant
- Coding assistant
- File/project understanding
- Compiler, LSP, test, and lint integration
- Code vulnerability and dependency analysis
- Local speech-to-text and text-to-speech
- Long-term relevant memory
- Document assistance
- Email and external-tool integrations through controlled permissions/MCP
- Vision and image capabilities loaded only when needed
- Optional local model fine-tuning from approved datasets

## Model Strategy

Sarathi uses a dynamic specialist architecture. Models are loaded on demand and unloaded when inactive so the application remains practical on machines with limited unified memory.

Examples include:

- Small general LLM
- Small coding LLM
- Whisper-family speech-to-text model
- Lightweight text-to-speech model
- Small vulnerability-detection model
- Vision/document models when required

Deterministic tools should be preferred for tasks such as parsing files, running tests, inspecting dependencies, and reading compiler/LSP diagnostics.

## Memory

Memory is local and separated into:

- **Working memory** — current task
- **Episodic memory** — useful past interactions
- **Semantic memory** — stable facts and preferences
- **Project memory** — architecture, conventions, and confirmed decisions

Memory is retrieved by relevance. Temporary or low-value memories can be summarized, compacted, or expired instead of continually growing the model context.

## Human-First Safety

Sarathi should not silently modify the user's work. Consequential actions such as changing files, deleting files, installing/removing packages, or sending messages require user approval.

For coding tasks, the default workflow is:

1. Understand the request/error.
2. Identify the relevant file/context.
3. Explain what Sarathi intends to change.
4. Show a focused diff.
5. Ask for approval.
6. Apply the change.
7. Run validation and report the result.

## Project Layout

```text
cmd/sarathi/          # Application entrypoint (Wails/Go)
internal/
  conversation/       # Conversation manager
  router/             # Intent router
  models/             # Model manager
  tools/              # Deterministic tools
  memory/             # Local memory
  permissions/        # Approval gate
frontend/src/         # React + TypeScript UI
models/registry/      # Model metadata (not weights)
data/                 # Local runtime state (gitignored)
scripts/              # Dev helpers
docs/                 # Architecture and plan
```

## Run API + UI

```bash
# Terminal 1 — Go API (talks to Python MCP)
go run ./cmd/sarathi -addr 127.0.0.1:8080

# Terminal 2 — frontend
cd frontend && npm run dev
```

Click the top-right robot → mic → **streaming** `POST /api/voice/stream` (SSE):
tokens appear live, Kokoro speaks **sentence by sentence**.

## Voice pipeline (local)

```text
Audio → Whisper → Speech Normalizer → Qwen (Ollama) → Kokoro → Audio
```

```bash
source tools/voice/.venv/bin/activate
cd tools/voice

# Text path (skip mic/STT)
python pipeline_cli.py --text "Open file number twenty five and change line ten."

# Audio path
python pipeline_cli.py --audio /path/to/clip.wav --out ../../data/voice/reply.wav

# MCP server (stdio) — for Cursor / future desktop host
python mcp_server.py
```

Cursor config example: `tools/voice/mcp.cursor.example.json`

MCP tools: `speech_to_text_tool`, `normalize_speech_tool`, `chat_qwen_tool`, `text_to_speech_tool`, `voice_pipeline`

Models: `qwen2.5:3b` (Ollama), `mlx-community/whisper-base-mlx`, Kokoro-82M.

## Documentation

- [Implementation Plan](docs/IMPLEMENTATION_PLAN.md)
- [Architecture](docs/ARCHITECTURE.md)

## Status

Floating agent overlay UI is in place. Local voice pipeline runs via **Frontend → Go API → Python MCP**. Next: Wails transparent desktop shell.
