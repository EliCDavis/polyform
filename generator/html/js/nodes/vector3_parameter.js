import * as THREE from 'three';
import { TransformControls } from 'three/addons/controls/TransformControls.js';

export class Vector3ParameterNodeController {
    constructor(lightNode, nodeManager, id, parameterData, app) {
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

        this.lightNode = lightNode;
        this.lightNode.title = parameterData.name;

        control.visible = false;
        control.enabled = false;

        this.lightNode.onSelected = (obj) => {
            control.visible = true;
            control.enabled = true;
        }
        
        this.lightNode.onDeselected = (obj) => {
            control.visible = false;
            control.enabled = false;
        }

    }

    update(parameterData) {
        const curVal = parameterData.currentValue;
        this.mesh.position.set(curVal.x, curVal.y, curVal.z)
    }
}     