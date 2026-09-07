import { Observable, Subject } from "rxjs";
import {
    BufferAttribute,
    BufferGeometry,
    Group,
    Line,
    LineBasicMaterial,
    Mesh,
    MeshBasicMaterial,
    PerspectiveCamera,
    Scene,
    SphereGeometry,
    Vector3,
} from "three";
import { CSS2DObject } from "three/examples/jsm/renderers/CSS2DRenderer.js";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { TransformGizmo } from "./transform";

export interface PointPathPoint {
    x: number;
    y: number;
    z: number;
}

export interface PointPathGizmoConfig {
    camera: PerspectiveCamera;
    parent: Group;
    domElement: HTMLElement;
    scene: Scene;
    orbitControls: OrbitControls;
    points?: Array<PointPathPoint>;
    closed?: boolean;
}

export interface PointPathChange {
    index: number;
    position: PointPathPoint;
}

const LINE_COLOR = 0x2fd6d6;
const MARKER_COLOR = 0x2fd6d6;
const FOCUSED_MARKER_COLOR = 0xffc95c;
const PREVIEW_COLOR = 0xffc95c;

/**
 * Where a point inserted at `index` lands: between its neighbours, or carried
 * on past the end it was added to so it never spawns inside an existing point.
 */
export function pathInsertionPoint(
    points: ReadonlyArray<PointPathPoint>,
    index: number
): PointPathPoint {
    const before = points[index - 1];
    const after = points[index];

    if (before && after) {
        return {
            x: (before.x + after.x) / 2,
            y: (before.y + after.y) / 2,
            z: (before.z + after.z) / 2,
        };
    }

    const anchor = before ?? after;
    if (!anchor) {
        return { x: 0, y: 0, z: 0 };
    }

    const neighbour = before ? points[index - 2] : points[index + 1];
    const step = neighbour
        ? { x: anchor.x - neighbour.x, y: anchor.y - neighbour.y, z: anchor.z - neighbour.z }
        : { x: before ? 1 : -1, y: 0, z: 0 };

    if (Math.hypot(step.x, step.y, step.z) < 1e-6) {
        step.x = before ? 1 : -1;
        step.y = 0;
        step.z = 0;
    }

    return { x: anchor.x + step.x, y: anchor.y + step.y, z: anchor.z + step.z };
}

interface PointEntry {
    gizmo: TransformGizmo;
    marker: Mesh;
    material: MeshBasicMaterial;
    label: CSS2DObject;
    labelElement: HTMLDivElement;
}

export class PointPathGizmo {

    private config: PointPathGizmoConfig;

    private entries: Array<PointEntry>;

    private line: Line;

    private lineGeometry: BufferGeometry;

    private group: Group;

    private change$: Subject<PointPathChange>;

    private focus: number | null;

    private enabled: boolean;

    private closed: boolean;

    private preview: Mesh;

    private previewMaterial: MeshBasicMaterial;

    private previewIndex: number | null;

    constructor(config: PointPathGizmoConfig) {
        this.config = config;
        this.entries = [];
        this.change$ = new Subject<PointPathChange>();
        this.focus = null;
        this.enabled = true;
        this.closed = config.closed === true;

        this.group = new Group();
        config.parent.add(this.group);

        this.lineGeometry = new BufferGeometry();
        this.lineGeometry.setAttribute("position", new BufferAttribute(new Float32Array(0), 3));
        this.line = new Line(
            this.lineGeometry,
            new LineBasicMaterial({ color: LINE_COLOR, toneMapped: false, depthTest: false })
        );
        this.line.renderOrder = 1;
        this.line.frustumCulled = false;
        this.group.add(this.line);

        this.previewIndex = null;
        this.previewMaterial = new MeshBasicMaterial({
            color: PREVIEW_COLOR,
            toneMapped: false,
            depthTest: false,
            transparent: true,
            opacity: 0.55,
        });
        this.preview = new Mesh(new SphereGeometry(1, 16, 12), this.previewMaterial);
        this.preview.renderOrder = 3;
        this.preview.visible = false;
        this.group.add(this.preview);

        this.setPoints(config.points ?? []);
    }

    setPoints(points: Array<PointPathPoint>): void {
        while (this.entries.length > points.length) {
            this.disposeEntry(this.entries.pop());
        }

        while (this.entries.length < points.length) {
            this.entries.push(this.newEntry(this.entries.length, points[this.entries.length]));
        }

        points.forEach((point, i) => {
            const entry = this.entries[i];
            entry.gizmo.setPosition(point.x, point.y, point.z);
            entry.marker.position.set(point.x, point.y, point.z);
        });

        this.resizeMarkers();
        this.redraw();
        this.applyVisibility();
        this.positionPreview();
    }

    /**
     * Isolates a single point: every other gizmo is hidden so the one being
     * inspected can be grabbed without its neighbours in the way.
     */
    setFocus(index: number | null): void {
        this.focus = index;
        this.applyVisibility();
    }

    /**
     * Ghosts the point a call to insert at `index` would produce, so the gap
     * being pointed at in the sidebar is identifiable in the viewport.
     */
    setInsertPreview(index: number | null): void {
        this.previewIndex = index;
        this.positionPreview();
    }

