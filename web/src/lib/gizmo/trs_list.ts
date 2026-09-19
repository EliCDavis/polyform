import { Observable, Subject } from "rxjs";
import {
    AxesHelper,
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
import { TRSGizmo, TRSGizmoMode, TRSValue } from "./trs";
import { pathInsertionPoint } from "./point_path";

const MARKER_COLOR = 0x2fd6d6;
const FOCUSED_MARKER_COLOR = 0xffc95c;

export interface TRSListGizmoConfig {
    camera: PerspectiveCamera;
    parent: Group;
    domElement: HTMLElement;
    scene: Scene;
    orbitControls: OrbitControls;
    items?: Array<TRSValue>;
    mode?: TRSGizmoMode;
}

export interface TRSListChange {
    index: number;
    value: TRSValue;
}

interface Entry {
    gizmo: TRSGizmo;
    axes: AxesHelper;
    marker: Mesh;
    material: MeshBasicMaterial;
    label: CSS2DObject;
    labelElement: HTMLDivElement;
}

export class TRSListGizmo {

    private config: TRSListGizmoConfig;

    private entries: Array<Entry>;

    private group: Group;

    private line: Line;

    private lineGeometry: BufferGeometry;

    private preview: Mesh;

    private previewMaterial: MeshBasicMaterial;

    private previewIndex: number | null;

    private markerRadius: number;

    private change$: Subject<TRSListChange>;

    private focus: number | null;

    private enabled: boolean;

    private mode: TRSGizmoMode;

    constructor(config: TRSListGizmoConfig) {
        this.config = config;
        this.entries = [];
        this.change$ = new Subject<TRSListChange>();
        this.focus = null;
        this.enabled = true;
        this.mode = config.mode ?? "translate";

        this.group = new Group();
        config.parent.add(this.group);

        this.lineGeometry = new BufferGeometry();
        this.lineGeometry.setAttribute("position", new BufferAttribute(new Float32Array(0), 3));
        this.line = new Line(
            this.lineGeometry,
            new LineBasicMaterial({ color: 0x2fd6d6, toneMapped: false, depthTest: false })
        );
        this.line.renderOrder = 1;
        this.line.frustumCulled = false;
        this.group.add(this.line);

        this.previewIndex = null;
        this.markerRadius = 0.05;
        this.previewMaterial = new MeshBasicMaterial({
            color: FOCUSED_MARKER_COLOR,
            toneMapped: false,
            depthTest: false,
            transparent: true,
            opacity: 0.55,
        });
        this.preview = new Mesh(new SphereGeometry(1, 16, 12), this.previewMaterial);
        this.preview.renderOrder = 3;
        this.preview.visible = false;
        this.group.add(this.preview);

        this.setItems(config.items ?? []);
    }

    setItems(items: Array<TRSValue>): void {
        while (this.entries.length > items.length) {
            this.disposeEntry(this.entries.pop());
        }
        while (this.entries.length < items.length) {
            this.entries.push(this.newEntry(this.entries.length, items[this.entries.length]));
        }
        items.forEach((item, i) => {
            this.entries[i].gizmo.setValue(item);
            this.place(this.entries[i], item);
        });
        this.resizeAxes();
        this.redraw();
        this.applyVisibility();
        this.positionPreview();
    }

    setFocus(index: number | null): void {
        this.focus = index;
        this.applyVisibility();
    }

    // Ghosts where an item inserted at index would land.
    setInsertPreview(index: number | null): void {
        this.previewIndex = index;
        this.positionPreview();
    }

    setMode(mode: TRSGizmoMode): void {
        this.mode = mode;
        for (const entry of this.entries) {
            entry.gizmo.setMode(mode);
        }
    }

    setEnabled(enabled: boolean): void {
        this.enabled = enabled;
        this.applyVisibility();
    }

    changes$(): Observable<TRSListChange> {
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

    private newEntry(index: number, item: TRSValue): Entry {
        const gizmo = new TRSGizmo({
            camera: this.config.camera,
            domElement: this.config.domElement,
            orbitControls: this.config.orbitControls,
            parent: this.config.parent,
            scene: this.config.scene,
            initial: item,
            mode: this.mode,
        });

        const axes = new AxesHelper(1);
        this.group.add(axes);

        const material = new MeshBasicMaterial({
            color: MARKER_COLOR,
            toneMapped: false,
            depthTest: false,
            transparent: true,
        });
        const marker = new Mesh(new SphereGeometry(1, 16, 12), material);
        marker.renderOrder = 2;
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

        const entry: Entry = { gizmo, axes, marker, material, label, labelElement };
        this.place(entry, item);

        gizmo.changes$().subscribe((value) => {
            this.place(entry, value);
            this.redraw();
            this.change$.next({ index: this.entries.indexOf(entry), value });
        });

        return entry;
    }

    private place(entry: Entry, item: TRSValue): void {
        entry.axes.position.set(item.position.x, item.position.y, item.position.z);
        entry.axes.quaternion.set(item.rotation.x, item.rotation.y, item.rotation.z, item.rotation.w);
        entry.marker.position.copy(entry.axes.position);
    }

    private positionPreview(): void {
        if (this.previewIndex === null) {
            this.preview.visible = false;
            return;
        }
        const points = this.entries.map((entry) => entry.marker.position);
        const point = pathInsertionPoint(points, this.previewIndex);
        this.preview.position.set(point.x, point.y, point.z);
        this.preview.scale.setScalar(this.markerRadius);
        this.preview.visible = this.enabled;
    }

    private disposeEntry(entry: Entry | undefined): void {
        if (!entry) {
            return;
        }
        entry.label.removeFromParent();
        entry.labelElement.remove();
        entry.marker.removeFromParent();
        entry.marker.geometry.dispose();
        entry.material.dispose();
        entry.axes.removeFromParent();
        entry.axes.dispose();
        entry.gizmo.dispose();
    }

    private applyVisibility(): void {
        this.entries.forEach((entry, i) => {
            const focused = this.focus === null || this.focus === i;
            entry.gizmo.setEnabled(this.enabled && focused);
            entry.axes.visible = this.enabled;
            entry.marker.visible = this.enabled;
            entry.material.color.setHex(this.focus === i ? FOCUSED_MARKER_COLOR : MARKER_COLOR);
            entry.material.opacity = focused ? 1 : 0.35;
            entry.labelElement.style.opacity = focused ? "1" : "0.35";
        });
        this.line.visible = this.enabled && this.entries.length > 1;
        this.preview.visible = this.enabled && this.previewIndex !== null;
    }

    private redraw(): void {
        const count = this.entries.length;
        if (count < 2) {
            this.line.visible = false;
            return;
        }

        let positions = this.lineGeometry.getAttribute("position") as BufferAttribute;
        if (!positions || positions.count !== count) {
            positions = new BufferAttribute(new Float32Array(count * 3), 3);
            this.lineGeometry.setAttribute("position", positions);
        }
        this.entries.forEach((entry, i) => {
            const p = entry.axes.position;
            positions.setXYZ(i, p.x, p.y, p.z);
        });
        positions.needsUpdate = true;
        this.lineGeometry.computeBoundingSphere();
        this.line.visible = this.enabled;
    }

    private resizeAxes(): void {
        if (this.entries.length === 0) {
            return;
        }
        const min = new Vector3(Infinity, Infinity, Infinity);
        const max = new Vector3(-Infinity, -Infinity, -Infinity);
        for (const entry of this.entries) {
            min.min(entry.axes.position);
            max.max(entry.axes.position);
        }
        const diagonal = max.sub(min).length();
        const size = diagonal > 0 ? Math.max(diagonal * 0.08, 0.05) : 0.25;
        this.markerRadius = diagonal > 0 ? Math.max(diagonal * 0.02, 0.01) : 0.05;
        for (const entry of this.entries) {
            entry.axes.scale.setScalar(size);
            entry.marker.scale.setScalar(this.markerRadius);
        }
    }
}
