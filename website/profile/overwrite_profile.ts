import { SchemaManager } from "../schema_manager";
import { Popup, PopupButtonType, PopupConfig } from "../popups/popup";

export class OverwriteProfilePopup extends Popup {

    constructor(private profile: string, private schemaManager: SchemaManager) {
        super();
    }

    protected build(): PopupConfig {
        return {
            title: "Overwrite Profile",
            content: [{
                style: {
                    display: "flex",
                    flexDirection: "column",
                },
                children: [
                    { text: "Are you sure you want to overwrite this profile with the current status of the graph?" },
                ]
            }],
            buttons: [
                { text: "Close", click: this.close.bind(this) },
                { text: "Overwrite", click: this.newClicked.bind(this), type: PopupButtonType.Primary },
            ]
        };
    }

    protected destroy(): void {
    }

    newClicked(): void {
        this.close();

        fetch("./profile/overwrite", {
            method: "POST",
            body: JSON.stringify({
                name: this.profile,
            })
        }).then((resp) => {
            if (!resp.ok) {
                alert("unable to rename profile");
            } else {
                this.schemaManager.refreshSchema("profile overwritten");
            }
        });
    }

} 