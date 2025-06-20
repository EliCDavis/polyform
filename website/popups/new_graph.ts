import { Element, ElementConfig } from "../element";
import { Popup, PopupButtonType } from "./popup";

interface NewGraphParameters {
    name: string,
    author: string,
    description: string,
    version: string,
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
                { tag: "h3", text: "Open Example", style: { marginLeft: "8px", fontWeight: "bold" } },
                { style: { "width": "170px" }, children: exampleButtons }
            ]
        };

        const newGraph: ElementConfig = {
            children: [
                { tag: "h3", text: "New", style: { fontWeight: "bold" } },

                { text: "Name" },
                { type: "text", name: "name", change: this.nameChange },

                { text: "Description", style: { marginTop: "8px" } },
                { tag: "textarea", name: "description", change: this.descriptionChange },

                { text: "Author", style: { marginTop: "8px" } },
                { type: "text", name: "author", change: this.authorChange },

                { text: "Version", style: { marginTop: "8px" } },
                { type: "text", name: "version", change: this.versionChange },
            ]
        }


        this.popup = Popup({
            title: "New Graph",
            buttons: [
                { text: "Close", click: this.closePopup.bind(this) },
                { text: "New", click: this.newClicked.bind(this), type: PopupButtonType.Primary },
            ],
            content: [{
                style: { "display": "flex" },
                children: [
                    newGraph,
                    { text: "OR", style: { "margin": "80px" } },
                    exampleGraph,
                ]
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