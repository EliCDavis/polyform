import { Element, ElementConfig } from "../element";
import { Observable } from '../observable';

interface NewVariableParameters {
    name: string,
    description: string,
    type: string
}

enum VariableType {
    Float = "float64",
    Float2 = "Float2",
    Float3 = "Float3",
    Int = "int",
    Int2 = "Int2",
    Int3 = "Int3",
    String = "string",
    Bool = "bool",
    AABB = "AABB",
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

    name: Observable<string>;
    type: Observable<string>;
    description: Observable<string>;

    constructor() {
        this.name = new Observable<string>("New Variable");
        this.type = new Observable<string>(VariableType.Float);
        this.description = new Observable<string>("");

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
                { type: "text", name: "name", value$: this.name },

                { text: "Description" },
                { type: "text", name: "description", value$: this.description },

                { text: "Type" },
                {
                    tag: "select",
                    value$: this.type,
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
            "name": this.inputValue(this.name.value(), "New Variable"),
            "type": this.inputValue(this.type.value(), "Float"),
            "description": this.inputValue(this.description.value(), ""),
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
            if (!resp.ok) {
                console.error(resp);
                // location.reload();
            }
            console.log(resp)
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