import { InstancedMesh, Mesh, Object3D, Points } from "three";

export interface ModelStats {
    vertices: number;
    triangles: number;
    draws: number;
}

export function countModelStats(root: Object3D | null): ModelStats {
    const stats: ModelStats = { vertices: 0, triangles: 0, draws: 0 };
    if (!root) {
        return stats;
    }

    root.traverse((object) => {
        const geometry = (object as Mesh).geometry;
        if (!geometry || !geometry.getAttribute) {
            return;
        }

        const position = geometry.getAttribute("position");
        if (!position) {
            return;
        }

        // Instanced geometry is stored once but drawn `count` times, so the
        // stored counts have to be multiplied to describe what's on screen.
        const instances = object instanceof InstancedMesh ? object.count : 1;

        if (object instanceof Points) {
            stats.draws += 1;
            stats.vertices += position.count * instances;
            return;
        }

        if (!(object instanceof Mesh)) {
            return;
        }

        // A mesh split across material groups is submitted once per group.
        stats.draws += Array.isArray(object.material) ? object.material.length : 1;
        stats.vertices += position.count * instances;
        const indices = geometry.getIndex();
        stats.triangles += ((indices ? indices.count : position.count) / 3) * instances;
    });

    stats.triangles = Math.floor(stats.triangles);
    return stats;
}
