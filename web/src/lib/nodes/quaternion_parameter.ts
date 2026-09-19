import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { TRSGizmo } from '../gizmo/trs';
import { eulerToQuat, identityQuat, quatToEuler, type Quat } from '../rotation';
import type { QuaternionNodeParameter } from '../../types/parameter';

const axes = ["x", "y", "z"] as const;

const atOrigin = (rotation: Quat) => ({
    position: { x: 0, y: 0, z: 0 },
    rotation,
    scale: { x: 1, y: 1, z: 1 },
});

export class QuaternionParameterNodeController {

    id: string;

    flowNode: FlowNode;

    nodeManager: NodeManager;

    gizmo: TRSGizmo;

    updating: boolean;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: QuaternionNodeParameter, app: ThreeApp) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.flowNode = flowNode;
        this.updating = false;

        this.gizmo = new TRSGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            initial: atOrigin(parameterData.currentValue ?? identityQuat),
            mode: "rotate",
        });
        this.gizmo.changes$().subscribe((v) => {
            this.setProperties(v.rotation);
            this.nodeManager.nodeParameterChanged({ id: this.id, data: v.rotation, binary: false });
        });

        this.flowNode.setTitle(parameterData.name);
        this.setProperties(parameterData.currentValue ?? identityQuat);

        this.flowNode.addSelectListener(() => this.gizmo.setEnabled(true));
        this.flowNode.addUnselectListener(() => this.gizmo.setEnabled(false));

        for (const axis of axes) {
            this.flowNode.addPropertyChangeListener(axis, this.propertyChange.bind(this));
        }
    }

    private setProperties(rotation: Quat): void {
        this.updating = true;
        const euler = quatToEuler(rotation);
        for (const axis of axes) {
            this.flowNode.setProperty(axis, Number(euler[axis].toFixed(4)));
        }
        this.updating = false;
    }

    propertyChange() {
        if (this.updating) {
            return;
        }
        const rotation = eulerToQuat({
            x: Number(this.flowNode.getProperty("x")),
            y: Number(this.flowNode.getProperty("y")),
            z: Number(this.flowNode.getProperty("z")),
        });
        this.gizmo.setValue(atOrigin(rotation));
        this.nodeManager.nodeParameterChanged({ id: this.id, data: rotation, binary: false });
    }

    update(parameterData: QuaternionNodeParameter) {
        const rotation = parameterData.currentValue ?? identityQuat;
        this.gizmo.setValue(atOrigin(rotation));
        this.setProperties(rotation);
    }
}
