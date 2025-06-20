import { BehaviorSubject, combineLatestWith, map, mergeMap, skip, Subject } from "rxjs";
import { Variable } from "../schema";
import { inputContainerStyle, LabledField, setVariableValue } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig } from "../element";
import { TransformGizmo } from "../gizmo.ts/transform";
import { ThreeApp } from "../three_app";
import { Toggle } from "../components/toggle";
import { VariableElement } from "./variable";

export class Vector3VariableElement extends VariableElement {

    gizmo: TransformGizmo;

    valueX$: Subject<string>;
    valueY$: Subject<string>;
    valueZ$: Subject<string>;

    constructor(
        key: string,
        variable: Variable,
        schemaManager: SchemaManager,
        nodeManager: NodeManager,
        private app: ThreeApp,
        private mapper: (s: string) => number,
        private step: string
    ) {
        super(key, variable, schemaManager, nodeManager);
    }

    buildVariable(): ElementConfig {
        const changeX$ = new BehaviorSubject<string>(`${this.variable.value.x}`);
        const changeY$ = new BehaviorSubject<string>(`${this.variable.value.y}`);
        const changeZ$ = new BehaviorSubject<string>(`${this.variable.value.z}`);
        this.valueX$ = new Subject<string>();
        this.valueY$ = new Subject<string>();
        this.valueZ$ = new Subject<string>();

        this.addSubscription(changeX$.pipe(
            map(this.mapper),
            combineLatestWith(changeY$.pipe(map(this.mapper)), changeZ$.pipe(map(this.mapper))),
            skip(1),
            mergeMap((val) => setVariableValue(this.key, { x: val[0], y: val[1], z: val[2] }))
        ).subscribe((resp: Response) => {
            console.log(resp);
        }));

        const showGizmo = new BehaviorSubject<boolean>(false);
        let gizmo = new TransformGizmo({
            camera: this.app.Camera,
            domElement: this.app.Renderer.domElement,
            orbitControls: this.app.OrbitControls,
            parent: this.app.ViewerScene,
            scene: this.app.Scene,
            initialPosition: {
                x: this.variable.value.x,
                y: this.variable.value.x,
                z: this.variable.value.x
            }
        });
        this.addSubscription(showGizmo.subscribe((show) => gizmo.setEnabled(show)));
        this.addSubscription(gizmo.position$().subscribe(pos => {
            setVariableValue(this.key, { x: pos.x, y: pos.y, z: pos.z })
        }))

        return {
            style: inputContainerStyle,
            children: [
                LabledField("X:", {
                    tag: "input",
                    type: "number",
                    size: 1,
                    classList: ['variable-number-input'],
                    change$: changeX$,
                    step: this.step,
                    value: `${this.variable.value.x}`,
                    value$: this.valueX$
                }),
                LabledField("Y:", {
                    tag: "input",
                    change$: changeY$,
                    type: "number",
                    size: 1,
                    classList: ['variable-number-input'],
                    value: `${this.variable.value.y}`,
                    step: this.step,
                    value$: this.valueY$
                }),
                LabledField("Z:", {
                    tag: "input",
                    change$: changeZ$,
                    type: "number",
                    size: 1,
                    classList: ['variable-number-input'],
                    value: `${this.variable.value.z}`,
                    step: this.step,
                    value$: this.valueZ$
                }),
                {
                    style: {
                        display: "flex",
                        flexDirection: "row",
                        alignItems: "center",
                        gap: "8px"
                    },
                    children: [
                        {
                            tag: "i",
                            classList: ["fa-solid", "fa-eye"],
                            style: { color: "#196d6d" }
                        },
                        { text: "Gizmo" },
                        { style: { flex: "1" } },
                        Toggle({ initialValue: false, change: showGizmo })
                    ]
                },
            ]
        };
    }

    onDestroy(): void {
        this.gizmo.dispose();
        this.valueX$.complete();
        this.valueY$.complete();
        this.valueZ$.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.valueX$.next(`${this.variable.value.x}`)
        this.valueY$.next(`${this.variable.value.y}`)
        this.valueZ$.next(`${this.variable.value.z}`)
    }
}