    setEnabled(enabled: boolean): void {
        this.enabled = enabled;
        this.applyVisibility();
    }

    setClosed(closed: boolean): void {
        this.closed = closed;
        this.redraw();
    }

    changes$(): Observable<PointPathChange> {
        return this.change$.asObservable();
    }

    dispose(): void {
        while (this.entries.length > 0) {
            this.disposeEntry(this.entries.pop());
        }
        this.change$.complete();
        this.preview.geometry.dispose();
        this.previewMaterial.dispose();
        this.lineGeometry.dispose();
        (this.line.material as LineBasicMaterial).dispose();
        this.group.removeFromParent();
    }

    private newEntry(index: number, point: PointPathPoint): PointEntry {
        const gizmo = new TransformGizmo({
            camera: this.config.camera,
            domElement: this.config.domElement,
            orbitControls: this.config.orbitControls,
            parent: this.config.parent,
            scene: this.config.scene,
            initialPosition: point,
        });

        const material = new MeshBasicMaterial({
            color: MARKER_COLOR,
            toneMapped: false,
            depthTest: false,
            transparent: true,
        });
        const marker = new Mesh(new SphereGeometry(1, 16, 12), material);
        marker.renderOrder = 2;
        marker.position.set(point.x, point.y, point.z);
        this.group.add(marker);

        const labelElement = document.createElement("div");
        labelElement.textContent = String(index);
        labelElement.style.cssText = [
            "color: #d8f4f4",
            "background: rgba(0, 32, 40, 0.75)",
            "border-radius: 4px",
            "padding: 0 4px",
            "font-family: monospace",
            "font-size: 11px",
            "pointer-events: none",
            "user-select: none",
        ].join(";");
        const label = new CSS2DObject(labelElement);
        label.center.set(0, 1.4);
        marker.add(label);

        const entry: PointEntry = { gizmo, marker, material, label, labelElement };

        gizmo.dragging$().subscribe((position) => {
            marker.position.copy(position);
            this.redraw();
        });

        gizmo.position$().subscribe((position) => {
            marker.position.copy(position);
            this.redraw();
            this.change$.next({
                index: this.entries.indexOf(entry),
                position: { x: position.x, y: position.y, z: position.z },
            });
        });

        return entry;
    }

    private disposeEntry(entry: PointEntry | undefined): void {
        if (!entry) {
            return;
        }
        entry.label.removeFromParent();
        entry.labelElement.remove();
        entry.marker.removeFromParent();
        entry.marker.geometry.dispose();
        entry.material.dispose();
        entry.gizmo.dispose();
    }

    private applyVisibility(): void {
        this.entries.forEach((entry, i) => {
            const focused = this.focus === null || this.focus === i;
            entry.gizmo.setEnabled(this.enabled && focused);
            entry.marker.visible = this.enabled;
            entry.material.color.setHex(
                this.focus === i ? FOCUSED_MARKER_COLOR : MARKER_COLOR
            );
            entry.material.opacity = focused ? 1 : 0.35;
            entry.labelElement.style.opacity = focused ? "1" : "0.35";
        });
        this.line.visible = this.enabled && this.entries.length > 1;
        this.preview.visible = this.enabled && this.previewIndex !== null;
    }

    private positionPreview(): void {
        if (this.previewIndex === null) {
            this.preview.visible = false;
            return;
        }
        const points = this.entries.map((entry) => entry.marker.position);
        const point = pathInsertionPoint(points, this.previewIndex);
        this.preview.position.set(point.x, point.y, point.z);
        this.preview.scale.setScalar(this.entries[0]?.marker.scale.x ?? 0.05);
        this.preview.visible = this.enabled;
    }

    /**
     * Markers are sized off the path's own extent so they stay legible whether
     * the path spans centimeters or hundreds of meters.
     */
    private resizeMarkers(): void {
        if (this.entries.length === 0) {
            return;
        }

        const min = new Vector3(Infinity, Infinity, Infinity);
        const max = new Vector3(-Infinity, -Infinity, -Infinity);
        for (const entry of this.entries) {
            min.min(entry.marker.position);
            max.max(entry.marker.position);
        }

        const diagonal = max.sub(min).length();
        const radius = diagonal > 0 ? Math.max(diagonal * 0.02, 0.01) : 0.05;
        for (const entry of this.entries) {
            entry.marker.scale.setScalar(radius);
        }
    }

    private redraw(): void {
        const count = this.entries.length;
        if (count < 2) {
            this.line.visible = false;
            return;
        }

        const vertices = this.closed ? count + 1 : count;
        let positions = this.lineGeometry.getAttribute("position") as BufferAttribute;
        if (!positions || positions.count !== vertices) {
            positions = new BufferAttribute(new Float32Array(vertices * 3), 3);
            this.lineGeometry.setAttribute("position", positions);
        }

        this.entries.forEach((entry, i) => {
            const p = entry.marker.position;
            positions.setXYZ(i, p.x, p.y, p.z);
        });
        if (this.closed) {
            const p = this.entries[0].marker.position;
            positions.setXYZ(count, p.x, p.y, p.z);
        }

        positions.needsUpdate = true;
        this.lineGeometry.computeBoundingSphere();
        this.line.visible = this.enabled;
    }
}
