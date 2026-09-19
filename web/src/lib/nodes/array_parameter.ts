import { FlowNode, GlobalWidgetFactory } from '@elicdavis/node-flow';
import type { Widget } from './widget';
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { ColorWidgets } from './color_widget';
import { PointPathGizmo, pathInsertionPoint } from '../gizmo/point_path';
import { TRSListGizmo } from '../gizmo/trs_list';
import { eulerToQuat, identityQuat, quatToEuler } from '../rotation';
import type { NodeParameterBase, TRS, Vec2, Vec3 } from '../../types/parameter';

// A viewport handle for the whole list, shown while the node is selected.
export interface ListGizmo<T> {
    setItems(items: Array<T>): void;
    setEnabled(enabled: boolean): void;
    dispose(): void;
}

export type ListGizmoFactory<T> = (app: ThreeApp, items: Array<T>, onChange: (index: number, item: T) => void) => ListGizmo<T>;

export interface ItemComponent<T> {
    label?: string;
    get: (item: T) => number;
    set: (item: T, value: number) => T;
}

// How one item of the list is shown and edited on the node. Vectors get one
// number widget per component; scalars get a single number or string widget.
export interface ArrayItemKind<T> {
    // What Add appends, given what is already there.
    blank: (items: Array<T>) => T;
    components: ReadonlyArray<ItemComponent<T>> | null;
    widget: "number" | "string" | "color";
    parse?: (n: number) => number;
    gizmo?: ListGizmoFactory<T>;
}

function fields<T extends object>(...keys: Array<keyof T & string>): ItemComponent<T>[] {
    return keys.map((k) => ({
        get: (item) => item[k] as unknown as number,
        set: (item, value) => ({ ...item, [k]: value }),
    }));
}

function pointPath<T>(
    planar: boolean,
    snap: (n: number) => number,
    lift: (item: T) => Vec3,
    drop: (p: Vec3) => T,
): ListGizmoFactory<T> {
    return (app, items, onChange) => {
        const gizmo = new PointPathGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            points: items.map(lift),
            planar,
            closed: planar,
        });
        gizmo.changes$().subscribe(({ index, position }) => {
            onChange(index, drop({ x: snap(position.x), y: snap(position.y), z: snap(position.z) }));
        });
        return {
            setItems: (next) => gizmo.setPoints(next.map(lift)),
            setEnabled: (enabled) => gizmo.setEnabled(enabled),
            dispose: () => gizmo.dispose(),
        };
    };
}

const same = (n: number) => n;
const liftPlanar = (p: Vec2): Vec3 => ({ x: p.x, y: 0, z: p.y });
const dropPlanar = (p: Vec3): Vec2 => ({ x: p.x, y: p.z });
const keep3 = (p: Vec3): Vec3 => p;

const last = <T,>(fallback: T) => (items: Array<T>) => items[items.length - 1] ?? fallback;
const nextPoint = (items: Array<Vec3>) => pathInsertionPoint(items, items.length);

export const floatItem: ArrayItemKind<number> = { blank: last(0), components: null, widget: "number" };
export const intItem: ArrayItemKind<number> = { blank: last(0), components: null, widget: "number", parse: Math.round };
export const stringItem: ArrayItemKind<string> = { blank: () => "", components: null, widget: "string" };
export const colorItem: ArrayItemKind<string> = { blank: last("#ffffff"), components: null, widget: "color" };
export const vec2Item: ArrayItemKind<Vec2> = {
    blank: (items) => dropPlanar(nextPoint(items.map(liftPlanar))),
    components: fields<Vec2>("x", "y"),
    widget: "number",
    gizmo: pointPath(true, same, liftPlanar, dropPlanar),
};
export const vec2IntItem: ArrayItemKind<Vec2> = {
    ...vec2Item,
    parse: Math.round,
    gizmo: pointPath(true, Math.round, liftPlanar, dropPlanar),
};
export const vec3IntItem: ArrayItemKind<Vec3> = {
    blank: nextPoint,
    components: fields<Vec3>("x", "y", "z"),
    widget: "number",
    parse: Math.round,
    gizmo: pointPath(false, Math.round, keep3, keep3),
};

