import { AnimatedBot } from "../components/bot/AnimatedBot";
import "./AgentOverlay.css";

/** Floating desktop agent — transparent canvas, small bot top-right only. */
export function AgentOverlay() {
  return (
    <main className="agent-overlay">
      <div className="agent-overlay__slot">
        <AnimatedBot />
      </div>
    </main>
  );
}
