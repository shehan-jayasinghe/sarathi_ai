import { Lottie } from "lottie-react";
import robotAnimation from "../../assets/images/sarathi-robot.json";
import "./AnimatedBot.css";

export function AnimatedBot() {
  return (
    <div className="animated-bot">
      <Lottie
        className="animated-bot__lottie"
        src={robotAnimation}
        loop
        autoplay
        aria-label="Sarathi AI agent"
      />
    </div>
  );
}
