import { SchemaManager } from "../schema_manager";
import { Popup, PopupButtonType, PopupConfig } from "../popups/popup";
import { RequestManager } from "../requests";

export class DeleteProfilePopup extends Popup {

    constructor(
        private profile: string, 
        private schemaManager: SchemaManager,
        private requestManager: RequestManager
    ) {
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
        this.requestManager.deleteProfile(
            this.profile,
            () => this.schemaManager.refreshSchema("deleted a variable profile"),
            () => alert("unable to delete profile")
        );
    }

} 