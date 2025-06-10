import { BehaviorSubject } from "rxjs";
import { SchemaManager } from "../schema_manager";
import { VariableType } from "../variable_type";
import { NodeManager } from "../node_manager";
import { CreateVariableResponse } from "../schema";
import { VariableTypeDropdown } from "./variable_type_dropdown";
import { Popup } from "./popup";

interface NewVariableParameters {
    variable: {
        type: string
    }
    description: string,
}

const buttonStyle = {
    "padding": "8px",
    "border-radius": "8px",
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

        this.popup = Popup([
            {
                style: {
                    display: "flex",
                    flexDirection: "column"
                },
                children: [
                    {
                        text: "New Variable", style: { fontWeight: "bold" }
                    },

                    { text: "Name" },
                    { type: "text", name: "name", change$: this.name },

                    { text: "Description" },
                    { type: "text", name: "description", change$: this.description },

                    { text: "Type" },
                    VariableTypeDropdown(this.type),
                ]
            },
            {
                style: { marginTop: "20px" },
                children: [
                    { tag: "button", text: "Create", style: buttonStyle, onclick: this.newClicked.bind(this) },
                    { tag: "button", text: "Close", style: buttonStyle, onclick: this.closePopup.bind(this) }
                ]
            }
        ]);

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
            variable: {
                type: inputValue(this.type.value, "Float"),
            },
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