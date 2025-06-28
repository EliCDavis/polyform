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

        let styling: Partial<CSSStyleDeclaration> = {}
        if (this.inputType === "color") {
            styling = {
                minHeight: "25px",
                width: "25px",
                maxWidth: "25px",
                padding: "0",
                cursor: "pointer"
            }
        }

        const inputEle: ElementConfig = {
            tag: "input",
            type: this.inputType,
            change$: change$,
            value: `${this.variable.value}`,
            value$: this.value$,
            style: styling,
            step: this.step,
            size: 1
        };

        if (this.inputType === "color") {
            return {
                style: {
                    flexDirection: "row",
                    display: "flex",
                    gap: "16px"
                },
                children: [
                    inputEle,
                    {
                        text: this.variable.value,
                        text$: this.value$,
                    }
                ]
            };
        }

        return inputEle;
    }

    onDestroy(): void {
        this.value$.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.value$.next(`${this.variable.value}`)
    }
}