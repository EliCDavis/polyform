import { SchemaManager } from "../schema_manager";
import { Popup, PopupButtonType, PopupConfig } from "../popups/popup";
import { RequestManager } from "../requests";

export class OverwriteProfilePopup extends Popup {

    constructor(
        private profile: string,
        private schemaManager: SchemaManager,
        private requestManager: RequestManager
    ) {
        super();
    }

    protected build(): PopupConfig {
        return {
            title: "Overwrite Profile",
            content: [{
                style: { display: "flex", flexDirection: "column", },
                children: [{
                    text: `Are you sure you want to overwrite "${this.profile}" with the current state of the graph?`
                }]
            }],
            buttons: [
                { text: "Close", click: this.close.bind(this) },
                { text: "Overwrite", click: this.newClicked.bind(this), type: PopupButtonType.Primary },
            ]
        };
    }

    protected destroy(): void { }

    newClicked(): void {
        this.close();
        this.requestManager.overwriteProfile(
            this.profile,
            () => this.schemaManager.refreshSchema("profile overwritten"),
            () => alert("unable to rename profile")
        );
    }

} 