import { HTMLMesh } from 'three/examples/jsm/interactive/HTMLMesh.js';
import { CanvasTexture, Mesh, MeshBasicMaterial, PerspectiveCamera, PlaneGeometry, Ray, Raycaster, SRGBColorSpace, Vector2, Vector3 } from 'three';

const PANEL_WIDTH_METERS = 2;
const PANEL_DISTANCE_METERS = 1.5;
const MIN_REDRAW_INTERVAL_MS = 50;
const BAR_WIDTH_METERS = 0.35;
const BAR_HEIGHT_METERS = 0.025;
const BAR_GAP_METERS = 0.04;
const BAR_IDLE_OPACITY = 0.5;
const MIN_MOVE_DISTANCE_METERS = 0.3;

interface RedrawableTexture extends CanvasTexture {
    update(): void;
}

const nativeValueSetter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!;
const nativeSelectIndexSetter = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'selectedIndex')!.set!;

enum MouseButton {
    Primary = 0,
    Secondary = 2,
}

const FOCUSABLE ='input, textarea, select, button, a[href], [tabindex]';

const svgImages = new Map<string, HTMLImageElement>();

function visibleRect(element: Element, rect: DOMRect): DOMRect | null {
    let left = rect.left, top = rect.top, right = rect.right, bottom = rect.bottom;
    for (let parent = element.parentElement; parent; parent = parent.parentElement) {
        const style = getComputedStyle(parent);
        if (style.overflowX === 'visible' && style.overflowY === 'visible') {
            continue;
        }
        const clip = parent.getBoundingClientRect();
        left = Math.max(left, clip.left);
        top = Math.max(top, clip.top);
        right = Math.min(right, clip.right);
        bottom = Math.min(bottom, clip.bottom);
        if (right <= left || bottom <= top) {
            return null;
        }
    }
    return new DOMRect(left, top, right - left, bottom - top);
}

// Input bypasses HTMLMesh's own dispatch: its events don't bubble, so React never sees them.
export class XRPagePanel {
    readonly mesh: HTMLMesh;

    #root: HTMLElement;
    #texture: RedrawableTexture;
    #labelCanvas: HTMLCanvasElement;
    #labelTexture: CanvasTexture;
    #pressed: HTMLElement | null = null;
    #hovered: HTMLElement | null = null;
    #redrawRequested = false;
    #lastRedraw = 0;
    #placed = false;
    #bar: Mesh;
    #barMaterial: MeshBasicMaterial;
    #moveDistance: number | null = null;
    #moveAnchor: Vector3 | null = null;
    #redraws = 0;
    #lastRedrawMs = 0;
    #worstRedrawMs = 0;

    constructor(root: HTMLElement) {
        this.#root = root;
        this.mesh = new HTMLMesh(root);
        this.mesh.name = "XR Page Panel";
        const nativeWidth = this.mesh.geometry.parameters.width;
        this.mesh.scale.setScalar(nativeWidth > 0 ? PANEL_WIDTH_METERS / nativeWidth : 1);

        this.#texture = this.mesh.material.map as RedrawableTexture;
        const update = this.#texture.update.bind(this.#texture);
        this.#texture.update = () => {
            const start = performance.now();
            update();
            this.#drawSvgs();
            this.#drawIconFonts();
            this.#lastRedraw = performance.now();
            this.#lastRedrawMs = this.#lastRedraw - start;
            this.#worstRedrawMs = Math.max(this.#worstRedrawMs, this.#lastRedrawMs);
            this.#redraws++;
            this.#drawLabel();
        };

