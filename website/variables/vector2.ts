import { BehaviorSubject, combineLatestWith, map, mergeMap, skip, Subject } from "rxjs";
import { Variable } from "../schema";
import { inputContainerStyle, LabledField } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig } from "../element";
import { VariableElement } from "./variable";
import { RequestManager } from "../requests";

export class Vector2VariableElement extends VariableElement {

    valueX$: Subject<string>;
    valueY$: Subject<string>;

    constructor(
        key: string,
        variable: Variable,
        schemaManager: SchemaManager,
        nodeManager: NodeManager,
        private requestManager: RequestManager,
        private mapper: (s: string) => number,
        private step: string
    ) {
        super(key, variable, schemaManager, nodeManager);
    }

    buildVariable(): ElementConfig {
        const changeX$ = new BehaviorSubject<string>(`${this.variable.value.x}`);
        const changeY$ = new BehaviorSubject<string>(`${this.variable.value.y}`);
        this.valueX$ = new Subject<string>();
        this.valueY$ = new Subject<string>();

        this.addSubscription(changeX$.pipe(
            map(this.mapper),
            combineLatestWith(changeY$.pipe(map(this.mapper))),
            skip(1),
            mergeMap((val) => this.requestManager.setVariableValue(this.key, { x: val[0], y: val[1] }))
        ).subscribe((resp: Response) => {
            console.log(resp);
        }));

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
            ]
        };
    }

    onDestroy(): void {
        this.valueX$.complete();
        this.valueY$.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.valueX$.next(`${this.variable.value.x}`)
        this.valueY$.next(`${this.variable.value.y}`)
    }
}