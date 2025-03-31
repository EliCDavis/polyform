import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { Group } from 'three';
import { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { ThreeApp } from '../three_app';

export class Vector3ParameterNodeController {
    
    id: string;

    flowNode: FlowNode;

    nodeManager: NodeManager;

    mesh: Group;
    
    updating: boolean;

    app: ThreeApp;

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

        const control = new TransformControls(app.Camera, app.Renderer.domElement);
        control.setMode('translate');
        control.setSpace("local");

        this.mesh = new Group();

        control.addEventListener('dragging-changed', (event) => {
            app.OrbitControls.enabled = !event.value;

            if (!app.OrbitControls.enabled) {
                return;
            }

            nodeManager.nodeParameterChanged({
                id: id,
                data: {
                    x: this.mesh.position.x,
                    y: this.mesh.position.y,
                    z: this.mesh.position.z,
                },
                binary: false
            });
        });

        app.ViewerScene.add(this.mesh);


        const curVal = parameterData.currentValue;
        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);
        this.flowNode.setProperty("z", curVal.z);
        this.mesh.position.set(curVal.x, curVal.y, curVal.z);

        const helper = control.getHelper();
        app.Scene.add(helper)
        control.attach(this.mesh);

        this.flowNode.setTitle(parameterData.name);

        helper.visible = false;
        // helper.enabled = false;
        control.enabled = false;

        this.flowNode.addSelectListener(() => {
            helper.visible = true;
            // helper.enabled = true;
            control.enabled = true;
        });

        this.flowNode.addUnselectListener(() => {
            helper.visible = false;
            // helper.enabled = false;
            control.enabled = false;
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
        this.mesh.position.set(curVal.x, curVal.y, curVal.z);
        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);
        this.flowNode.setProperty("z", curVal.z);
        this.updating = false;
    }
}     