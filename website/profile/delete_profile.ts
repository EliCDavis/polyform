import { SchemaManager } from "../schema_manager";
import { Popup, PopupButtonType, PopupConfig } from "../popups/popup";

export class DeleteProfilePopup extends Popup {

    constructor(private profile: string, private schemaManager: SchemaManager) {
        super();
    }

    protected build(): PopupConfig {
        return {
            title: "Delete Profile",
            content: [{
                text: "Are you sure you want to delete this profile?",
            }],
            buttons: [
                { text: "Close", click: this.close.bind(this) },
                { text: "Overwrite", click: this.delete.bind(this), type: PopupButtonType.Destructive },
            ]
        };
    }

    protected destroy(): void {
    }

    delete(): void {
        this.close();

        fetch("./profile", {
            method: "DELETE",
            body: JSON.stringify({
                name: this.profile,
            })
        }).then((resp) => {
            if (!resp.ok) {
                alert("unable to delete profile");
            } else {
                this.schemaManager.refreshSchema();
            }
        });
    }

} 