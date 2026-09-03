import { useCallback, useRef, useState } from "react";
import { AnimatedBot } from "../components/bot/AnimatedBot";
import { extensionForBlob, recordMicrophone } from "../hooks/useMicRecorder";
import { createAudioQueue, streamVoice } from "../hooks/useVoiceStream";
import "./AgentOverlay.css";

/** Floating desktop agent — transparent canvas, small bot top-right only. */
export function AgentOverlay() {
  const [busy, setBusy] = useState(false);
  const [listening, setListening] = useState(false);
  const [status, setStatus] = useState<string>("");
  const [heard, setHeard] = useState<string>("");
  const [reply, setReply] = useState<string>("");
  const audioQueue = useRef(createAudioQueue());

  const runVoice = useCallback(async () => {
    if (busy) return;
    setBusy(true);
    setReply("");
    setHeard("");
    setStatus("Listening… speak now");
    setListening(true);
    audioQueue.current = createAudioQueue();

    try {
      const blob = await recordMicrophone(5000);
      setListening(false);
      setStatus("Transcribing…");

      const form = new FormData();
      form.append("audio", blob, `mic${extensionForBlob(blob)}`);

      await streamVoice(form, {
        onStatus: (phase) => {
          if (phase === "transcribing") setStatus("Transcribing…");
          if (phase === "thinking") setStatus("Thinking…");
        },
        onTranscript: (_raw, normalized) => {
          setHeard(normalized || _raw);
          setStatus("Streaming reply…");
        },
        onToken: (text) => {
          setReply((prev) => prev + text);
        },
        onSentence: (_text, audioUrl) => {
          if (audioUrl) {
            audioQueue.current.enqueue(audioUrl);
          }
        },
        onDone: (full) => {
          setReply(full);
          setStatus("Done");
        },
        onError: (message) => {
          setStatus(message);
        },
      });
    } catch (err) {
      setListening(false);
      setStatus(err instanceof Error ? err.message : "Voice request failed");
    } finally {
      setBusy(false);
      setListening(false);
    }
  }, [busy]);

  return (
    <main className="agent-overlay">
      <div className={`agent-overlay__slot${listening ? " is-listening" : ""}`}>
        <button
          type="button"
          className="agent-overlay__hit"
          onClick={runVoice}
          disabled={busy}
          aria-label="Talk to Sarathi"
          title="Click, then speak for 5 seconds"
        >
          <AnimatedBot />
        </button>
      </div>
      {(status || heard || reply) && (
        <div className="agent-overlay__toast" role="status">
          {status && <p className="agent-overlay__status">{status}</p>}
          {heard && (
            <p className="agent-overlay__heard">
              <span>You:</span> {heard}
            </p>
          )}
          {reply && (
            <p className="agent-overlay__reply">
              <span>Sarathi:</span> {reply}
            </p>
          )}
        </div>
      )}
    </main>
  );
}
