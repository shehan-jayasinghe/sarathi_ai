/** Record microphone audio until stopped; returns a Blob (webm/ogg). */
export async function recordMicrophone(maxMs = 5000): Promise<Blob> {
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error("Microphone not available in this browser");
  }

  const stream = await navigator.mediaDevices.getUserMedia({
    audio: {
      echoCancellation: true,
      noiseSuppression: true,
    },
  });

  const mimeType = pickMimeType();
  const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
  const chunks: BlobPart[] = [];

  return new Promise<Blob>((resolve, reject) => {
    const stopTracks = () => {
      for (const track of stream.getTracks()) {
        track.stop();
      }
    };

    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        chunks.push(event.data);
      }
    };

    recorder.onerror = () => {
      stopTracks();
      reject(new Error("Microphone recording failed"));
    };

    recorder.onstop = () => {
      stopTracks();
      const type = recorder.mimeType || mimeType || "audio/webm";
      resolve(new Blob(chunks, { type }));
    };

    recorder.start();
    window.setTimeout(() => {
      if (recorder.state === "recording") {
        recorder.stop();
      }
    }, maxMs);
  });
}

function pickMimeType(): string | undefined {
  const candidates = [
    "audio/webm;codecs=opus",
    "audio/webm",
    "audio/ogg;codecs=opus",
    "audio/mp4",
  ];
  for (const type of candidates) {
    if (MediaRecorder.isTypeSupported(type)) {
      return type;
    }
  }
  return undefined;
}

export function extensionForBlob(blob: Blob): string {
  if (blob.type.includes("ogg")) return ".ogg";
  if (blob.type.includes("mp4") || blob.type.includes("m4a")) return ".m4a";
  return ".webm";
}
