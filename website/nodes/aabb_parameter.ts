import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager.js';
import { ThreeApp } from '../three_app.js';
import { AABB, BoxGizmo } from '../gizmo/box.js';

export class AABBParameterNodeController {

    id: string;

    lightNode: FlowNode;

    nodeManager: NodeManager;

    updating: boolean;

    gizmo: BoxGizmo;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData, app: ThreeApp) {
        this.lightNode = flowNode;
        this.lightNode.setTitle(parameterData.name);
        this.updating = false;
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

        this.lightNode.addPropertyChangeListener("min-x", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("min-y", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("min-z", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("max-x", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("max-y", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("max-z", this.propertyChange.bind(this));

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
        if (this.updating) {
            return
        }

        this.gizmo.setRight(this.lightNode.getProperty("max-x"));
        this.gizmo.setLeft(this.lightNode.getProperty("min-x"));
        this.gizmo.setUp(this.lightNode.getProperty("max-y"));
        this.gizmo.setDown(this.lightNode.getProperty("min-y"));
        this.gizmo.setForward(this.lightNode.getProperty("max-z"));
        this.gizmo.setBackwards(this.lightNode.getProperty("min-z"));
    }

    update(parameterData) {
        this.updating = true;

        const curVal: AABB = parameterData.currentValue;

        this.gizmo.set(curVal);
        this.lightNode.setProperty("min-x", curVal.center.x - curVal.extents.x);
        this.lightNode.setProperty("min-y", curVal.center.y - curVal.extents.y);
        this.lightNode.setProperty("min-z", curVal.center.z - curVal.extents.z);
        this.lightNode.setProperty("max-x", curVal.center.x + curVal.extents.x);
        this.lightNode.setProperty("max-y", curVal.center.y + curVal.extents.y);
        this.lightNode.setProperty("max-z", curVal.center.z + curVal.extents.z);

        this.updating = false;
    }

}     