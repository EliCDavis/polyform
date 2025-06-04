import { SchemaManager } from "../schema_manager";
import { NodeManager } from "../node_manager";
import { Variable } from "../schema";
import { Popup } from "./popup";

const buttonStyle = {
    "padding": "8px",
    "border-radius": "8px",
}

export class DeleteVariablePopup {

    popup: HTMLElement

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

        this.nodeManager = nodeManager;

        this.popup = Popup([
            {
                style: {
                    display: "flex",
                    flexDirection: "column"
                },
                children: [
                    {
                        text: "Delete Variable", style: { fontWeight: "bold" }
                    },

                    {
                        text: "Are you sure you want to delete " + variable.name
                    },
                ]
            },
            {
                style: { marginTop: "20px" },
                children: [
                    { tag: "button", text: "Delete", style: buttonStyle, onclick: this.saveClicked.bind(this) },
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
        this.deleteVariable();
    }

    deleteVariable(): void {
        fetch("./variable/instance/" + this.variableKey, {
            method: "DELETE",
        }).then((resp) => {
            if (!resp.ok) {
                alert("Error deleting variable");
                console.log(resp);
            } else {
                this.schemaManager.refreshSchema();
                // this.nodeManager.registerCustomNodeType(createResp.nodeType)
            }
        });
    }
}