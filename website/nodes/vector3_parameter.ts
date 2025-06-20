import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { Group } from 'three';
import { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { ThreeApp } from '../three_app';
import { TransformGizmo } from '../gizmo.ts/transform';

export class Vector3ParameterNodeController {

    id: string;

    flowNode: FlowNode;

    nodeManager: NodeManager;

    gizmo: TransformGizmo;

    updating: boolean;

    constructor(
        flowNode: FlowNode,
        nodeManager: NodeManager,
        id: string,
        parameterData,
        app: ThreeApp
    ) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.flowNode = flowNode;
        this.updating = false;

        const curVal = parameterData.currentValue;
        this.gizmo = new TransformGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            initialPosition: {
                x: curVal.x,
                y: curVal.x,
                z: curVal.x
            }
        })

        this.gizmo.position$().subscribe(position => {
            nodeManager.nodeParameterChanged({
                id: id,
                data: {
                    x: position.x,
                    y: position.y,
                    z: position.z,
                },
                binary: false
            });
        })

        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);
        this.flowNode.setProperty("z", curVal.z);


        this.flowNode.setTitle(parameterData.name);

        this.flowNode.addSelectListener(() => {
            this.gizmo.setEnabled(true);
        });

        this.flowNode.addUnselectListener(() => {
            this.gizmo.setEnabled(false);
        });

        this.flowNode.addPropertyChangeListener("x", this.propertyChange.bind(this));
        this.flowNode.addPropertyChangeListener("y", this.propertyChange.bind(this));
        this.flowNode.addPropertyChangeListener("z", this.propertyChange.bind(this));
    }

    propertyChange() {
        if (this.updating) {
            return
        }
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: {
                x: this.flowNode.getProperty("x"),
                y: this.flowNode.getProperty("y"),
                z: this.flowNode.getProperty("z")
            },
            binary: false
        });
    }

    update(parameterData) {
        this.updating = true;
        const curVal = parameterData.currentValue;
        this.gizmo.setPosition(curVal.x, curVal.y, curVal.z)
        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);
        this.flowNode.setProperty("z", curVal.z);
        this.updating = false;
    }
}     