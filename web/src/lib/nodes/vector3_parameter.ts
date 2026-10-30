import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { TransformGizmo } from '../gizmo/transform';
import type { Vector3NodeParameter } from '../../types/parameter';

export class Vector3ParameterNodeController {

    id: string;

    flowNode: FlowNode;

    nodeManager: NodeManager;

    gizmo: TransformGizmo;

    constructor(
        flowNode: FlowNode,
        nodeManager: NodeManager,
        id: string,
        parameterData: Vector3NodeParameter,
        app: ThreeApp
    ) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.flowNode = flowNode;

        const curVal = parameterData.currentValue;
        this.gizmo = new TransformGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            initialPosition: {
                x: curVal.x,
                y: curVal.y,
                z: curVal.z
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

        this.flowNode.addPropertyEditListener("x", this.propertyChange.bind(this));
        this.flowNode.addPropertyEditListener("y", this.propertyChange.bind(this));
        this.flowNode.addPropertyEditListener("z", this.propertyChange.bind(this));
    }

    propertyChange() {
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

    update(parameterData: Vector3NodeParameter) {
        const curVal = parameterData.currentValue;
        this.gizmo.setPosition(curVal.x, curVal.y, curVal.z)
        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);
        this.flowNode.setProperty("z", curVal.z);
    }
}     