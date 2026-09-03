/** Consume Sarathi SSE voice stream (POST /api/voice/stream). */

export type StreamHandlers = {
  onStatus?: (phase: string) => void;
  onTranscript?: (raw: string, normalized: string) => void;
  onToken?: (text: string) => void;
  onSentence?: (text: string, audioUrl: string) => void;
  onDone?: (replyText: string) => void;
  onError?: (message: string) => void;
};

export async function streamVoice(
  body: FormData | Record<string, unknown>,
  handlers: StreamHandlers,
): Promise<void> {
  const init: RequestInit =
    body instanceof FormData
      ? { method: "POST", body }
      : {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        };

  const res = await fetch("/api/voice/stream", init);
  if (!res.ok || !res.body) {
    let message = `HTTP ${res.status}`;
    try {
      const data = (await res.json()) as { error?: string };
      if (data.error) message = data.error;
    } catch {
      /* ignore */
    }
    throw new Error(message);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    let splitAt: number;
    while ((splitAt = buffer.indexOf("\n\n")) >= 0) {
      const chunk = buffer.slice(0, splitAt);
      buffer = buffer.slice(splitAt + 2);
      dispatchSSE(chunk, handlers);
    }
  }
  if (buffer.trim()) {
    dispatchSSE(buffer, handlers);
  }
}

function dispatchSSE(raw: string, handlers: StreamHandlers) {
  let event = "message";
  const dataLines: string[] = [];
  for (const line of raw.split("\n")) {
    if (line.startsWith("event:")) {
      event = line.slice(6).trim();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice(5).trim());
    }
  }
  if (!dataLines.length) return;
  const payload = dataLines.join("\n");
  let data: Record<string, string> = {};
  try {
    data = JSON.parse(payload) as Record<string, string>;
  } catch {
    return;
  }

  switch (event) {
    case "status":
      handlers.onStatus?.(data.phase || "");
      break;
    case "transcript":
      handlers.onTranscript?.(data.raw_transcript || "", data.normalized_text || "");
      break;
    case "token":
      handlers.onToken?.(data.text || "");
      break;
    case "sentence":
      handlers.onSentence?.(data.text || "", data.audio_url || "");
      break;
    case "done":
      handlers.onDone?.(data.reply_text || "");
      break;
    case "error":
      handlers.onError?.(data.error || "stream error");
      break;
  }
}

/** Play sentence WAVs in order without overlapping. */
export function createAudioQueue() {
  const queue: string[] = [];
  let playing = false;

  const pump = async () => {
    if (playing) return;
    playing = true;
    while (queue.length) {
      const url = queue.shift()!;
      await playOnce(url);
    }
    playing = false;
  };

  return {
    enqueue(audioUrl: string) {
      queue.push(`${audioUrl}?t=${Date.now()}`);
      void pump();
    },
  };
}

function playOnce(url: string): Promise<void> {
  return new Promise((resolve) => {
    const audio = new Audio(url);
    audio.onended = () => resolve();
    audio.onerror = () => resolve();
    void audio.play().catch(() => resolve());
  });
}
