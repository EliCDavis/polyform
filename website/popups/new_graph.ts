import { Element, ElementConfig } from "../element";

interface NewGraphParameters {
    name: string,
    author: string,
    description: string,
    version: string,
}

const buttonStyle = {
    "padding": "8px",
    "border-radius": "8px",
}

const NewGraphPopupStyle = {
    "position": "fixed",
    "top": "20%",
    "width": "100%",
    "display": "none",
    "justify-content": "center",
}

export class NewGraphPopup {

    popup: HTMLElement

    name: string;
    description: string;
    author: string;
    version: string;

    constructor(exampleGraphs: Array<string>) {

        const exampleButtons = new Array<ElementConfig>();
        for (let i = 0; i < exampleGraphs.length; i++) {
            const element = exampleGraphs[i];
            exampleButtons.push({
                text: element,
                classList: ["example-graph-item"],
                onclick: () => {
                    this.exampleClicked(element);
                }
            });
        }

        const exampleGraph: ElementConfig = {
            children: [
                { text: "Open Example", style: { "margin-left": "8px", "font-weight": "bold" } },
                { style: { "width": "170px" }, children: exampleButtons }
            ]
        };

        const newGraph: ElementConfig = {
            children: [
                { text: "New", style: { "font-weight": "bold" } },

                { text: "Graph New" },
                { type: "text", name: "name", change: this.nameChange },

                { text: "Graph Description" },
                { type: "text", name: "description", change: this.descriptionChange },

                { text: "Author" },
                { type: "text", name: "author", change: this.authorChange },

                { text: "Version" },
                { type: "text", name: "version", change: this.versionChange },
            ]
        }

        const popupContents: ElementConfig = {
            style: {
                "background-color": "#00000069",
                "backdrop-filter": "blur(10px)",
                "padding": "24px",
                "border-radius": "24px",
                "display": "flex",
                "flex-direction": "column",
                "align-items": "center",
            },
            children: [
                {
                    style: { "display": "flex" },
                    children: [
                        newGraph,
                        { text: "OR", style: { "margin": "80px" } },
                        exampleGraph,
                    ]
                },

                {
                    style: { "margin-top": "20px" },
                    children: [
                        { tag: "button", text: "New", style: buttonStyle, onclick: this.newClicked.bind(this) },
                        { tag: "button", text: "Close", style: buttonStyle, onclick: this.closePopup.bind(this) }
                    ]
                }
            ]
        };

        this.popup = Element({
            style: NewGraphPopupStyle,
            children: [popupContents]
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

    GraphParametersFromPopup(): NewGraphParameters {
        return {
            "name": this.inputValue(this.name, "New Graph"),
            "author": this.inputValue(this.author, ""),
            "description": this.inputValue(this.description, ""),
            "version": this.inputValue(this.version, "v0.0.0"),
        }
    }

    closePopup(): void {
        this.popup.style.display = "none";
    }

    nameChange(evt: InputEvent): void {
        this.name = evt.data;
    }

    authorChange(evt: InputEvent): void {
        this.author = evt.data;
    }

    versionChange(evt: InputEvent): void {
        this.version = evt.data;
    }

    descriptionChange(evt: InputEvent): void {
        this.description = evt.data;
    }


    newClicked(): void {
        this.closePopup();
        this.newGraph({
            author: this.author,
            description: this.description,
            name: this.name,
            version: this.version
        });
    }

    newGraph(parameters: NewGraphParameters): void {
        this.closePopup();
        fetch("./new-graph", {
            method: "POST",
            body: JSON.stringify(parameters)
        }).then((resp) => {
            if (resp.ok) {
                location.reload();
            } else {
                console.error(resp);
            }
        });
    }

    exampleClicked(example: string): void {
        this.closePopup();
        fetch("./load-example", { method: "POST", body: example })
            .then((resp) => {
                if (resp.ok) {
                    location.reload();
                } else {
                    console.error(resp);
                }
            });
    }
} 