        this.#labelCanvas = document.createElement('canvas');
        this.#labelCanvas.width = 512;
        this.#labelCanvas.height = 64;
        this.#labelTexture = new CanvasTexture(this.#labelCanvas);
        this.#labelTexture.colorSpace = SRGBColorSpace;
        const label = new Mesh(
            new PlaneGeometry(0.4, 0.05).scale(1 / this.mesh.scale.x, 1 / this.mesh.scale.y, 1),
            new MeshBasicMaterial({ map: this.#labelTexture, toneMapped: false })
        );
        label.position.y = (this.mesh.geometry.parameters.height / 2) + 0.03 / this.mesh.scale.y;
        this.mesh.add(label);
        this.#drawLabel();

        this.#barMaterial = new MeshBasicMaterial({ color: 0xd0d0d0, transparent: true, opacity: BAR_IDLE_OPACITY, toneMapped: false });
        this.#bar = new Mesh(
            new PlaneGeometry(BAR_WIDTH_METERS, BAR_HEIGHT_METERS).scale(1 / this.mesh.scale.x, 1 / this.mesh.scale.y, 1),
            this.#barMaterial
        );
        this.#bar.position.y = -(this.mesh.geometry.parameters.height / 2) - BAR_GAP_METERS / this.mesh.scale.y;
        this.mesh.add(this.#bar);

        this.#redrawRequested = true;
    }

    intersectBar(raycaster: Raycaster): { point: Vector3, distance: number } | null {
        const hits = raycaster.intersectObject(this.#bar, false);
        return hits.length > 0 ? { point: hits[0].point, distance: hits[0].distance } : null;
    }

    setBarHovered(hovered: boolean): void {
        this.#barMaterial.opacity = hovered || this.#moveDistance !== null ? 1 : BAR_IDLE_OPACITY;
    }

    get moving(): boolean {
        return this.#moveDistance !== null;
    }

    beginMove(point: Vector3, distance: number): void {
        this.#moveDistance = distance;
        this.#moveAnchor = this.mesh.worldToLocal(point.clone());
    }

    moveAlong(ray: Ray, eye: Vector3): void {
        if (this.#moveDistance === null || !this.#moveAnchor) {
            return;
        }
        const target = ray.at(this.#moveDistance, new Vector3());
        this.mesh.position.copy(target);
        this.mesh.lookAt(eye.x, target.y, eye.z);
        this.mesh.updateMatrixWorld();
        const anchor = this.#moveAnchor.clone().applyMatrix4(this.mesh.matrixWorld);
        this.mesh.position.add(target.sub(anchor));
    }

    pushMove(meters: number): void {
        if (this.#moveDistance !== null) {
            this.#moveDistance = Math.max(MIN_MOVE_DISTANCE_METERS, this.#moveDistance + meters);
        }
    }

    endMove(): void {
        this.#moveDistance = null;
        this.#moveAnchor = null;
        this.setBarHovered(false);
    }

    #drawLabel(): void {
        const ctx = this.#labelCanvas.getContext('2d')!;
        ctx.fillStyle = '#07212a';
        ctx.fillRect(0, 0, this.#labelCanvas.width, this.#labelCanvas.height);
        ctx.fillStyle = '#afb9bb';
        ctx.font = '28px monospace';
        ctx.fillText(
            `redraw ${this.#lastRedrawMs.toFixed(0)} ms  worst ${this.#worstRedrawMs.toFixed(0)}  n=${this.#redraws}`,
            12, 42
        );
        this.#labelTexture.needsUpdate = true;
    }

    #drawSvgs(): void {
        const canvas = this.#texture.image as HTMLCanvasElement;
        const ctx = canvas.getContext('2d');
        if (!ctx) {
            return;
        }
        const offset = this.#root.getBoundingClientRect();

        for (const svg of this.#root.querySelectorAll('svg')) {
            if (svg.parentElement?.closest('svg') || svg.getClientRects().length === 0) {
                continue;
            }
            const style = getComputedStyle(svg);
            if (style.visibility === 'hidden') {
                continue;
            }
            const rect = svg.getBoundingClientRect();
            if (rect.width === 0 || rect.height === 0) {
                continue;
            }
            const clip = visibleRect(svg, rect);
            if (!clip) {
                continue;
            }

            const image = this.#svgImage(svg, style.color, rect.width, rect.height);
            if (!image.complete || image.naturalWidth === 0) {
                continue;
            }

            ctx.save();
            ctx.globalAlpha = parseFloat(style.opacity) || 1;
            ctx.beginPath();
            ctx.rect(clip.left - offset.left, clip.top - offset.top, clip.width, clip.height);
            ctx.clip();
            ctx.drawImage(image, rect.left - offset.left, rect.top - offset.top, rect.width, rect.height);
            ctx.restore();
        }
        this.#texture.needsUpdate = true;
    }

    #drawIconFonts(): void {
        const canvas = this.#texture.image as HTMLCanvasElement;
        const ctx = canvas.getContext('2d');
        if (!ctx) {
            return;
        }
        const offset = this.#root.getBoundingClientRect();

        for (const icon of this.#root.querySelectorAll<HTMLElement>('i[class*="fa-"]')) {
            if (icon.getClientRects().length === 0) {
                continue;
            }
            const before = getComputedStyle(icon, '::before');
            const content = before.content;
            if (!content || content === 'none' || content === 'normal') {
                continue;
            }
            const glyph = /^"((?:[^"\\]|\\.)*)"/.exec(content)?.[1];
            const rect = icon.getBoundingClientRect();
            if (!glyph || rect.width === 0 || rect.height === 0 || getComputedStyle(icon).visibility === 'hidden') {
                continue;
            }
            const clip = visibleRect(icon, rect);
            if (!clip) {
                continue;
            }

