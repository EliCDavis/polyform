import {
    CursorStyle,
    FlowNode,
    Theme,
    ValueWidget,
    drawField,
    requestRender,
    widgetFont,
    type Box,
    type ValueWidgetConfig,
} from "@elicdavis/node-flow";

export interface SelectWidgetConfig extends ValueWidgetConfig<string> {
    options: string[];
}

export class SelectWidget extends ValueWidget<string> {

    #options: string[];

    #menu: HTMLDivElement | null = null;

    constructor(node: FlowNode, config: SelectWidgetConfig) {
        super(node, config, config.options[0] ?? "");
        this.#options = config.options;
    }

    Cursor(): CursorStyle {
        return CursorStyle.Pointer;
    }

    protected render(ctx: CanvasRenderingContext2D, box: Box, scale: number): void {
        drawField(ctx, box, scale, {
            hovered: this.hovered() || this.#menu !== null,
            pressed: this.pressed(),
            label: this.label(),
            value: this.value(),
        });

        ctx.font = widgetFont(scale);
        ctx.fillStyle = Theme.Widget.LabelColor;
        ctx.textAlign = "right";
        ctx.textBaseline = "middle";
        ctx.fillText(
            "▾",
            box.Position.x + box.Size.x - Theme.Widget.Padding * scale,
            box.Position.y + box.Size.y / 2,
        );
    }

    protected onClick(): void {
        if (this.#menu !== null) {
            this.#close();
            return;
        }
        this.#open();
    }

    #open(): void {
        const canvas = this.canvas();
        if (canvas === null) {
            return;
        }

        const box = this.box();
        const scale = this.scale();
        const rect = canvas.getBoundingClientRect();
        const sx = canvas.width > 0 ? rect.width / canvas.width : 1;
        const sy = canvas.height > 0 ? rect.height / canvas.height : 1;
        const left = rect.left + box.Position.x * sx;
        const top = rect.top + (box.Position.y + box.Size.y) * sy;
        const fontSize = Theme.Widget.FontSize * scale * sy;
        const padding = Theme.Widget.Padding * scale * sx;

        const menu = document.createElement("div");
        menu.style.cssText = [
            "position: fixed",
            `left: ${left}px`,
            `top: ${top}px`,
            `min-width: ${box.Size.x * sx}px`,
            `max-height: ${Math.max(0, rect.bottom - top)}px`,
            "overflow-y: auto",
            "box-sizing: border-box",
            `font: ${fontSize}px ${Theme.FontFamily}`,
            `color: ${Theme.Widget.FontColor}`,
            `background: ${Theme.Widget.BackgroundColor}`,
            `border: ${Theme.Widget.Border.Size * scale}px solid ${Theme.Widget.Border.Color}`,
            `border-radius: ${Theme.Widget.Border.Radius * scale}px`,
            "z-index: 1000",
        ].join(";");

        const before = this.value();
        for (const option of this.#options) {
            const item = document.createElement("div");
            item.textContent = option;
            item.style.cssText = [
                `padding: ${padding / 2}px ${padding}px`,
                "cursor: pointer",
                "white-space: nowrap",
                option === before ? `color: ${Theme.Widget.Input.FontColor}` : "",
            ].join(";");
            item.addEventListener("mouseenter", () => {
                item.style.background = Theme.Widget.Hover.BackgroundColor;
            });
            item.addEventListener("mouseleave", () => {
                item.style.background = "";
            });
            item.addEventListener("click", (event) => {
                event.stopPropagation();
                this.#close();
                this.commit(option, before);
            });
            menu.appendChild(item);
        }

        const onKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                event.preventDefault();
                event.stopPropagation();
                this.#close();
            }
        };
        const onPointerDown = (event: PointerEvent) => {
            if (!menu.contains(event.target as Node)) {
                this.#close();
            }
        };
        document.addEventListener("keydown", onKeyDown, true);
        document.addEventListener("pointerdown", onPointerDown, true);

        this.#menu = menu;
        this.#closeMenu = () => {
            document.removeEventListener("keydown", onKeyDown, true);
            document.removeEventListener("pointerdown", onPointerDown, true);
            menu.remove();
        };
        document.body.appendChild(menu);
        requestRender();
    }

    #closeMenu: (() => void) | null = null;

    #close(): void {
        this.#closeMenu?.();
        this.#closeMenu = null;
        this.#menu = null;
        requestRender();
    }
}
