#!/usr/bin/env python3
"""CLI for Sarathi voice pipeline.

Examples:
  python pipeline_cli.py --text "Open file number twenty five and change line ten."
  python pipeline_cli.py --audio ./clip.wav --out ./reply.wav
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

# Allow running as: python pipeline_cli.py from tools/voice/
sys.path.insert(0, str(Path(__file__).resolve().parent))

from sarathi_voice.pipeline import run_pipeline


def main() -> int:
    parser = argparse.ArgumentParser(description="Sarathi voice pipeline")
    src = parser.add_mutually_exclusive_group(required=True)
    src.add_argument("--audio", type=Path, help="Input WAV/audio for Whisper STT")
    src.add_argument("--text", type=str, help="Skip STT; feed text to normalizer + Qwen")
    parser.add_argument(
        "--out",
        type=Path,
        default=Path(__file__).resolve().parents[2] / "data" / "voice" / "reply.wav",
        help="Output WAV path for Kokoro TTS",
    )
    parser.add_argument("--skip-tts", action="store_true", help="Only print Qwen reply")
    parser.add_argument("--whisper-model", default="mlx-community/whisper-base-mlx")
    parser.add_argument("--ollama-model", default="qwen2.5:3b")
    parser.add_argument("--voice", default="af_heart", help="Kokoro voice id")
    args = parser.parse_args()

    result = run_pipeline(
        audio_path=args.audio,
        text=args.text,
        output_audio=args.out,
        skip_tts=args.skip_tts,
        whisper_model=args.whisper_model,
        ollama_model=args.ollama_model,
        voice=args.voice,
    )

    print("=== Whisper (raw) ===")
    print(result.raw_transcript)
    print("=== Normalized ===")
    print(result.normalized_text)
    print("=== Qwen ===")
    print(result.reply_text)
    if result.output_audio:
        print("=== Kokoro ===")
        print(result.output_audio)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
