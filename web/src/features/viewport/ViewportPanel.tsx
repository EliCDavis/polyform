import { ModelStats } from "./ModelStats";

export function ViewportPanel() {
  return (
    <div id="three-viewer-container">
      <canvas id="three-canvas" />
      <ModelStats />
    </div>
  );
}
