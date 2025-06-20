import { BehaviorSubject, combineLatestWith, map, mergeMap, skip, Subject } from "rxjs";
import { Variable } from "../schema";
import { inputContainerStyle, LabledField, setVariableValue } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig } from "../element";
import { VariableElement } from "./variable";

export class AABBVariableElement extends VariableElement {

    valuecenterx: Subject<string>;
    valuecentery: Subject<string>;
    valuecenterz: Subject<string>;
    valueextentsx: Subject<string>;
    valueextentsy: Subject<string>;
    valueextentsz: Subject<string>;

    constructor(
        key: string,
        variable: Variable,
        schemaManager: SchemaManager,
        nodeManager: NodeManager,
    ) {
        super(key, variable, schemaManager, nodeManager);
    }

    buildVariable(): ElementConfig {
        const centerx = new BehaviorSubject<string>(`${this.variable.value.center.x}`);
        const centery = new BehaviorSubject<string>(`${this.variable.value.center.y}`);
        const centerz = new BehaviorSubject<string>(`${this.variable.value.center.z}`);
        const extentsx = new BehaviorSubject<string>(`${this.variable.value.extents.x}`);
        const extentsy = new BehaviorSubject<string>(`${this.variable.value.extents.y}`);
        const extentsz = new BehaviorSubject<string>(`${this.variable.value.extents.z}`);

        this.valuecenterx = new Subject<string>();
        this.valuecentery = new Subject<string>();
        this.valuecenterz = new Subject<string>();
        this.valueextentsx = new Subject<string>();
        this.valueextentsy = new Subject<string>();
        this.valueextentsz = new Subject<string>();

        centerx.pipe(
            map(parseFloat),
            combineLatestWith(
                centery.pipe(map(parseFloat)),
                centerz.pipe(map(parseFloat)),
                extentsx.pipe(map(parseFloat)),
                extentsy.pipe(map(parseFloat)),
                extentsz.pipe(map(parseFloat))
            ),
            skip(1), // Ignore the first change, as it's just the initial value
            mergeMap((val) => setVariableValue(this.key, {
                center: { x: val[0], y: val[1], z: val[2] },
                extents: { x: val[3], y: val[4], z: val[5] },
            }))
        ).subscribe((resp: Response) => {
            console.log(resp);
        })

        return {
            style: inputContainerStyle,
            children: [
                { text: "center" },

                LabledField("X:", this.input(centerx, `${this.variable.value.center.x}`, this.valuecenterx)),
                LabledField("Y:", this.input(centery, `${this.variable.value.center.y}`, this.valuecentery)),
                LabledField("Z:", this.input(centerz, `${this.variable.value.center.z}`, this.valuecenterz)),

                { text: "extents" },
                LabledField("X:", this.input(extentsx, `${this.variable.value.extents.x}`, this.valueextentsx)),
                LabledField("Y:", this.input(extentsy, `${this.variable.value.extents.y}`, this.valueextentsy)),
                LabledField("Z:", this.input(extentsz, `${this.variable.value.extents.z}`, this.valueextentsz)),
            ]
        };
    }

    input(change: Subject<string>, value: string, value$: Subject<string>): ElementConfig {
        return { tag: "input", change$: change, type: "number", classList: ['variable-number-input'], value: value, value$: value$ }
    }

    onDestroy(): void {
        this.valuecenterx.complete();
        this.valuecentery.complete();
        this.valuecenterz.complete();
        this.valueextentsx.complete();
        this.valueextentsy.complete();
        this.valueextentsz.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.valuecenterx.next(`${this.variable.value.center.x}`);
        this.valuecentery.next(`${this.variable.value.center.y}`);
        this.valuecenterz.next(`${this.variable.value.center.z}`);
        this.valueextentsx.next(`${this.variable.value.extents.x}`);
        this.valueextentsy.next(`${this.variable.value.extents.y}`);
        this.valueextentsz.next(`${this.variable.value.extents.z}`);
    }
}