// Rotation is edited as Euler degrees and folded into the quaternion on set.
function trsComponents(): ItemComponent<TRS>[] {
    const axes = ["x", "y", "z"] as const;
    const vec = (label: string, key: "position" | "scale"): ItemComponent<TRS>[] =>
        axes.map((axis, i) => ({
            label: i === 0 ? label : undefined,
            get: (item) => item[key][axis],
            set: (item, value) => ({ ...item, [key]: { ...item[key], [axis]: value } }),
        }));
    const rotation: ItemComponent<TRS>[] = axes.map((axis, i) => ({
        label: i === 0 ? "rotation" : undefined,
        get: (item) => Number(quatToEuler(item.rotation)[axis].toFixed(4)),
        set: (item, value) => ({ ...item, rotation: eulerToQuat({ ...quatToEuler(item.rotation), [axis]: value }) }),
    }));
    return [...vec("position", "position"), ...rotation, ...vec("scale", "scale")];
}

const identityTRS: TRS = { position: { x: 0, y: 0, z: 0 }, rotation: identityQuat, scale: { x: 1, y: 1, z: 1 } };

export const trsItem: ArrayItemKind<TRS> = {
    blank: (items) => ({
        ...structuredClone(items[items.length - 1] ?? identityTRS),
        position: nextPoint(items.map((it) => it.position)),
    }),
    components: trsComponents(),
    widget: "number",
    gizmo: (app, items, onChange) => {
        const gizmo = new TRSListGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            items,
        });
        gizmo.changes$().subscribe(({ index, value }) => onChange(index, value));
        return gizmo;
    },
};

export class ArrayParameterNodeController<T> {

    flowNode: FlowNode;

    nodeManager: NodeManager;

    id: string;

    kind: ArrayItemKind<T>;

    items: Array<T>;

    widgets: Array<Widget>;

    colors: ColorWidgets;

    gizmo: ListGizmo<T> | null;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: NodeParameterBase<string, Array<T>>, kind: ArrayItemKind<T>, app?: ThreeApp) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.kind = kind;
        this.items = parameterData.currentValue ?? [];
        this.widgets = [];
        this.colors = new ColorWidgets(flowNode);
        this.gizmo = null;

        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        if (kind.gizmo && app) {
            this.gizmo = kind.gizmo(app, this.items, (index, item) => {
                this.commit(this.items.map((it, i) => (i === index ? item : it)));
            });
            this.gizmo.setEnabled(false);
            this.flowNode.addSelectListener(() => this.gizmo?.setEnabled(true));
            this.flowNode.addUnselectListener(() => this.gizmo?.setEnabled(false));
        }

        this.rebuildWidgets();
    }

    private setItems(items: Array<T>): void {
        this.items = items;
        this.rebuildWidgets();
        this.gizmo?.setItems(items);
    }

    private rebuildWidgets(): void {
        for (const w of this.widgets) {
            this.flowNode.removeWidget(w);
        }
        this.widgets = [];

        this.items.forEach((item, index) => {
            this.add(GlobalWidgetFactory.create(this.flowNode, "text", { value: String(index) }));

            const components = this.kind.components;
            if (components === null) {
                this.add(this.scalarWidget(item, (v) => v as T, index));
                return;
            }
            for (const c of components) {
                if (c.label) {
                    this.add(GlobalWidgetFactory.create(this.flowNode, "text", { value: c.label }));
                }
                this.add(this.scalarWidget(c.get(item), (v) => c.set(item, v as number), index));
            }
        });

        this.add(GlobalWidgetFactory.create(this.flowNode, "button", {
            text: "Add",
            callback: () => this.commit([...this.items, this.kind.blank(this.items)]),
        }));
        this.add(GlobalWidgetFactory.create(this.flowNode, "button", {
            text: "Remove Last",
            callback: () => this.commit(this.items.slice(0, -1)),
        }));
    }

    private add(widget: Widget): void {
        this.widgets.push(widget);
        this.flowNode.addWidget(widget);
    }

    private scalarWidget(value: unknown, patched: (v: unknown) => T, index: number): Widget {
        if (this.kind.widget === "color") {
            return this.colors.create(`item-${index}`, String(value), (v) =>
                this.commit(this.items.map((it, i) => (i === index ? patched(v) : it))));
        }
        const parse = this.kind.parse ?? ((n: number) => n);
        return GlobalWidgetFactory.create(this.flowNode, this.kind.widget, {
            value,
            callback: (v: unknown) => {
                const next = this.kind.widget === "number" ? parse(Number(v)) : v;
                this.commit(this.items.map((it, i) => (i === index ? patched(next) : it)));
            },
        });
    }

    private commit(items: Array<T>): void {
        this.setItems(items);
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: items,
            binary: false
        });
    }

    update(parameterData: NodeParameterBase<string, unknown>): void {
        this.setItems((parameterData.currentValue as Array<T> | undefined) ?? []);
    }
}
