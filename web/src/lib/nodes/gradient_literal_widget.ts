import { GlobalWidgetFactory, InBox, RowWidget, Theme, WidgetBase, type Box, type FlowNode, type Vector2, type Widget, type WidgetPointer } from '@elicdavis/node-flow';
import { opaque } from '../color_alpha';
import { InsertRuleWidget, RemoveWidget } from './list_widgets';
import type { GradientKey } from '../../types/parameter';

export interface GradientLiteralConfig {
    property: string;
    value?: unknown;
}

const BAR_HEIGHT = 14;
const WIDTH = 190;

function keysOf(value: unknown): Array<GradientKey> {
    const keys = (value as { keys?: Array<GradientKey> } | undefined)?.keys;
    if (!Array.isArray(keys)) {
        return [];
    }
    return keys.map((k) => ({ time: Number(k.time), value: String(k.value) })).sort((a, b) => a.time - b.time);
}

function sameKeys(a: Array<GradientKey>, b: Array<GradientKey>): boolean {
    return a.length === b.length && a.every((k, i) => k.time === b[i].time && k.value === b[i].value);
}

export class GradientLiteralWidget extends WidgetBase {

    #node: FlowNode;

    #property: string;

    #keys: Array<GradientKey>;

    #children: Array<Widget> = [];

    #boxes: Array<Box> = [];

    #captured: Widget | null = null;

    #capturedBox: Box | null = null;

    #mouse: Vector2 | undefined;

    constructor(node: FlowNode, config: GradientLiteralConfig) {
        super();
        this.#node = node;
        this.#property = config.property;
        this.#keys = keysOf(node.getProperty(config.property) ?? config.value);
        node.setProperty(this.#property, { keys: this.#keys });
        node.addPropertyChangeListener(this.#property, (_, next) => {
            const keys = keysOf(next);
            if (!sameKeys(keys, this.#keys)) {
                this.#keys = keys;
                this.#build();
            }
        });
        this.#build();
    }

    Size(): Vector2 {
        let y = BAR_HEIGHT;
        for (const child of this.#children) {
            y += Theme.Widget.Gap + child.Size().y;
        }
        return { x: WIDTH, y };
    }

    Draw(ctx: CanvasRenderingContext2D, box: Box, scale: number, mousePosition: Vector2 | undefined): void {
        this.#mouse = mousePosition;
        super.Draw(ctx, box, scale, mousePosition);
    }

    protected render(ctx: CanvasRenderingContext2D, box: Box, scale: number): void {
        const bar: Box = { Position: { ...box.Position }, Size: { x: box.Size.x, y: BAR_HEIGHT * scale } };
        this.#drawBar(ctx, bar, scale);

        this.#boxes = [];
        let y = box.Position.y + bar.Size.y;
        for (const child of this.#children) {
            y += Theme.Widget.Gap * scale;
            const childBox: Box = { Position: { x: box.Position.x, y }, Size: { x: box.Size.x, y: child.Size().y * scale } };
            this.#boxes.push(childBox);
            child.Draw(ctx, childBox, scale, this.#captured === null || this.#captured === child ? this.#mouse : undefined);
            y += childBox.Size.y;
        }
    }

    #drawBar(ctx: CanvasRenderingContext2D, bar: Box, scale: number): void {
        const fill = ctx.createLinearGradient(bar.Position.x, 0, bar.Position.x + bar.Size.x, 0);
        if (this.#keys.length === 0) {
            fill.addColorStop(0, "#000000");
            fill.addColorStop(1, "#000000");
        }
        for (const key of this.#keys) {
            fill.addColorStop(Math.min(Math.max(key.time, 0), 1), opaque(key.value));
        }
        ctx.fillStyle = fill;
        ctx.fillRect(bar.Position.x, bar.Position.y, bar.Size.x, bar.Size.y);
        ctx.strokeStyle = "#00000088";
        ctx.lineWidth = Math.max(1, scale);
        ctx.strokeRect(bar.Position.x, bar.Position.y, bar.Size.x, bar.Size.y);
    }

    protected onPointerDown(pointer: WidgetPointer): void {
        const index = this.#boxes.findIndex((b) => InBox(b, pointer.position));
        if (index < 0 || this.#children[index].Disabled?.()) {
            return;
        }
        this.#captured = this.#children[index];
        this.#capturedBox = this.#boxes[index];
        this.#captured.PointerDown?.(this.#forChild(pointer));
    }

    protected onPointerMove(pointer: WidgetPointer): void {
        this.#captured?.PointerMove?.(this.#forChild(pointer));
    }

    protected onPointerUp(pointer: WidgetPointer): void {
        const child = this.#captured;
        const forChild = this.#forChild(pointer);
        this.#captured = null;
        this.#capturedBox = null;
        child?.PointerUp?.(forChild);
    }

    #forChild(pointer: WidgetPointer): WidgetPointer {
        return { position: pointer.position, inside: this.#capturedBox !== null && InBox(this.#capturedBox, pointer.position) };
    }

    #build(): void {
        this.#children = [];
        if (this.#keys.length === 0) {
            this.#children.push(new InsertRuleWidget(() => this.#commit([{ time: 0, value: "#000000" }, { time: 1, value: "#ffffff" }])));
            return;
        }

        const pinned = (i: number) => this.#keys.length > 1 && (i === 0 || i === this.#keys.length - 1);
        const replace = (index: number, patch: Partial<GradientKey>) =>
            this.#commit(this.#keys.map((k, i) => (i === index ? { ...k, ...patch } : k)));

        this.#keys.forEach((key, index) => {
            if (index > 0) {
                this.#children.push(new InsertRuleWidget(() => this.#insertBefore(index)));
            }
            const color = GlobalWidgetFactory.create(this.#node, "color", {
                value: opaque(key.value),
                width: 90,
                callback: (hex: string) => replace(index, { value: hex + key.value.slice(7) }),
            });
            const time = GlobalWidgetFactory.create(this.#node, "number", {
                value: key.time,
                label: "t",
                width: 64,
                min: 0,
                max: 1,
                step: 0.005,
                callback: (t: number) => replace(index, { time: t }),
            });
            if (pinned(index)) {
                time.setDisabled(true);
            }
            const remove = new RemoveWidget(() => this.#commit(this.#keys.filter((_, i) => i !== index)));
            this.#children.push(new RowWidget([color, time, remove], { grow: [1, 0, 0] }));
        });
    }

    #insertBefore(index: number): void {
        const left = this.#keys[index - 1];
        const right = this.#keys[index];
        const next = [...this.#keys];
        next.splice(index, 0, { time: (left.time + right.time) / 2, value: left.value });
        this.#commit(next);
    }

    #commit(keys: Array<GradientKey>): void {
        this.#keys = [...keys].sort((a, b) => a.time - b.time);
        this.#build();
        this.#node.editProperty(this.#property, { keys: this.#keys });
    }
}

GlobalWidgetFactory.register("gradient", (node: FlowNode, config?: GradientLiteralConfig) =>
    new GradientLiteralWidget(node, config ?? { property: "gradient" }));
