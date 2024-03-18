import * as THREE from 'three';
import { TransformControls } from 'three/addons/controls/TransformControls.js';


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
                    data: newData
                });
            }
        });


        app.ViewerScene.add(this.mesh);

        const curVal = parameterData.currentValue;
        this.mesh.position.set(curVal.x, curVal.y, curVal.z);

        app.Scene.add(control)
        control.attach(this.mesh);

        this.lightNode = LiteGraph.createNode("math3d/xyz-to-vec3");
        this.lightNode.title = parameterData.name;
        app.LightGraph.add(this.lightNode);

        this.lightNode.onPropertyChanged = (property, value) => {
            const newData = {
                x: this.mesh.position.x,
                y: this.mesh.position.y,
                z: this.mesh.position.z,
            }
            newData[property] = value;
            nodeManager.nodeParameterChanged({
                id: id,
                data: newData
            });
        }
    }

    update(parameterData) {
        const curVal = parameterData.currentValue;
        this.mesh.position.set(curVal.x, curVal.y, curVal.z)
    }
}     