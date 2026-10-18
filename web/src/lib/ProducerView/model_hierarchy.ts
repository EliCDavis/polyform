import { InstancedMesh, Mesh, Object3D, Points } from "three";
import { countModelStats } from "./model_stats";

export type HierarchyKind = "mesh" | "points" | "group";

export interface HierarchyEntry {
    path: string;
    name: string;
    kind: HierarchyKind;
    instances: number;
    visible: boolean;
    triangles: number;
    vertices: number;
    children: HierarchyEntry[];
}

function kindOf(object: Object3D): HierarchyKind {
    if (object instanceof Points) {
        return "points";
    }
    if (object instanceof Mesh) {
        return "mesh";
    }
    return "group";
}

// Paths survive a regenerated model, so a part hidden before an edit stays hidden after it.
function childSegments(parent: Object3D): string[] {
    const seen = new Map<string, number>();
    return parent.children.map((child) => {
        const name = child.name || child.type;
        const occurrence = seen.get(name) ?? 0;
        seen.set(name, occurrence + 1);
        return `${name}#${occurrence}`;
    });
}

// The loaded scene sits under one wrapper group; listing it adds a level with nothing in it.
export function hierarchyRoots(root: Object3D | null): Object3D | null {
    if (!root) {
        return null;
    }
    let current = root;
    while (current.children.length === 1 && kindOf(current.children[0]) === "group" && !current.children[0].name) {
        current = current.children[0];
    }
    return current;
}

export function walkHierarchy(root: Object3D | null, visit: (path: string, object: Object3D) => void): void {
    const start = hierarchyRoots(root);
    if (!start) {
        return;
    }
    const walk = (parent: Object3D, prefix: string) => {
        const segments = childSegments(parent);
        parent.children.forEach((child, i) => {
            const path = prefix + "/" + segments[i];
            visit(path, child);
            walk(child, path);
        });
    };
    walk(start, "");
}

export function buildHierarchy(root: Object3D | null): HierarchyEntry[] {
    const start = hierarchyRoots(root);
    if (!start) {
        return [];
    }
    const build = (parent: Object3D, prefix: string): HierarchyEntry[] => {
        const segments = childSegments(parent);
        return parent.children.map((child, i) => {
            const path = prefix + "/" + segments[i];
            const stats = countModelStats(child, false);
            return {
                path,
                name: child.name || child.type,
                kind: kindOf(child),
                instances: child instanceof InstancedMesh ? child.count : 1,
                visible: child.visible,
                triangles: stats.triangles,
                vertices: stats.vertices,
                children: build(child, path),
            };
        });
    };
    return build(start, "");
}
