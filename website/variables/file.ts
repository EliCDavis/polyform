import { BehaviorSubject } from "rxjs";
import { Variable } from "../schema";
import { uploadBinaryAsVariableValue } from "./variable_manager";
import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { ElementConfig } from "../element";
import { VariableElement } from "./variable";


function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';

    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const k = 1024;
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    const size = bytes / Math.pow(k, i);

    // Show up to one decimal if needed
    return `${size.toFixed(size < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
}

export class FileVariableElement extends VariableElement {

    text$: BehaviorSubject<string>;

    constructor(
        key: string,
        variable: Variable,
        schemaManager: SchemaManager,
        nodeManager: NodeManager
    ) {
        super(key, variable, schemaManager, nodeManager);
    }

    buildVariable(): ElementConfig {
        this.text$ = new BehaviorSubject<string>(formatFileSize(this.variable.value.size));

        return {
            style: {
                display: "flex",
                flexDirection: "column",
                gap: "8px"
            },
            children: [
                {
                    text$: this.text$,
                    style: {
                        maxWidth: "100%"
                    }
                },
                {
                    tag: "button",
                    text: "Set File",
                    onclick: () => {
                        uploadBinaryAsVariableValue(this.key, () => { });
                    }
                }
            ]
        };
    }

    onDestroy(): void {
        this.text$.complete();
    }

    set(data: Variable): void {
        this.variable = data;
        this.text$.next(formatFileSize(this.variable.value.size));
    }
}