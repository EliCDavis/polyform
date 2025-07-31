import { BehaviorSubject } from "rxjs";
import { SchemaManager } from "../schema_manager";
import { Popup, PopupButtonType, PopupConfig } from "../popups/popup";

const defaultProfileName = "New Profile";

function inputValue(value: string | undefined, fallback: string): string {
    if (value) {
        return value;
    }
    return fallback
}

export class NewProfilePopup extends Popup {

    name: BehaviorSubject<string>;

    constructor(private schemaManager: SchemaManager) {
        super();
    }

    protected build(): PopupConfig {
        this.name = new BehaviorSubject<string>(defaultProfileName);
        return {
            title: "New Profile",
            content: [{
                style: {
                    display: "flex",
                    flexDirection: "column",
                },
                children: [
                    { type: "text", name: "name", change$: this.name, value: defaultProfileName },
                ]
            }],
            buttons: [
                { text: "Close", click: this.close.bind(this) },
                { text: "Create", click: this.newClicked.bind(this), type: PopupButtonType.Primary },
            ]
        };
    }

    protected destroy(): void {
        this.name.complete();
    }

    newClicked(): void {
        this.close();

        fetch("./profile", {
            method: "POST",
            body: JSON.stringify({
                name: inputValue(this.name.value, defaultProfileName),
            })
        }).then((resp) => {
            if (!resp.ok) {
                alert("unable to create profile");
            } else {
                this.schemaManager.refreshSchema("created a profile");
            }
        });
    }

} 