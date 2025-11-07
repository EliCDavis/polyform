import { BehaviorSubject } from "rxjs";
import { SchemaManager } from "../schema_manager";
import { Popup, PopupButtonType, PopupConfig } from "../popups/popup";
import { RequestManager } from "../requests";

export class RenameProfilePopup extends Popup {

    name: BehaviorSubject<string>;

    constructor(
        private profile: string,
        private schemaManager: SchemaManager,
        private requestManager: RequestManager
    ) {
        super();
    }

    protected build(): PopupConfig {
        this.name = new BehaviorSubject<string>(this.profile);
        return {
            title: "Rename Profile",
            content: [{
                style: {
                    display: "flex",
                    flexDirection: "column",
                },
                children: [
                    { type: "text", name: "name", change$: this.name, value: this.profile },
                ]
            }],
            buttons: [
                { text: "Close", click: this.close.bind(this) },
                { text: "Rename", click: this.newClicked.bind(this), type: PopupButtonType.Primary },
            ]
        };
    }

    protected destroy(): void {
        this.name.complete();
    }

    newClicked(): void {
        this.close();

        // Nothing to update
        if (this.name.value === this.profile) {
            return;
        }

        if (this.name.value.trim() === "") {
            alert("Name can not be empty");
            return;
        }

        this.requestManager.renameProfile(
            this.profile,
            this.name.value,
            () => this.schemaManager.refreshSchema("renamed a profile"),
            () => alert("unable to rename profile")
        );
    }

} 