import { FlowNode, GlobalWidgetFactory } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { PointPathGizmo, pathInsertionPoint } from '../gizmo/point_path';
import type { Vector3ArrayNodeParameter, Vec3 } from '../../types/parameter';

export class Vector3ArrayParameterNodeController {

    flowNode: FlowNode;

    nodeManager: NodeManager;

    id: string;

    gizmo: PointPathGizmo;

    points: Array<Vec3>;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: Vector3ArrayNodeParameter, app: ThreeApp) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.points = parameterData.currentValue ?? [];

        this.gizmo = new PointPathGizmo({
            camera: app.Camera,
            domElement: app.Renderer.domElement,
            orbitControls: app.OrbitControls,
            parent: app.ViewerScene,
            scene: app.Scene,
            points: this.points,
        });
        this.gizmo.setEnabled(false);

        this.gizmo.changes$().subscribe(({ index, position }) => {
            this.commit(this.points.map((p, i) => (i === index ? position : p)));
        });

        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        const addPointButton = GlobalWidgetFactory.create(this.flowNode, "button", {
            text: "Add Point",
            callback: () => this.commit([...this.points, pathInsertionPoint(this.points, this.points.length)]),
        });
        this.flowNode.addWidget(addPointButton);

        this.flowNode.addSelectListener(() => {
            this.gizmo.setEnabled(true);
        });

        this.flowNode.addUnselectListener(() => {
            this.gizmo.setEnabled(false);
        });
    }

    private commit(points: Array<Vec3>): void {
        this.points = points;
        this.gizmo.setPoints(points);
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: points,
            binary: false
        });
    }

    update(parameterData: Vector3ArrayNodeParameter): void {
        this.points = parameterData.currentValue ?? [];
        this.gizmo.setPoints(this.points);
    }
}
