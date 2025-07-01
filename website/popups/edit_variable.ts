import { BehaviorSubject } from "rxjs";
import { SchemaManager } from "../schema_manager";
import { GeneratorVariablePublisherPath, NodeManager } from "../node_manager";
import { CreateVariableResponse, Variable } from "../schema";
import { CreatePopupElement, PopupButtonType } from "./popup";

interface EditVariableParameters {
    name: string,
    description: string,
}



export class EditVariablePopup {

    popup: HTMLElement

    name: BehaviorSubject<string>;

    description: BehaviorSubject<string>;

    variableKey: string;

    variable: Variable;


    constructor(
        variableKey: string,
        variable: Variable
    ) {
        this.variableKey = variableKey;
        this.variable = variable;

        this.name = new BehaviorSubject<string>(variableKey);
        this.description = new BehaviorSubject<string>(variable.description);

        this.popup = CreatePopupElement({
            title: "Edit Variable",
            buttons: [
                { text: "Cancel", click: this.closePopup.bind(this) },
                { text: "Save", click: this.saveClicked.bind(this), type: PopupButtonType.Primary },
            ],
            content: [{
                style: {
                    display: "flex",
                    flexDirection: "column",
                    width: "400px"
                },
                children: [
                    { text: "Name" },
                    {
                        type: "text",
                        name: "name",
                        style: {flex: "1"},
                        value: variableKey,
                        change$: this.name
                    },

                    { text: "Description", style: { marginTop: "16px" } },
                    {
                        tag: "textarea",
                        type: "text",
                        name: "description",
                        value: variable.description,
                        change$: this.description
                    },
                ]
            }]
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
        if (this.name.value === this.variableKey && this.description.value === this.variable.description) {
            return;
        }
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
                    location.reload();
                    // this.schemaManager.refreshSchema();
                    // this.nodeManager.updateVariableInfo(
                    //     GeneratorVariablePublisherPath + this.variableKey,
                    //     GeneratorVariablePublisherPath + this.name.value,
                    //     this.name.value,
                    //     this.description.value
                    // )
                }
            })
        });
    }
} 