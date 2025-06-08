import { BehaviorSubject } from "rxjs";
import { SchemaManager } from "../schema_manager";
import { GeneratorVariablePublisherPath, NodeManager } from "../node_manager";
import { CreateVariableResponse, Variable } from "../schema";
import { Popup } from "./popup";

interface EditVariableParameters {
    name: string,
    description: string,
}

const buttonStyle = {
    "padding": "8px",
    "border-radius": "8px",
}

export class EditVariablePopup {

    popup: HTMLElement

    name: BehaviorSubject<string>;

    description: BehaviorSubject<string>;

    variableKey: string;

    variable: Variable;

    nodeManager: NodeManager;

    constructor(
        private schemaManager: SchemaManager,
        nodeManager: NodeManager,
        variableKey: string,
        variable: Variable
    ) {
        this.variableKey = variableKey;
        this.variable = variable;

        this.name = new BehaviorSubject<string>(variableKey);
        this.nodeManager = nodeManager;
        this.description = new BehaviorSubject<string>(variable.description);

        this.popup = Popup([
            {
                style: {
                    display: "flex",
                    flexDirection: "column"
                },
                children: [
                    {
                        text: "Edit Variable", style: { fontWeight: "bold" }
                    },

                    { text: "Name" },
                    {
                        type: "text",
                        name: "name",
                        value: variableKey,
                        change$: this.name
                    },

                    { text: "Description" },
                    {
                        type: "text",
                        name: "description",
                        value: variable.description,
                        change$: this.description
                    },
                ]
            },
            {
                style: { marginTop: "20px" },
                children: [
                    { tag: "button", text: "Save", style: buttonStyle, onclick: this.saveClicked.bind(this) },
                    { tag: "button", text: "Cancel", style: buttonStyle, onclick: this.closePopup.bind(this) }
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

    saveClicked(): void {
        this.closePopup();
        this.updateVariable({
            "name": this.name.value,
            "description": this.description.value,
        });
    }

    updateVariable(parameters: EditVariableParameters): void {
        fetch("./variable/info/" + this.variableKey, {
            method: "POST",
            body: JSON.stringify(parameters)
        }).then((resp) => {
            resp.json().then((body) => {
                if (!resp.ok) {
                    alert(body.error);
                } else {
                    this.schemaManager.refreshSchema();
                    this.nodeManager.updateNodeInfo(
                        GeneratorVariablePublisherPath + this.variableKey,
                        GeneratorVariablePublisherPath + this.name.value,
                        this.name.value,
                        this.description.value
                    )
                }
            })
        });
    }
} 