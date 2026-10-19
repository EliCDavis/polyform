import { ModelStats } from "./ModelStats";
import { AnimationControls } from "./AnimationControls";

export function ViewportPanel() {
  return (
    <div id="three-viewer-container">
      <canvas id="three-canvas" />
      <ModelStats />
      <AnimationControls />
    </div>
  );
}
