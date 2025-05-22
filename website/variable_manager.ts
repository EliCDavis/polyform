import { NewVariablePopup } from "./popups/new_variable";

export class VariableManager {

    constructor(parent: HTMLElement) {
        const newVariableButton = parent.querySelector("#new-variable")
        const newFolderButton = parent.querySelector("#new-folder")

        newVariableButton.addEventListener('click', (event) => {
            const popup = new NewVariablePopup();
            popup.show();
            // popups`();
        });
    }

}