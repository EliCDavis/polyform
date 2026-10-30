import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager.js';
import { ThreeApp } from '../three_app.js';
import { AABB, BoxGizmo } from '../gizmo/box.js';
import type { AABBNodeParameter } from '../../types/parameter';

export class AABBParameterNodeController {

    id: string;

    lightNode: FlowNode;

    nodeManager: NodeManager;

    gizmo: BoxGizmo;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: AABBNodeParameter, app: ThreeApp) {
        this.lightNode = flowNode;
        this.lightNode.setTitle(parameterData.name);
        this.nodeManager = nodeManager;
        this.id = id;

        const curVal: AABB = parameterData.currentValue;
        this.gizmo = new BoxGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            initial: curVal
        });

        this.gizmo.aabb$().subscribe((newAABB) => {
            this.setParameter(newAABB);
        });

        this.lightNode.setProperty("min-x", curVal.center.x - curVal.extents.x);
        this.lightNode.setProperty("min-y", curVal.center.y - curVal.extents.y);
        this.lightNode.setProperty("min-z", curVal.center.z - curVal.extents.z);
        this.lightNode.setProperty("max-x", curVal.center.x + curVal.extents.x);
        this.lightNode.setProperty("max-y", curVal.center.y + curVal.extents.y);
        this.lightNode.setProperty("max-z", curVal.center.z + curVal.extents.z);

        for (const property of ["min-x", "min-y", "min-z", "max-x", "max-y", "max-z"]) {
            this.lightNode.addPropertyEditListener(property, this.propertyChange.bind(this));
        }

        this.gizmo.setEnabled(false);

        this.lightNode.addSelectListener(() => {
            this.gizmo.setEnabled(true);
        });

        this.lightNode.addUnselectListener(() => {
            this.gizmo.setEnabled(false);
        });
    }

    private setParameter(newAABB: AABB): void {
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: newAABB,
            binary: false
        });
    }

    propertyChange() {
        this.gizmo.setRight(this.lightNode.getProperty("max-x"));
        this.gizmo.setLeft(this.lightNode.getProperty("min-x"));
        this.gizmo.setUp(this.lightNode.getProperty("max-y"));
        this.gizmo.setDown(this.lightNode.getProperty("min-y"));
        this.gizmo.setForward(this.lightNode.getProperty("max-z"));
        this.gizmo.setBackwards(this.lightNode.getProperty("min-z"));
    }

    update(parameterData: AABBNodeParameter) {
        const curVal: AABB = parameterData.currentValue;

        this.gizmo.set(curVal);
        this.lightNode.setProperty("min-x", curVal.center.x - curVal.extents.x);
        this.lightNode.setProperty("min-y", curVal.center.y - curVal.extents.y);
        this.lightNode.setProperty("min-z", curVal.center.z - curVal.extents.z);
        this.lightNode.setProperty("max-x", curVal.center.x + curVal.extents.x);
        this.lightNode.setProperty("max-y", curVal.center.y + curVal.extents.y);
        this.lightNode.setProperty("max-z", curVal.center.z + curVal.extents.z);
    }

}     