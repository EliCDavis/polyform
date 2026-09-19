import { FlowNode, GlobalWidgetFactory } from '@elicdavis/node-flow';
import type { Widget } from './widget';
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { TRSGizmo, TRSGizmoMode } from '../gizmo/trs';
import { eulerToQuat, identityQuat, quatToEuler } from '../rotation';
import type { TRS, TRSNodeParameter } from '../../types/parameter';

const identityTRS: TRS = {
    position: { x: 0, y: 0, z: 0 },
    rotation: identityQuat,
    scale: { x: 1, y: 1, z: 1 },
};

const axes = ["x", "y", "z"] as const;

export class TRSParameterNodeController {

    id: string;

    flowNode: FlowNode;

    nodeManager: NodeManager;

    gizmo: TRSGizmo;

    updating: boolean;

    modeButtons: Array<Widget>;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: TRSNodeParameter, app: ThreeApp) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.flowNode = flowNode;
        this.updating = false;
        this.modeButtons = [];

        const value = parameterData.currentValue ?? identityTRS;
        this.gizmo = new TRSGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            initial: value,
        });
        this.gizmo.changes$().subscribe((next) => {
            this.setProperties(next);
            this.nodeManager.nodeParameterChanged({ id: this.id, data: next, binary: false });
        });

        this.flowNode.setTitle(parameterData.name);
        this.setProperties(value);

        for (const mode of ["translate", "rotate", "scale"] as const) {
            this.flowNode.addWidget(GlobalWidgetFactory.create(this.flowNode, "button", {
                text: mode[0].toUpperCase() + mode.slice(1),
                callback: () => this.gizmo.setMode(mode as TRSGizmoMode),
            }));
        }

        this.flowNode.addSelectListener(() => this.gizmo.setEnabled(true));
        this.flowNode.addUnselectListener(() => this.gizmo.setEnabled(false));

        for (const prefix of ["p", "r", "s"]) {
            for (const axis of axes) {
                this.flowNode.addPropertyChangeListener(prefix + axis, this.propertyChange.bind(this));
            }
        }
    }

    private setProperties(value: TRS): void {
        this.updating = true;
        const euler = quatToEuler(value.rotation);
        for (const axis of axes) {
            this.flowNode.setProperty("p" + axis, value.position[axis]);
            this.flowNode.setProperty("r" + axis, Number(euler[axis].toFixed(4)));
            this.flowNode.setProperty("s" + axis, value.scale[axis]);
        }
        this.updating = false;
    }

    private value(): TRS {
        const read = (prefix: string) => ({
            x: Number(this.flowNode.getProperty(prefix + "x")),
            y: Number(this.flowNode.getProperty(prefix + "y")),
            z: Number(this.flowNode.getProperty(prefix + "z")),
        });
        return {
            position: read("p"),
            rotation: eulerToQuat(read("r")),
            scale: read("s"),
        };
    }

    propertyChange() {
        if (this.updating) {
            return;
        }
        const next = this.value();
        this.gizmo.setValue(next);
        this.nodeManager.nodeParameterChanged({ id: this.id, data: next, binary: false });
    }

    update(parameterData: TRSNodeParameter) {
        const value = parameterData.currentValue ?? identityTRS;
        this.gizmo.setValue(value);
        this.setProperties(value);
    }
}
