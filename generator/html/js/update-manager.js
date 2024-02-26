class UpdateManager {
    constructor() {
        this.funcs = [];
    }

    addToUpdate(func) {
        this.funcs.push(func);
    }

    removeFromUpdate(func) {
        const index = this.funcs.indexOf(func);
        if (index > -1) { // only splice array when item is found
            this.funcs.splice(index, 1); // 2nd parameter means remove one item only
        } 
    }

    run() {
        this.funcs.forEach(f => f());
    }
}