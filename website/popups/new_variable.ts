import { BehaviorSubject } from "rxjs";
import { SchemaManager } from "../schema_manager";
import { VariableType } from "../variables/variable_type";
import { NodeManager } from "../node_manager";
import { CreateVariableResponse } from "../schema";
import { VariableTypeDropdown } from "./variable_type_dropdown";
import { CreatePopupElement, PopupButtonType } from "./popup";

interface NewVariableParameters {
    type: string
    description: string,
}

function inputValue(value: string | undefined, fallback: string): string {
    if (value) {
        return value;
    }
    return fallback
}

export class NewVariablePopup {

    popup: HTMLElement

    name: BehaviorSubject<string>;
    type: BehaviorSubject<string>;
    description: BehaviorSubject<string>;

    nodeManager: NodeManager;

    constructor(private schemaManager: SchemaManager, nodeManager: NodeManager) {
        this.name = new BehaviorSubject<string>("New Variable");
        this.nodeManager = nodeManager;
        this.type = new BehaviorSubject<string>(VariableType.Float);
        this.description = new BehaviorSubject<string>("");

        this.popup = CreatePopupElement({
            title: "New Variable",
            content: [{
                style: {
                    display: "flex",
                    flexDirection: "column",
                },
                children: [
                    { text: "Name" },
                    { type: "text", name: "name", change$: this.name, value: "New Variable" },

                    { text: "Description", style: { marginTop: "16px" } },
                    { tag: "textarea", type: "text", name: "description", change$: this.description },

                    { text: "Type", style: { marginTop: "16px" } },
                    VariableTypeDropdown(this.type),
                ]
            }],
            buttons: [
                { text: "Close", click: this.closePopup.bind(this) },
                { text: "Create", click: this.newClicked.bind(this), type: PopupButtonType.Primary },
            ]
        });

        document.body.appendChild(this.popup);
    }

    show(): void {
        this.popup.style.display = "flex";
    }

    closePopup(): void {
        this.popup.style.display = "none";
    }

    newClicked(): void {
        this.closePopup();

        this.newVariable({
            type: inputValue(this.type.value, "Float"),
            description: inputValue(this.description.value, ""),
        });
    }

    newVariable(parameters: NewVariableParameters): void {
        fetch("./variable/instance/" + this.name.value, {
            method: "POST",
            body: JSON.stringify(parameters)
        }).then((resp) => {
            resp.json().then((body) => {
                if (!resp.ok) {
                    alert(body.error);
                } else {
                    const createResp: CreateVariableResponse = body;
                    this.schemaManager.refreshSchema();
                    this.nodeManager.registerCustomNodeType(createResp.nodeType)
                }
            })
        });
    }
} 