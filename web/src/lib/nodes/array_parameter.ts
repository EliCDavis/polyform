import { FlowNode, GlobalWidgetFactory, RowWidget, type ValueWidget, type Widget } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { PointPathGizmo, pathInsertionPoint } from '../gizmo/point_path';
import { TRSListGizmo } from '../gizmo/trs_list';
import { eulerToQuat, identityQuat, quatToEuler } from '../rotation';
import { InsertRuleWidget, RemoveWidget } from './list_widgets';
import { lockIfReadOnly } from './read_only';
import type { NodeParameterBase, TRS, Vec2, Vec3 } from '../../types/parameter';

// A viewport handle for the whole list, shown while the node is selected.
export interface ListGizmo<T> {
    setItems(items: Array<T>): void;
    setEnabled(enabled: boolean): void;
    dispose(): void;
}

export type ListGizmoFactory<T> = (app: ThreeApp, items: Array<T>, onChange: (index: number, item: T) => void) => ListGizmo<T>;

export interface ItemComponent<T> {
    group?: string;
    label: string;
    get: (item: T) => number;
    set: (item: T, value: number) => T;
}

// How one item of the list is shown and edited on the node. Vectors get one
// number widget per component; scalars get a single number or string widget.
export interface ArrayItemKind<T> {
    // What an insert at `index` puts there, given what is already there.
    blank: (items: Array<T>, index: number) => T;
    components: ReadonlyArray<ItemComponent<T>> | null;
    widget: "number" | "string" | "color";
    parse?: (n: number) => number;
    gizmo?: ListGizmoFactory<T>;
}

