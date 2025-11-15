import { SchemaManager } from "../schema_manager";
import { GeneratorVariablePublisherPath, NodeManager } from "../node_manager";
import { Variable } from "../schema";
import { CreatePopupElement, PopupButtonType } from "./popup";
import { RequestManager } from "../requests";

export class DeleteVariablePopup {

    popup: HTMLElement

    variableKey: string;

    variable: Variable;

    nodeManager: NodeManager;

    constructor(
        private schemaManager: SchemaManager,
        private requestManager: RequestManager,
        nodeManager: NodeManager,
        variableKey: string,
        variable: Variable
    ) {
        this.variableKey = variableKey;
        this.variable = variable;
        this.nodeManager = nodeManager;

        this.popup = CreatePopupElement({
            title: "Delete Variable",
            content: [{
                text: "Are you sure you want to delete " + this.variableKey
            }],
            buttons: [
                { text: "Cancel", click: this.closePopup.bind(this) },
                { text: "Delete", click: this.saveClicked.bind(this), type: PopupButtonType.Destructive },
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

    saveClicked(): void {
        this.closePopup();
        this.deleteVariable();
    }

    deleteVariable(): void {
        this.requestManager.deleteVariable(
            this.variableKey,
            () => {
                this.schemaManager.refreshSchema("Deleted a variable");
                this.nodeManager.unregisterNodeType(GeneratorVariablePublisherPath + this.variableKey)
            },
            (resp) => {
                alert("Error deleting variable");
                console.log(resp);
            }
        )
    }
}