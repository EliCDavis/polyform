import { Element, ElementConfig } from "../element";
import { CreatePopupElement, PopupButtonType } from "./popup";

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
                { type: "text", name: "name", change: this.nameChange.bind(this), },

                { text: "Description", style: { marginTop: "8px" } },
                { tag: "textarea", name: "description", change: this.descriptionChange.bind(this) },

                { text: "Author", style: { marginTop: "8px" } },
                { type: "text", name: "author", change: this.authorChange.bind(this) },

                { text: "Version", style: { marginTop: "8px" } },
                { type: "text", name: "version", change: this.versionChange.bind(this) },
            ]
        }


        this.popup = CreatePopupElement({
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
        this.name = (evt.target as any).value;
    }

    authorChange(evt: InputEvent): void {
        this.author = (evt.target as any).value;
    }

    versionChange(evt: InputEvent): void {
        this.version = (evt.target as any).value;
    }

    descriptionChange(evt: InputEvent): void {
        this.description = (evt.target as any).value;
    }


    newClicked(): void {
        this.newGraph({
            author: this.author,
            description: this.description,
            name: this.name,
            version: this.version
        });
        this.closePopup();
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