# Sarathi AI — Implementation Plan

## Goal
Build Sarathi AI as a fast, human-first desktop AI helper. It should use specialist local models only when needed, keep relevant memory locally, explain intended actions, and ask before meaningful changes.

## Phase 1 — Desktop Foundation
- [ ] Build the desktop shell with Go + Wails.
- [ ] Add React + TypeScript UI.
- [ ] Create a small assistant/conversation window.
- [ ] Add local filesystem/project access.
- [ ] Establish a Go service layer for model and tool orchestration.

## Phase 2 — Conversation Core
- [ ] Add a small general-purpose local LLM through Ollama.
- [ ] Implement conversation state and task context.
- [ ] Implement intent detection/routing.
- [ ] Define the Sarathi interaction contract: understand → explain → ask → execute.
- [ ] Prevent silent file/system changes.

## Phase 3 — Memory System
- [ ] Implement working memory for the current task.
- [ ] Implement episodic memory for useful past interactions.
- [ ] Implement project memory for architecture, conventions, and confirmed decisions.
- [ ] Store memory locally on the filesystem/local database.
- [ ] Retrieve only memories relevant to the current task.
- [ ] Add memory scoring, compaction, summarization, and expiry for low-value temporary memories.
- [ ] Keep secrets and credentials outside normal model context.

## Phase 4 — Coding Assistant
- [ ] Add a small coding LLM, initially through Ollama.
- [ ] Detect project language/framework from files and project metadata.
- [ ] Read only the minimum relevant code context.
- [ ] Connect compiler/linter/LSP diagnostics to the assistant.
- [ ] Map an error to the relevant file and location.
- [ ] Generate one focused proposed change at a time.
- [ ] Show a human-readable explanation and diff before applying.
- [ ] Apply changes only after user confirmation.
- [ ] Add rollback/revert support.

## Phase 5 — Code Intelligence and Security
- [ ] Add lightweight vulnerability detection for supported languages.
- [ ] Integrate deterministic SAST/security scanners where available.
- [ ] Scan dependencies for known vulnerabilities.
- [ ] Detect potentially unused dependencies.
- [ ] Separate scanner findings from LLM interpretation.
- [ ] Never automatically remove dependencies or modify security-sensitive code.

## Phase 6 — Voice Assistant
- [ ] Add local speech-to-text using a small Whisper-family model.
- [ ] Add local text-to-speech using a lightweight engine/model.
- [ ] Implement voice activity detection.
- [ ] Allow voice conversation to use the same conversation and memory layer.
- [ ] Load speech models only when voice is active.
- [ ] Unload/stop inactive models to control memory usage.

## Phase 7 — Dynamic Model Manager
- [ ] Create a model registry describing each model's capability, size, runtime, and hardware requirements.
- [ ] Load models on demand.
- [ ] Keep frequently used models warm for a short configurable period.
- [ ] Unload inactive heavy models.
- [ ] Avoid keeping multiple large models resident on 16 GB machines.
- [ ] Run lightweight workers in parallel where useful.
- [ ] Add resource/memory monitoring and graceful fallback.

## Phase 8 — Documents and General Assistant Tasks
- [ ] Add document reading and summarization.
- [ ] Add controlled document editing with diff/preview.
- [ ] Add email drafting and other external actions through approved tools/APIs.
- [ ] Add MCP/tool integration behind a permission layer.
- [ ] Require confirmation for sending messages, modifying files, installing/removing packages, or other consequential actions.

## Phase 9 — Vision and Additional Specialists
- [ ] Add a small vision model for image/screenshot understanding when required.
- [ ] Add image generation only when requested.
- [ ] Add specialist models for additional domains only when a real use case exists.
- [ ] Keep each specialist independently loadable/unloadable.

## Phase 10 — Local Training / Improvement Pipeline
- [ ] Create a dedicated training-data directory structure.
- [ ] Collect approved question/answer and interaction examples.
- [ ] Remove or redact sensitive information before training.
- [ ] Clean and validate datasets.
- [ ] Use an authorized teacher model to generate/evaluate suitable examples where appropriate.
- [ ] Fine-tune an open small model rather than training a foundation model from scratch.
- [ ] Evaluate new models against a fixed Sarathi benchmark before adoption.
- [ ] Version datasets and model releases.

## Target Architecture

```text
Sarathi Desktop (Go + Wails + React)
            |
     Conversation Manager
            |
       Intent Router
            |
     Model / Tool Manager
       /      |       \
 Coding    Voice    Documents
 Model     Models      Models
       \      |       /
          Memory
            |
      Permission Layer
            |
      Local Files / MCP
```

## First Deliverable
The first usable milestone is intentionally small:

1. Launch Sarathi as a native desktop app.
2. Open a local project.
3. Chat with a local model.
4. Detect the current task/context.
5. Remember the current task.
6. Make one focused coding proposal.
7. Show the diff.
8. Ask the user before applying it.

Everything else should be added incrementally after this loop works reliably.
