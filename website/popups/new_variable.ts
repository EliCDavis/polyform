import { BehaviorSubject } from "rxjs";
import { Element, ElementConfig } from "../element";
import { SchemaManager } from "../schema_manager";
import { VariableType } from "../variable_type";
import { Publisher } from "@elicdavis/node-flow";
import { NodeManager } from "../node_manager";
import { CreateVariableResponse } from "../schema";

interface NewVariableParameters {
    name: string,
    description: string,
    type: string
}


const buttonStyle = {
    "padding": "8px",
    "border-radius": "8px",
}

const NewGraphPopupStyle = {
    "position": "fixed",
    "justify-content": "center",
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

        const newGraph: ElementConfig = {
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
                {
                    tag: "select",
                    change$: this.type,
                    children: [
                        { tag: "option", value: VariableType.Float, text: "Float" },
                        { tag: "option", value: VariableType.Float2, text: "Float2" },
                        { tag: "option", value: VariableType.Float3, text: "Float3" },
                        { tag: "option", value: VariableType.Int, text: "Int" },
                        { tag: "option", value: VariableType.Int2, text: "Int2" },
                        { tag: "option", value: VariableType.Int3, text: "Int3" },
                        { tag: "option", value: VariableType.String, text: "String" },
                        { tag: "option", value: VariableType.Bool, text: "Bool" },
                        { tag: "option", value: VariableType.AABB, text: "AABB" },
                        { tag: "option", value: VariableType.Color, text: "Color" },
                    ]
                },
            ]
        }

        const popupContents: ElementConfig = {
            style: {
                backgroundColor: "#00000069",
                backdropFilter: "blur(10px)",
                padding: "24px",
                borderRadius: "24px",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
            },
            children: [
                newGraph,
                {
                    style: { marginTop: "20px" },
                    children: [
                        { tag: "button", text: "Create", style: buttonStyle, onclick: this.newClicked.bind(this) },
                        { tag: "button", text: "Close", style: buttonStyle, onclick: this.closePopup.bind(this) }
                    ]
                }
            ]
        };

        this.popup = Element({
            style: {
                position: "absolute",
                width: "100%",
                height: "100%",
                backgroundColor: "rgba(0,0,0,0.5)",
                top: "0",
                left: "0",
                display: "none",
                justifyContent: "center",
                alignItems: "center"
            },
            children: [{
                style: NewGraphPopupStyle,
                children: [popupContents]
            }]
        })

        document.body.appendChild(this.popup);
    }

    inputValue(value: string | undefined, fallback: string): string {
        if (value) {
            return value;
        }
        return fallback
    }

    show(): void {
        this.popup.style.display = "flex";
    }

    VariableParametersFromPopup(): NewVariableParameters {
        return {
            "name": this.inputValue(this.name.value, "New Variable"),
            "type": this.inputValue(this.type.value, "Float"),
            "description": this.inputValue(this.description.value, ""),
        }
    }

    closePopup(): void {
        this.popup.style.display = "none";
    }

    newClicked(): void {
        this.closePopup();
        this.newVariable(this.VariableParametersFromPopup());
    }

    newVariable(parameters: NewVariableParameters): void {
        console.log(parameters);
        fetch("./variable/instance/" + parameters.name.replace(/\s/g, ''), {
            method: "POST",
            body: JSON.stringify(parameters)
        }).then((resp) => {
            resp.json().then((body) => {
                if (!resp.ok) {
                    alert(body.error);
                } else {
                    const createresp: CreateVariableResponse = body;
                    this.schemaManager.refreshSchema();
                    this.nodeManager.registerCustomNodeType(createresp.nodeType)
                    console.log(createresp)
                }
            })
        });
    }

    exampleClicked(example: string): void {
        this.closePopup();
        // fetch("./load-example", { method: "POST", body: example })
        //     .then((resp) => {
        //         if (resp.ok) {
        //             location.reload();
        //         } else {
        //             console.error(resp);
        //         }
        //     });
    }
} 