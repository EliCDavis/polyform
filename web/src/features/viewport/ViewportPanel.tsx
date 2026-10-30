import { ModelStats } from "./ModelStats";
import { AnimationControls } from "./AnimationControls";

export function ViewportPanel() {
  return (
    <div id="three-viewer-container">
      <canvas id="three-canvas" />
      <div id="running-message" role="status">
        <span className="running-spinner" aria-hidden="true" />
        <span>Processing</span>
        <span className="running-elapsed" />
      </div>
      <ModelStats />
      <AnimationControls />
    </div>
  );
}