function fields<T extends object>(...keys: Array<keyof T & string>): ItemComponent<T>[] {
    return keys.map((k) => ({
        label: k.toUpperCase(),
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

const before = <T,>(fallback: T) => (items: Array<T>, index: number) => items[index - 1] ?? fallback;

export const floatItem: ArrayItemKind<number> = { blank: before(0), components: null, widget: "number" };
export const intItem: ArrayItemKind<number> = { blank: before(0), components: null, widget: "number", parse: Math.round };
export const stringItem: ArrayItemKind<string> = { blank: () => "", components: null, widget: "string" };
export const colorItem: ArrayItemKind<string> = { blank: before("#ffffff"), components: null, widget: "color" };
export const vec2Item: ArrayItemKind<Vec2> = {
    blank: (items, index) => dropPlanar(pathInsertionPoint(items.map(liftPlanar), index)),
    components: fields<Vec2>("x", "y"),
    widget: "number",
    gizmo: pointPath(true, same, liftPlanar, dropPlanar),
};
export const vec2IntItem: ArrayItemKind<Vec2> = {
    ...vec2Item,
    parse: Math.round,
    gizmo: pointPath(true, Math.round, liftPlanar, dropPlanar),
};
export const vec3Item: ArrayItemKind<Vec3> = {
    blank: pathInsertionPoint,
    components: fields<Vec3>("x", "y", "z"),
    widget: "number",
    gizmo: pointPath(false, same, keep3, keep3),
};
export const vec3IntItem: ArrayItemKind<Vec3> = {
    ...vec3Item,
    parse: Math.round,
    gizmo: pointPath(false, Math.round, keep3, keep3),
};

// Rotation is edited as Euler degrees and folded into the quaternion on set.
function trsComponents(): ItemComponent<TRS>[] {
    const axes = ["x", "y", "z"] as const;
    const vec = (group: string, key: "position" | "scale"): ItemComponent<TRS>[] =>
        axes.map((axis) => ({
            group,
            label: axis.toUpperCase(),
            get: (item) => item[key][axis],
            set: (item, value) => ({ ...item, [key]: { ...item[key], [axis]: value } }),
        }));
    const rotation: ItemComponent<TRS>[] = axes.map((axis) => ({
        group: "rotation",
        label: axis.toUpperCase(),
        get: (item) => Number(quatToEuler(item.rotation)[axis].toFixed(4)),
        set: (item, value) => ({ ...item, rotation: eulerToQuat({ ...quatToEuler(item.rotation), [axis]: value }) }),
    }));
    return [...vec("position", "position"), ...rotation, ...vec("scale", "scale")];
}

const identityTRS: TRS = { position: { x: 0, y: 0, z: 0 }, rotation: identityQuat, scale: { x: 1, y: 1, z: 1 } };

export const trsItem: ArrayItemKind<TRS> = {
    blank: (items, index) => ({
        ...structuredClone(items[index - 1] ?? identityTRS),
        position: pathInsertionPoint(items.map((it) => it.position), index),
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

    gizmo: ListGizmo<T> | null;

    private slots: Array<(item: T) => void> = [];

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: NodeParameterBase<string, Array<T>>, kind: ArrayItemKind<T>, app?: ThreeApp) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.kind = kind;
        this.items = parameterData.currentValue ?? [];
        this.widgets = [];
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

        this.buildWidgets();
    }

    private setItems(items: Array<T>): void {
        this.items = items;
        if (items.length === this.slots.length) {
            items.forEach((item, index) => this.slots[index](item));
        } else {
            this.buildWidgets();
        }
        this.gizmo?.setItems(items);
    }

    private buildWidgets(): void {
        for (const w of this.widgets) {
            this.flowNode.removeWidget(w);
        }
        this.widgets = [];
        this.slots = [];

        this.items.forEach((item, index) => {
            this.add(new InsertRuleWidget(() => this.insert(index)));
            this.add(new RowWidget([
                GlobalWidgetFactory.create(this.flowNode, "text", { value: String(index) }),
                new RemoveWidget(() => this.commit(this.items.filter((_, i) => i !== index))),
            ], { grow: [1, 0] }));
            this.slots.push(this.buildItem(item, index));
        });

        this.add(new InsertRuleWidget(() => this.insert(this.items.length)));
    }

    private buildItem(item: T, index: number): (item: T) => void {
        const replace = (next: T) => this.commit(this.items.map((it, i) => (i === index ? next : it)));

        const components = this.kind.components;
        if (components === null) {
            const widget = this.scalarWidget(item, (v) => replace(v as T));
            this.add(widget);
            return (next) => widget.setValue(next);
        }

        const parse = this.kind.parse ?? ((n: number) => n);
        const shown: Array<{ component: ItemComponent<T>; widget: ValueWidget<number> }> = [];
        const groups = new Map<string | undefined, Array<ItemComponent<T>>>();
        for (const c of components) {
            groups.set(c.group, [...(groups.get(c.group) ?? []), c]);
        }
        for (const [group, members] of groups) {
            if (group) {
                this.add(GlobalWidgetFactory.create(this.flowNode, "text", { value: group }));
            }
            const row = members.map((c) => {
                const widget = GlobalWidgetFactory.create(this.flowNode, "number", {
                    value: c.get(item),
                    label: c.label,
                    width: 72,
                    integer: this.kind.parse !== undefined,
                    callback: (v) => replace(c.set(this.items[index], parse(v))),
                });
                shown.push({ component: c, widget });
                return widget;
            });
            this.add(new RowWidget(row));
        }
        return (next) => shown.forEach(({ component, widget }) => widget.setValue(component.get(next)));
    }

    private insert(index: number): void {
        const next = [...this.items];
        next.splice(index, 0, this.kind.blank(this.items, index));
        this.commit(next);
    }

    private add(widget: Widget): void {
        this.widgets.push(widget);
        this.flowNode.addWidget(lockIfReadOnly(widget));
    }

    private scalarWidget(value: unknown, onChange: (v: unknown) => void): ValueWidget<any> {
        if (this.kind.widget === "color") {
            return GlobalWidgetFactory.create(this.flowNode, "color", { value: String(value), callback: onChange });
        }
        if (this.kind.widget === "string") {
            return GlobalWidgetFactory.create(this.flowNode, "string", { value: String(value), callback: onChange });
        }
        const parse = this.kind.parse ?? ((n: number) => n);
        return GlobalWidgetFactory.create(this.flowNode, "number", {
            value: Number(value),
            integer: this.kind.parse !== undefined,
            callback: (v) => onChange(parse(v)),
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