            const font = `${before.fontStyle} ${before.fontWeight} ${before.fontSize} ${before.fontFamily}`;
            if (!document.fonts.check(font, glyph)) {
                document.fonts.load(font, glyph).then(() => this.#redrawRequested = true);
                continue;
            }

            ctx.save();
            ctx.beginPath();
            ctx.rect(clip.left - offset.left, clip.top - offset.top, clip.width, clip.height);
            ctx.clip();
            ctx.font = font;
            ctx.fillStyle = before.color;
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(glyph, rect.left - offset.left + rect.width / 2, rect.top - offset.top + rect.height / 2);
            ctx.restore();
        }
        this.#texture.needsUpdate = true;
    }

    #svgImage(svg: SVGSVGElement, color: string, width: number, height: number): HTMLImageElement {
        const clone = svg.cloneNode(true) as SVGSVGElement;
        clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
        clone.setAttribute('width', String(width));
        clone.setAttribute('height', String(height));
        clone.style.color = color;
        const markup = new XMLSerializer().serializeToString(clone).split('currentColor').join(color);

        let image = svgImages.get(markup);
        if (!image) {
            if (svgImages.size > 512) {
                svgImages.clear();
            }
            image = new Image();
            image.onload = () => this.#redrawRequested = true;
            image.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(markup);
            svgImages.set(markup, image);
        }
        return image;
    }

    requestRedraw(): void {
        this.#redrawRequested = true;
    }

    update(viewer: PerspectiveCamera): void {
        if (!this.#placed) {
            this.#placed = true;
            this.#placeInFrontOf(viewer);
        }

        if (this.#redrawRequested && performance.now() - this.#lastRedraw >= MIN_REDRAW_INTERVAL_MS) {
            this.#redrawRequested = false;
            this.#texture.update();
        }
    }

    #placeInFrontOf(viewer: PerspectiveCamera): void {
        const eye = new Vector3();
        viewer.getWorldPosition(eye);

        const forward = new Vector3();
        viewer.getWorldDirection(forward);
        forward.y = 0;
        if (forward.lengthSq() < 1e-6) {
            forward.set(0, 0, -1);
        }
        forward.normalize();

        this.mesh.position.copy(eye).addScaledVector(forward, PANEL_DISTANCE_METERS);
        this.mesh.lookAt(eye.x, this.mesh.position.y, eye.z);
    }

    intersect(raycaster: Raycaster): { uv: Vector2, distance: number } | null {
        const hits = raycaster.intersectObject(this.mesh, false);
        if (hits.length === 0 || !hits[0].uv) {
            return null;
        }
        return { uv: new Vector2(hits[0].uv.x, 1 - hits[0].uv.y), distance: hits[0].distance };
    }

    client(uv: Vector2): Vector2 {
        const rect = this.#root.getBoundingClientRect();
        return new Vector2(rect.left + uv.x * rect.width, rect.top + uv.y * rect.height);
    }

    pixelDistance(a: Vector2, b: Vector2): number {
        return this.client(a).distanceTo(this.client(b));
    }

    elementAt(uv: Vector2): HTMLElement {
        const point = this.client(uv);
        let hit: HTMLElement = this.#root;
        const visit = (element: HTMLElement): void => {
            const children = element.children;
            for (let i = children.length - 1; i >= 0; i--) {
                const child = children[i];
                if (!(child instanceof HTMLElement)) {
                    continue;
                }
                const style = getComputedStyle(child);
                if (style.display === 'none' || style.visibility === 'hidden' || style.pointerEvents === 'none') {
                    continue;
                }
                const rect = child.getBoundingClientRect();
                if (point.x >= rect.left && point.x <= rect.right && point.y >= rect.top && point.y <= rect.bottom) {
                    hit = child;
                    visit(child);
                    return;
                }
            }
        };
        visit(this.#root);
        return hit;
    }

    #send(target: HTMLElement, type: string, uv: Vector2, extra: MouseEventInit = {}): MouseEvent {
        const point = this.client(uv);
        const init: MouseEventInit = {
            clientX: point.x,
            clientY: point.y,
            bubbles: true,
            cancelable: true,
            composed: true,
            button: MouseButton.Primary,
            view: window,
            ...extra,
        };
        const event = type.startsWith('pointer')
            ? new PointerEvent(type, { ...init, pointerId: 1, pointerType: 'mouse', isPrimary: true })
            : new MouseEvent(type, init);
        target.dispatchEvent(event);
        return event;
    }

    #hover(target: HTMLElement, uv: Vector2): void {
        const previous = this.#hovered;
        if (previous === target) {
            return;
        }
        this.#hovered = target;
        if (previous) {
            this.#send(previous, 'pointerout', uv);
            this.#send(previous, 'mouseout', uv);
            this.#send(previous, 'mouseleave', uv, { bubbles: false });
        }
        this.#send(target, 'pointerover', uv);
        this.#send(target, 'mouseover', uv);
        this.#send(target, 'mouseenter', uv, { bubbles: false });
        this.#redrawRequested = true;
    }

    #setRange(input: HTMLInputElement, uv: Vector2): void {
        const rect = input.getBoundingClientRect();
        const t = Math.min(1, Math.max(0, (this.client(uv).x - rect.left) / rect.width));
        const min = parseFloat(input.min || '0');
        const max = parseFloat(input.max || '100');
        nativeValueSetter.call(input, String(min + (max - min) * t));
        input.dispatchEvent(new Event('input', { bubbles: true }));
    }

    move(uv: Vector2): void {
        const target = this.#pressed ?? this.elementAt(uv);
        if (!this.#pressed) {
            this.#hover(target, uv);
        }
        this.#send(target, 'pointermove', uv);
        this.#send(target, 'mousemove', uv);

        if (this.#pressed instanceof HTMLInputElement && this.#pressed.type === 'range') {
            this.#setRange(this.#pressed, uv);
        }
    }

    leave(): void {
        const previous = this.#hovered;
        if (!previous) {
            return;
        }
        this.#hovered = null;
        previous.dispatchEvent(new MouseEvent('mouseout', { bubbles: true }));
        previous.dispatchEvent(new MouseEvent('mouseleave'));
        this.#redrawRequested = true;
    }

    down(uv: Vector2): void {
        const target = this.elementAt(uv);
        this.#hover(target, uv);
        this.#pressed = target;

        const focusable = target.closest<HTMLElement>(FOCUSABLE);
        if (focusable) {
            if (document.activeElement !== focusable) {
                focusable.focus();
            }
        } else if (document.activeElement instanceof HTMLElement) {
            document.activeElement.blur();
        }

        this.#send(target, 'pointerdown', uv);
        this.#send(target, 'mousedown', uv);

        if (target instanceof HTMLInputElement && target.type === 'range') {
            this.#setRange(target, uv);
        }
        this.#redrawRequested = true;
    }

    up(uv: Vector2 | null): void {
        const pressed = this.#pressed;
        this.#pressed = null;
        if (!pressed) {
            return;
        }

        const target = uv ? this.elementAt(uv) : pressed;
        const at = uv ?? new Vector2();
        this.#send(target, 'pointerup', at);
        this.#send(target, 'mouseup', at);

        if (pressed instanceof HTMLInputElement && pressed.type === 'range') {
            pressed.dispatchEvent(new Event('change', { bubbles: true }));
        }

        if (uv && (target === pressed || pressed.contains(target))) {
            if (target instanceof HTMLSelectElement) {
                if (target.options.length > 0) {
                    nativeSelectIndexSetter.call(target, (target.selectedIndex + 1) % target.options.length);
                    target.dispatchEvent(new Event('change', { bubbles: true }));
                }
            } else {
                this.#send(target, 'click', uv);
            }
        }
        this.#redrawRequested = true;
    }

    cancelPress(): void {
        if (this.#pressed) {
            this.#pressed.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, button: MouseButton.Primary }));
            this.#pressed = null;
        }
    }

    contextMenu(uv: Vector2): void {
        this.#send(this.elementAt(uv), 'contextmenu', uv, { button: MouseButton.Secondary });
        this.#redrawRequested = true;
    }

    wheel(uv: Vector2, deltaY: number): void {
        const target = this.elementAt(uv);
        const point = this.client(uv);
        const event = new WheelEvent('wheel', {
            clientX: point.x,
            clientY: point.y,
            deltaY,
            deltaMode: WheelEvent.DOM_DELTA_PIXEL,
            bubbles: true,
            cancelable: true,
        });
        target.dispatchEvent(event);
        if (event.defaultPrevented) {
            return;
        }

        let element: HTMLElement | null = target;
        while (element) {
            if (element.scrollHeight > element.clientHeight) {
                const overflow = getComputedStyle(element).overflowY;
                if (overflow === 'auto' || overflow === 'scroll') {
                    element.scrollTop += deltaY;
                    this.#redrawRequested = true;
                    return;
                }
            }
            element = element.parentElement;
        }
    }

    dispose(): void {
        this.cancelPress();
        this.leave();
        this.mesh.removeFromParent();
        this.mesh.traverse((child) => {
            if (child instanceof Mesh && child !== this.mesh) {
                child.geometry.dispose();
                child.material.dispose();
            }
        });
        this.#labelTexture.dispose();
        this.mesh.dispose();
    }
}
