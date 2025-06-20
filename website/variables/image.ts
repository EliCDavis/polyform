import { BehaviorSubject } from "rxjs";
import { Variable } from "../schema";
import { uploadBinaryAsVariableValue } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig } from "../element";
import { VariableElement } from "./variable";

export class ImageVariableElement extends VariableElement {

    children$: BehaviorSubject<Array<ElementConfig>>;

    constructor(
        key: string,
        variable: Variable,
        schemaManager: SchemaManager,
        nodeManager: NodeManager
    ) {
        super(key, variable, schemaManager, nodeManager);
    }

    buildVariable(): ElementConfig {
        const conf: ElementConfig = {
            tag: "img",
            src: "./variable/value/" + this.key,
            style: {
                maxWidth: "100%"
            }
        };
        this.children$ = new BehaviorSubject<Array<ElementConfig>>([conf]);

        return {
            style: {
                display: "flex",
                flexDirection: "column",
                gap: "8px"
            },
            children: [
                { children$: this.children$ },
                {
                    tag: "button",
                    text: "Set Image",
                    onclick: () => {
                        uploadBinaryAsVariableValue(this.key, () => {
                            this.children$.next([conf])
                        });
                    }
                }
            ]
        };
    }

    onDestroy(): void {
        this.children$.complete();
        // Left intentionally blank
    }

    set(data: Variable): void {
        this.variable = data;
    }
}