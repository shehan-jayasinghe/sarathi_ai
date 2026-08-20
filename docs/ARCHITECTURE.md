# Sarathi AI — Architecture

## Product Principle
Sarathi AI is a human-first desktop assistant, not a single always-running AI model. It routes each task to the smallest suitable model or deterministic tool, keeps relevant memory locally, and asks before consequential actions.

## Runtime Layers

### 1. Desktop Shell
- Go + Wails
- React + TypeScript UI
- Native desktop window; no browser-based product dependency

### 2. Conversation Manager
Maintains the current conversation, current task, selected context, and user-confirmed decisions.

### 3. Intent Router
Classifies the request and selects a capability such as coding, voice, documents, security, email, or general conversation.

### 4. Model Manager
Maintains a registry of local models and controls their lifecycle:
- load on demand
- warm briefly when useful
- unload when inactive
- avoid multiple large models competing for 16 GB unified memory
- expose model health and resource status

### 5. Specialist Workers
Examples:
- General local LLM
- Coding LLM
- Speech-to-text
- Text-to-speech
- Security/vulnerability model
- Vision model
- Document model

### 6. Deterministic Tools
Prefer deterministic tools over an LLM when the task does not require generation. Examples:
- file operations
- JSON/CSV parsing
- compiler and test execution
- LSP diagnostics
- dependency inspection
- security scanners
- Git operations

### 7. Memory Manager
Memory is local and separated by purpose:
- working memory: current task
- episodic memory: useful past interactions
- semantic memory: stable user/project facts
- project memory: architecture, conventions, and confirmed decisions

Memory should be retrieved by relevance rather than loading the entire history into every prompt. Low-value temporary memories can be summarized, compacted, or expired.

### 8. Permission Layer
Consequential actions require explicit approval. Examples:
- editing source files
- deleting files
- installing/removing packages
- sending emails/messages
- running potentially destructive commands

The normal flow is:

```text
Understand → Explain → Ask → Execute → Report
```

## Coding Flow

```text
Error / user request
        ↓
Intent Router
        ↓
Coding Worker
        ↓
Read minimal relevant context
        ↓
Compiler/LSP/tests + coding model
        ↓
Proposed change
        ↓
Human-readable explanation + diff
        ↓
User approval
        ↓
Apply change
        ↓
Run validation
        ↓
Report result
```

## Voice Flow

```text
Voice input
    ↓
Speech-to-text model
    ↓
Conversation Manager
    ↓
Intent Router
    ↓
Specialist / Tool
    ↓
Response text
    ↓
Text-to-speech model
```

Speech models are loaded only when voice is active.

## Privacy / Offline-First
Local models and local project data should remain on-device by default. External models or services should be explicit capabilities requiring user configuration and permission. Credentials must be kept in the operating system's secure credential store rather than normal model memory.

## Model Strategy
Start with existing small open models and local runtimes. Do not train a foundation model from scratch. A later training pipeline can use approved examples, evaluation data, and fine-tuning to improve specialist models.

## Hardware Target
Initial target: Apple Mac mini M4 with 16 GB unified memory and 256 GB SSD.

The design should favor small models, lazy loading, short model keep-alive periods, deterministic tools, and aggressive context minimization.
