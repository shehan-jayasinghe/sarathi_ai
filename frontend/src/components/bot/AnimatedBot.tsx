import bustSvg from "../../assets/images/sarathi-bust.svg";
import "./AnimatedBot.css";

export function AnimatedBot() {
  return (
    <div className="animated-bot">
      <img
        className="animated-bot__image"
        src={bustSvg}
        alt="Sarathi AI agent"
        draggable={false}
      />
    </div>
  );
}
