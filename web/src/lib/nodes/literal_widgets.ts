import type { FlowNode, WidgetConfig } from "@elicdavis/node-flow";
import { eulerToQuat, quatToEuler, type Quat } from "../rotation";
import { opaque, shownAlpha, withAlpha } from "../color_alpha";
import { alphaWidgetConfig } from "./color_widgets";
import "./gradient_literal_widget";

export type LiteralKind =
    | "float"
    | "unorm"
    | "snorm"
    | "int"
    | "bool"
    | "string"
    | "color"
    | "enum"
    | "vector2"
    | "vector2int"
    | "vector3"
    | "vector3int"
    | "vector4"
    | "quaternion"
    | "gradient";

const AXES = ["x", "y", "z", "w"];

const kindsByType: { [type: string]: LiteralKind } = {
    "float64": "float",
    "float32": "float",
    "github.com/EliCDavis/polyform/math/norm.UFloat64": "unorm",
    "github.com/EliCDavis/polyform/math/norm.SFloat64": "snorm",
    "int": "int",
    "bool": "bool",
    "string": "string",
    "github.com/EliCDavis/polyform/drawing/coloring.Color": "color",
    "github.com/EliCDavis/vector/vector2.Vector[float64]": "vector2",
    "github.com/EliCDavis/vector/vector2.Vector[int]": "vector2int",
    "github.com/EliCDavis/vector/vector3.Vector[float64]": "vector3",
    "github.com/EliCDavis/vector/vector3.Vector[int]": "vector3int",
    "github.com/EliCDavis/vector/vector4.Vector[float64]": "vector4",
    "github.com/EliCDavis/polyform/math/quaternion.Quaternion": "quaternion",
    "github.com/EliCDavis/polyform/drawing/coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color]": "gradient",
};

export function literalKind(portType: string, enumMembers?: string[]): LiteralKind | undefined {
    if (enumMembers && enumMembers.length > 0) {
        return "enum";
    }
    return kindsByType[portType];
}

function axesOf(kind: LiteralKind): string[] {
    switch (kind) {
        case "vector2":
        case "vector2int":
            return AXES.slice(0, 2);
        case "vector3":
        case "vector3int":
        case "quaternion":
            return AXES.slice(0, 3);
        case "vector4":
            return AXES;
    }
    return [];
}

export function literalProperty(port: string, axis?: string): string {
    return axis === undefined ? `literal:${port}` : `literal:${port}.${axis}`;
}

export function literalProperties(port: string, kind: LiteralKind): string[] {
    if (kind === "color") {
        return [literalProperty(port), literalProperty(port, "a")];
    }
    const axes = axesOf(kind);
    if (axes.length === 0) {
        return [literalProperty(port)];
    }
    return axes.map((axis) => literalProperty(port, axis));
}

function componentsOf(kind: LiteralKind, value: unknown): { [axis: string]: number } {
    if (kind === "quaternion") {
        const q = (value ?? { x: 0, y: 0, z: 0, w: 1 }) as Quat;
        const euler = quatToEuler(q);
        return { x: euler.x, y: euler.y, z: euler.z };
    }
    const v = (value ?? {}) as { [axis: string]: number };
    const out: { [axis: string]: number } = {};
    for (const axis of axesOf(kind)) {
        out[axis] = Number(v[axis] ?? 0);
    }
    return out;
}

export function literalWidgetConfig(
    port: string,
    kind: LiteralKind,
    defaultValue: unknown,
    enumMembers?: string[],
): WidgetConfig {
    const property = literalProperty(port);
    switch (kind) {
        case "float":
            return { type: "number", config: { property, value: Number(defaultValue ?? 0), width: 90 } };
        case "unorm":
        case "snorm":
            return {
                type: "slider",
                config: { property, value: Number(defaultValue ?? 0), min: kind === "snorm" ? -1 : 0, max: 1, width: 90 },
            };
        case "int":
            return { type: "number", config: { property, value: Number(defaultValue ?? 0), integer: true, width: 90 } };
        case "bool":
            return { type: "toggle", config: { property, value: defaultValue === true, width: 60, label: "" } };
        case "string":
            return { type: "string", config: { property, value: String(defaultValue ?? ""), width: 120 } };
        case "color": {
            const hex = String(defaultValue ?? "#000000");
            return {
                type: "row",
                config: {
                    grow: [1, 0],
                    widgets: [
                        { type: "color", config: { property, value: opaque(hex), width: 90 } },
                        { type: "number", config: { ...alphaWidgetConfig, property: literalProperty(port, "a"), value: shownAlpha(hex) } },
                    ],
                },
            };
        }
        case "enum":
            return {
                type: "select",
                config: { property, value: String(defaultValue ?? enumMembers?.[0] ?? ""), options: enumMembers ?? [], width: 120 },
            } as unknown as WidgetConfig;
        case "gradient":
            return { type: "gradient", config: { property, value: defaultValue } } as unknown as WidgetConfig;
    }

    const integer = kind === "vector2int" || kind === "vector3int";
    const components = componentsOf(kind, defaultValue);
    return {
        type: "row",
        config: {
            widgets: axesOf(kind).map((axis): WidgetConfig => ({
                type: "number",
                config: {
                    property: literalProperty(port, axis),
                    value: components[axis],
                    label: axis.toUpperCase(),
                    integer,
                    width: 64,
                },
            })),
        },
    };
}

export function applyLiteral(node: FlowNode, port: string, kind: LiteralKind, value: unknown): void {
    if (kind === "color") {
        node.setProperty(literalProperty(port), opaque(String(value)));
        node.setProperty(literalProperty(port, "a"), shownAlpha(String(value)));
        return;
    }
    const axes = axesOf(kind);
    if (axes.length === 0) {
        node.setProperty(literalProperty(port), value);
        return;
    }
    const components = componentsOf(kind, value);
    for (const axis of axes) {
        node.setProperty(literalProperty(port, axis), components[axis]);
    }
}

export function readLiteral(node: FlowNode, port: string, kind: LiteralKind): unknown {
    if (kind === "color") {
        return withAlpha(String(node.getProperty(literalProperty(port))), Number(node.getProperty(literalProperty(port, "a"))));
    }
    const axes = axesOf(kind);
    if (axes.length === 0) {
        const value = node.getProperty(literalProperty(port));
        switch (kind) {
            case "float":
            case "unorm":
            case "snorm":
                return Number(value);
            case "int":
                return Math.round(Number(value));
            case "bool":
                return value === true;
            case "gradient":
                return value;
            default:
                return String(value ?? "");
        }
    }
    const components: { [axis: string]: number } = {};
    for (const axis of axes) {
        components[axis] = Number(node.getProperty(literalProperty(port, axis)) ?? 0);
    }
    if (kind === "quaternion") {
        return eulerToQuat({ x: components.x, y: components.y, z: components.z });
    }
    return components;
}
