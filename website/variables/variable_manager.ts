import { Element, ElementConfig } from '../element';
import { NewVariablePopup } from "../popups/new_variable";
import { GraphInstance, Variable } from "../schema";
import { SchemaManager } from "../schema_manager";
import { VariableType } from './variable_type';
import { Observable, Subject } from "rxjs";
import { NodeManager } from '../node_manager';
import { Publisher } from '@elicdavis/node-flow';
import { ThreeApp } from '../three_app';
import { ElementManager } from './element_manager';
import { BasicVariableElement } from './basic_variable';
import { Vector2VariableElement } from './vector2';
import { Vector3VariableElement } from './vector3';
import { AABBVariableElement } from './aabb';
import { ImageVariableElement } from './image';
import { FileVariableElement } from './file';
import { Vector3ArrayVariableElement } from './vector3_array';
import { VariableElement } from './variable';

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

export function post$(url: string, body: BodyInit): Observable<Response> {
    const out = new Subject<Response>();
    fetch(url, {
        method: "POST",
        body: body
    }).then((resp) => {
        out.next(resp);
    });
    return out;
}


function postBinaryEmptyResponse(theUrl: string, body: any, callback): void {
    const xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = () => {
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
            callback();
        }
    }
    xmlHttp.open("POST", theUrl, true); // true for asynchronous 
    xmlHttp.send(body);
}

export function uploadBinaryAsVariableValue(variableKey: string, cb): void {
    const input = document.createElement('input');
    input.type = 'file';

    input.onchange = e => {
        const file = (e.target as HTMLInputElement).files[0];

        const reader = new FileReader();
        reader.readAsArrayBuffer(file);

        reader.onload = readerEvent => {
            const content = readerEvent.target.result as string; // this is the content!
            postBinaryEmptyResponse("./variable/value/" + variableKey, content, cb)
        }
    }

    input.click();
}


export function setVariableValue(variable: string, value: any): Observable<Response> {
    return post$("./variable/value/" + variable, JSON.stringify(value))
}



export class VariableManager {

    variableListView: Element;

    nodeManager: NodeManager;

    schemaManager: SchemaManager;

    publisher: Publisher;

    app: ThreeApp;

    elementManager: ElementManager<Variable>;

    constructor(parent: HTMLElement, schemaManager: SchemaManager, nodeManager: NodeManager, publisher: Publisher, app: ThreeApp) {
        this.nodeManager = nodeManager;
        this.schemaManager = schemaManager;
        this.publisher = publisher;
        this.app = app;

        const newVariableButton = parent.querySelector("#new-variable")
        // const newFolderButton = parent.querySelector("#new-folder")
        this.variableListView = parent.querySelector("#variable-list")

        newVariableButton.addEventListener('click', (event) => {
            const popup = new NewVariablePopup(schemaManager, this.nodeManager);
            popup.show();
        });

        schemaManager.subscribe(this.newSchemaInstance.bind(this));

        this.elementManager = new ElementManager<Variable>(this.variableListView, (a, b) => {
            return this.newVariable(a, b)
        });
    }

    newSchemaInstance(graphInstance: GraphInstance): void {
        this.elementManager.set(graphInstance.variables.variables);
    }

    newVariable(key: string, variable: Variable): VariableElement {
        console.log(variable)

        const intMap = (s: string) => parseInt(s)

        switch (variable.type) {
            case VariableType.Float:
                return new BasicVariableElement<number>(key, variable, this.schemaManager, this.nodeManager, parseFloat, "number", "");

            case VariableType.Float2:
                return new Vector2VariableElement(key, variable, this.schemaManager, this.nodeManager, parseFloat, "");

            case VariableType.Float3:
                return new Vector3VariableElement(key, variable, this.schemaManager, this.nodeManager, this.app, parseFloat, "");

            case VariableType.Int2:
                return new Vector2VariableElement(key, variable, this.schemaManager, this.nodeManager, intMap, "");

            case VariableType.Int3:
                return new Vector3VariableElement(key, variable, this.schemaManager, this.nodeManager, this.app, intMap, "");

            case VariableType.Int:
                return new BasicVariableElement<number>(key, variable, this.schemaManager, this.nodeManager, intMap, "number", "1");

            case VariableType.String:
                return new BasicVariableElement<string>(key, variable, this.schemaManager, this.nodeManager, (s) => s, "text", "");

            case VariableType.Color:
                return new BasicVariableElement<string>(key, variable, this.schemaManager, this.nodeManager, (s) => s, "color", "");

            case VariableType.Bool:
                return new BasicVariableElement<boolean>(key, variable, this.schemaManager, this.nodeManager, (s) => s === "true", "checkbox", "");

            case VariableType.AABB:
                return new AABBVariableElement(key, variable, this.schemaManager, this.nodeManager);

            case VariableType.Float3Array:
                return new Vector3ArrayVariableElement(key, variable, this.schemaManager, this.nodeManager, this.app, parseFloat, "")

            case VariableType.Image:
                return new ImageVariableElement(key, variable, this.schemaManager, this.nodeManager);

            case VariableType.File:
                return new FileVariableElement(key, variable, this.schemaManager, this.nodeManager);

            default:
                throw new Error("unimplemented variable type: " + variable.type);
        }
    }
}