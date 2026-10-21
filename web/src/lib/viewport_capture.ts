import { requestRender } from "./render_scheduler";

export type ViewportCapture = (canvas: HTMLCanvasElement) => void;

const pending: ViewportCapture[] = [];

// The canvas only holds a frame until it is composited, so callbacks must
// read it synchronously.
export function captureViewport(capture: ViewportCapture): void {
  pending.push(capture);
  requestRender();
}

export function hasViewportCaptures(): boolean {
  return pending.length > 0;
}

export function flushViewportCaptures(canvas: HTMLCanvasElement): void {
  for (const capture of pending.splice(0)) {
    try {
      capture(canvas);
    } catch (err) {
      console.error("viewport capture failed", err);
    }
  }
}
