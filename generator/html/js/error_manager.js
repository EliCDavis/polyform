
class PolyformErrorManager {

    constructor() {
        this.errors = {}
    }

    parent() {
        return document.getElementById("messageContainer");
    }

    /**
     * Show an error for a specific key
     * 
     * @param {string} key 
     * @param {string} message 
     */
    ShowError(key, message) {

        if (key in this.errors) {
            this.errors[key].innerText = message;
            return;
        }

        const error = document.createElement("div");
        error.className = "errorMessage";
        this.parent().appendChild(error);
        error.innerText = message;

        this.errors[key] = error;
    }

    /**
     * Deletes a error if one exists with that key. If not, nothing happens.
     * 
     * @param {string} key key to the error
     */
    ClearError(key) {
        if (key in this.errors) {
            this.parent().removeChild(this.errors[key]);
            delete this.errors[key];
        }
    }
}

const ErrorManager = new PolyformErrorManager();
