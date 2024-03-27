import * as THREE from 'three';
import { TransformControls } from 'three/addons/controls/TransformControls.js';

function BuildVector3ParameterNode(app) {
    const node = LiteGraph.createNode("polyform/vector3");
    console.log(node)
    app.LightGraph.add(node);
    return node;
}

export class NodeVector3Parameter {
    constructor(nodeManager, id, parameterData, app) {
        const control = new TransformControls(app.Camera, app.Renderer.domElement);
        control.setMode('translate');
        control.setSpace("local");

        this.mesh = new THREE.Group();

        control.addEventListener('dragging-changed', (event) => {
            app.OrbitControls.enabled = !event.value;

            if (app.OrbitControls.enabled) {
                const newData = {
                    x: this.mesh.position.x,
                    y: this.mesh.position.y,
                    z: this.mesh.position.z,
                }
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newData,
                    binary: false
                });
            }
        });


        app.ViewerScene.add(this.mesh);

        const curVal = parameterData.currentValue;
        this.mesh.position.set(curVal.x, curVal.y, curVal.z);

        app.Scene.add(control)
        control.attach(this.mesh);

        this.lightNode = BuildVector3ParameterNode(app);
        this.lightNode.title = parameterData.name;
        // app.LightGraph.add(this.lightNode);

        control.visible = false;
        control.enabled = false;

        this.lightNode.onSelected = (obj) => {
            console.log("selected", obj);
            console.log(control)
            control.visible = true;
            control.enabled = true;
        }
        
        this.lightNode.onDeselected = (obj) => {
            console.log("de-selected", obj)
            control.visible = false;
            control.enabled = false;
        }

        // this.lightNode.onPropertyChanged = (property, value) => {
        //     const newData = {
        //         x: this.mesh.position.x,
        //         y: this.mesh.position.y,
        //         z: this.mesh.position.z,
        //     }
        //     newData[property] = value;
        //     nodeManager.nodeParameterChanged({
        //         id: id,
        //         data: newData
        //     });
        // }
    }

    update(parameterData) {
        const curVal = parameterData.currentValue;
        this.mesh.position.set(curVal.x, curVal.y, curVal.z)
    }
}     