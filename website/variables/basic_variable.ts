import { map, mergeMap, Subject } from "rxjs";
import { Variable } from "../schema";
import { setVariableValue } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig, HTMLInputTypeAttribute } from "../element";
import { VariableElement } from "./variable";

export class BasicVariableElement<T> extends VariableElement {

    value$: Subject<string>;

    constructor(
        key: string,
        variable: Variable,
        schemaManager: SchemaManager,
        nodeManager: NodeManager,
        private mapper: (s: string) => T,
        private inputType: HTMLInputTypeAttribute,
        private step: string
    ) {
        super(key, variable, schemaManager, nodeManager);
    }

    buildVariable(): ElementConfig {
        const change$ = new Subject<string>();
        this.value$ = new Subject<string>();

        this.addSubscription(change$.pipe(
            map(this.mapper),
            mergeMap((val) => setVariableValue(this.key, val))
        ).subscribe((resp: Response) => {
            console.log(resp);
        }))

        return {
            tag: "input",
            type: this.inputType,
            change$: change$,
            value: `${this.variable.value}`,
            value$: this.value$,
            step: this.step,
            size: 1
        };
    }

    onDestroy(): void {
        this.value$.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.value$.next(`${this.variable.value}`)
    }
}