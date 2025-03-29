
export class PolyformErrorManager {

    errors: Map<string, HTMLDivElement>;

    constructor() {
        this.errors = new Map<string, HTMLDivElement>();
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
    ShowError(key: string, message: string): void {
        if (this.errors.has(key)) {
            this.errors.get(key).innerText = message;
            return;
        }

        const error = document.createElement("div");
        error.className = "errorMessage";
        this.parent().appendChild(error);
        error.innerText = message;

        this.errors.set(key, error);
    }

    /**
     * Deletes a error if one exists with that key. If not, nothing happens.
     * 
     * @param {string} key key to the error
     */
    ClearError(key: string): void {
        if (!this.errors.has(key)) {
            return;
        }
        this.parent().removeChild(this.errors.get(key));
        this.errors.delete(key);
    }
}

export const ErrorManager = new PolyformErrorManager();
