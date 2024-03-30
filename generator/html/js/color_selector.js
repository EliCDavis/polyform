
export class ColorSelector {

    /**
     * 
     * @param {string} selectorContainerId 
     */
    constructor(selectorContainerId) {
        this.okayCallback = null;
        this.cancelCallback = null;

        this.selectorContainer = document.getElementById(selectorContainerId);
        this.input = document.getElementById("colorSelectorInput");
        this.okButton = document.getElementById("colorSelectorOK");
        this.cancelButton = document.getElementById("colorSelectorCancel");

        this.okButton.onclick = () => {
            this.hide();
            if (this.okayCallback) {
                this.okayCallback(this.input.value);
            }
        }

        this.cancelButton.onclick = () => {
            this.hide();
            if (this.cancelCallback) {
                this.cancelCallback();
            }
        }

        this.hide();
    }

    hide() {
        this.selectorContainer.style.display = "none";
    }

    /**
     * 
     * @param {string} value 
     * @param {(col: string) => null} okay 
     * @param {() => null} cancel 
     */
    show(value, okay, cancel) {
        this.selectorContainer.style.display = "flex";
        this.input.value = value;
        this.okayCallback = okay;
        this.cancelCallback = cancel;
    }

}