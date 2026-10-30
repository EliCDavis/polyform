import { Theme, WidgetBase, type Box, type Vector2 } from '@elicdavis/node-flow';

const ACCENT = "#ffc95c";

// A thin rule with a plus in the middle, faint until hovered.
export class InsertRuleWidget extends WidgetBase {

    constructor(private readonly onInsert: () => void) {
        super();
    }

    Size(): Vector2 {
        return { x: 150, y: 12 };
    }

    protected render(ctx: CanvasRenderingContext2D, box: Box, scale: number): void {
        const midY = box.Position.y + box.Size.y / 2;
        const centerX = box.Position.x + box.Size.x / 2;
        const gap = 8 * scale;

        ctx.globalAlpha = this.hovered() ? 1 : 0.3;
        ctx.strokeStyle = ACCENT;
        ctx.fillStyle = ACCENT;
        ctx.lineWidth = Math.max(1, scale);
        ctx.beginPath();
        ctx.moveTo(box.Position.x, midY);
        ctx.lineTo(centerX - gap, midY);
        ctx.moveTo(centerX + gap, midY);
        ctx.lineTo(box.Position.x + box.Size.x, midY);
        ctx.stroke();
        ctx.font = `${11 * scale}px ${Theme.FontFamily}`;
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText("+", centerX, midY);
    }

    protected onClick(): void {
        this.onInsert();
    }
}

// A small ring with a minus, red when hovered.
export class RemoveWidget extends WidgetBase {

    constructor(private readonly onRemove: () => void) {
        super();
    }

    Size(): Vector2 {
        return { x: 18, y: 18 };
    }

    protected render(ctx: CanvasRenderingContext2D, box: Box, scale: number): void {
        const size = Math.min(box.Size.x, box.Size.y);
        const cx = box.Position.x + box.Size.x - size / 2;
        const cy = box.Position.y + box.Size.y / 2;
        const hovered = this.hovered();

        ctx.beginPath();
        ctx.arc(cx, cy, size / 2 - scale, 0, Math.PI * 2);
        if (hovered) {
            ctx.fillStyle = "#7f1d1d";
            ctx.fill();
        }
        ctx.strokeStyle = hovered ? "#7f1d1d" : Theme.Widget.Border.Color;
        ctx.lineWidth = Math.max(1, scale);
        ctx.stroke();
        ctx.fillStyle = hovered ? "#ffffff" : Theme.Widget.FontColor;
        ctx.font = `${12 * scale}px ${Theme.FontFamily}`;
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText("−", cx, cy);
    }

    protected onClick(): void {
        this.onRemove();
    }
}
