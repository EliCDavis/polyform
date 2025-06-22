import { BehaviorSubject, combineLatestWith, map, mergeMap, Observable, skip, Subject } from "rxjs";
import { Variable } from "../schema";
import { inputContainerStyle, LabledField, setVariableValue } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig } from "../element";
import { ThreeApp } from "../three_app";
import { VariableElement } from "./variable";
import { ElementInstance } from "./element_instance";


function bind(obj: any, field: string, mapper: (s: string) => any): Subject<string> {
    const x = new BehaviorSubject<string>(`${obj[field]}`);
    x.subscribe((val: string) => {
        obj[field] = mapper(val);
    })
    return x;
}

interface Vector3 { x: number, y: number, z: number }

// class Vector3Element extends ElementInstance<Vector3> {

//     vector: Vector3;

//     constructor(vector: Vector3) {
//         super();
//         this.vector = vector;
//     }

//     set(data: Vector3): void {
//         throw new Error("Method not implemented.");
//     }

//     onDestroy(): void {
//         throw new Error("Method not implemented.");
//     }

//     build(): ElementConfig {
//         return {
//             children: [
//                 {
//                     style: {
//                         display: "flex",
//                         flexDirection: "column"
//                     },
//                     children: [
//                         this.label("X:", x, `${data[i].x}`),
//                         this.label("Y:", y, `${data[i].y}`),
//                         this.label("Z:", z, `${data[i].z}`),
//                     ]
//                 },
//                 {
//                     tag: "button",
//                     text: "Delete",
//                     onclick: () => {
//                         data.splice(i, 1);
//                         setVariableValue(this.key, data).subscribe();
//                     }
//                 }
//             ]
//         }
//     }

//     label(name: string, change: Subject<string>, value: string): ElementConfig {
//         return LabledField(name, { tag: "input", change$: change, type: "number", size: 1, classList: ['variable-number-input'], value: value, step: this.step })
//     }

// }

export class Vector3ArrayVariableElement extends VariableElement {

    dataDisplay$: BehaviorSubject<Array<ElementConfig>>;
    length$: BehaviorSubject<string>;

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
        this.dataDisplay$ = new BehaviorSubject<Array<ElementConfig>>(this.buildDataDisplay());

        let data = [];

        if (this.variable.value) {
            data = this.variable.value;
        }
        this.length$ = new BehaviorSubject<string>(this.lengthDisplay());

        return {
            style: inputContainerStyle,
            children: [
                { text$: this.length$ },
                { children$: this.dataDisplay$ },
                {
                    tag: "button",
                    text: "Add",
                    onclick: () => {
                        data.push({ x: 0, y: 0, z: 0 })
                        setVariableValue(this.key, data).subscribe();
                    }
                }
            ],
        };
    }

    buildDataDisplay(): Array<ElementConfig> {
        let data = [];

        if (this.variable.value) {
            data = this.variable.value;
        }

        let dataDisplay = new Array<ElementConfig>();

        for (let i = 0; i < data.length; i++) {
            const x = bind(data[i], "x", this.mapper)
            const y = bind(data[i], "y", this.mapper)
            const z = bind(data[i], "z", this.mapper)
            x.pipe(
                combineLatestWith(y, z),
                skip(1),
                mergeMap(() => setVariableValue(this.key, data))
            ).subscribe(() => { })

            dataDisplay.push({
                style: {
                    paddingTop: "8px",
                },
                children: [
                    // { text: "" + i },
                    {
                        style: {
                            display: "flex",
                            flexDirection: "column"
                        },
                        children: [
                            this.label("X:", x, `${data[i].x}`),
                            this.label("Y:", y, `${data[i].y}`),
                            this.label("Z:", z, `${data[i].z}`),
                        ]
                    },
                    {
                        tag: "button",
                        text: "Delete",
                        onclick: () => {
                            data.splice(i, 1);
                            setVariableValue(this.key, data).subscribe();
                        }
                    }
                ]
            })
        }
        return dataDisplay;
    }

    lengthDisplay(): string {
        let data = [];

        if (this.variable.value) {
            data = this.variable.value;
        }
        return "Length: " + data.length;
    }

    label(name: string, change: Subject<string>, value: string): ElementConfig {
        return LabledField(name, { tag: "input", change$: change, type: "number", size: 1, classList: ['variable-number-input'], value: value, step: this.step })
    }

    onDestroy(): void {
        this.dataDisplay$.complete();
        this.length$.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.dataDisplay$.next(this.buildDataDisplay());
        this.length$.next(this.lengthDisplay());
    }
}