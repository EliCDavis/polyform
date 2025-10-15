import { Element, ElementConfig } from '../element';
import { NewVariablePopup } from "../popups/new_variable";
import {  Variable } from "../schema";
import { SchemaManager } from "../schema_manager";
import { VariableType } from './variable_type';
import { map } from "rxjs";
import { NodeManager } from '../node_manager';
import { ThreeApp } from '../three_app';
import { ElementList, ListItemEntry as ListItem } from './element_manager';
import { BasicVariableElement } from './basic_variable';
import { Vector2VariableElement } from './vector2';
import { Vector3VariableElement } from './vector3';
import { AABBVariableElement } from './aabb';
import { ImageVariableElement } from './image';
import { FileVariableElement } from './file';
import { Vector3ArrayVariableElement } from './vector3_array';
import { VariableElement } from './variable';
import { RequestManager } from '../requests';

export const inputContainerStyle: Partial<CSSStyleDeclaration> = {
    display: "flex",
    flexDirection: "column",
    gap: "8px",
    flexShrink: "1",
    // paddingLeft: "8px",
    // paddingRight: "8px",
}

export function LabledField(label: string, field: ElementConfig): ElementConfig {
    return {
        style: { display: "flex", flexDirection: "row" },
        children: [
            { text: label, style: { marginRight: "8px" } },
            field,
        ]
    };
}


export class VariableManager {

    elementManager: ElementList<Variable>;

    constructor(
        parent: HTMLElement,
        private schemaManager: SchemaManager,
        private nodeManager: NodeManager,
        private app: ThreeApp,
        private requestManager: RequestManager
    ) {
        const newVariableButton = parent.querySelector("#new-variable")
        // const newFolderButton = parent.querySelector("#new-folder")

        newVariableButton.addEventListener('click', (event) => {
            const popup = new NewVariablePopup(schemaManager, this.nodeManager, requestManager);
            popup.show();
        });

        this.elementManager = new ElementList<Variable>(
            schemaManager.instance$().pipe(map((graph) => {
                const items = Array<ListItem<Variable>>();
                for (const key in graph.variables.variables) {
                    items.push({
                        key: key,
                        data: graph.variables.variables[key],
                    })
                }
                return items;
            })),
            (a, b) => {
                return this.newVariable(a, b)
            }
        );

        const variableListView = parent.querySelector("#variable-list")
        variableListView.append(Element({
            childrenManager: this.elementManager,
        }));
    }

    newVariable(key: string, variable: Variable): VariableElement {
        console.log(variable)

        const intMap = (s: string) => parseInt(s)

        switch (variable.type) {
            case VariableType.Float:
                return new BasicVariableElement<number>(key, variable, this.schemaManager, this.nodeManager, this.requestManager, parseFloat, "number", "");

            case VariableType.Float2:
                return new Vector2VariableElement(key, variable, this.schemaManager, this.nodeManager, this.requestManager, parseFloat, "");

            case VariableType.Float3:
                return new Vector3VariableElement(key, variable, this.schemaManager, this.nodeManager, this.app, this.requestManager, parseFloat, "");

            case VariableType.Int2:
                return new Vector2VariableElement(key, variable, this.schemaManager, this.nodeManager, this.requestManager, intMap, "");

            case VariableType.Int3:
                return new Vector3VariableElement(key, variable, this.schemaManager, this.nodeManager, this.app, this.requestManager, intMap, "");

            case VariableType.Int:
                return new BasicVariableElement<number>(key, variable, this.schemaManager, this.nodeManager, this.requestManager, intMap, "number", "1");

            case VariableType.String:
                return new BasicVariableElement<string>(key, variable, this.schemaManager, this.nodeManager, this.requestManager, (s) => s, "text", "");

            case VariableType.Color:
                return new BasicVariableElement<string>(key, variable, this.schemaManager, this.nodeManager, this.requestManager, (s) => s, "color", "");

            case VariableType.Bool:
                return new BasicVariableElement<boolean>(key, variable, this.schemaManager, this.nodeManager, this.requestManager, (s) => s === "true", "checkbox", "");

            case VariableType.AABB:
                return new AABBVariableElement(key, variable, this.schemaManager, this.nodeManager, this.app, this.requestManager);

            case VariableType.Float3Array:
                return new Vector3ArrayVariableElement(key, variable, this.schemaManager, this.nodeManager, this.requestManager, this.app, parseFloat, "")

            case VariableType.Image:
                return new ImageVariableElement(key, variable, this.schemaManager, this.nodeManager, this.requestManager);

            case VariableType.File:
                return new FileVariableElement(key, variable, this.schemaManager, this.nodeManager, this.requestManager);

            default:
                throw new Error("unimplemented variable type: " + variable.type);
        }
    }
}