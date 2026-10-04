import { Timer } from 'three';

export interface UpdateEntry {
    name: string;
    loop: (delta: number) => void;
}

export class UpdateManager {

    timer: Timer;

    funcs: Array<UpdateEntry>;

    constructor() {
        this.timer = new Timer();
        this.funcs = [];
    }

    addToUpdate(func: UpdateEntry) {
        this.funcs.push(func);
    }

    removeFromUpdate(func: UpdateEntry) {
        const index = this.funcs.indexOf(func);
        if (index > -1) { // only splice array when item is found
            this.funcs.splice(index, 1); // 2nd parameter means remove one item only
        }
    }

    run(timestamp?: number) {
        const delta = this.timer.update(timestamp).getDelta();
        this.funcs.forEach(f => f.loop(delta));
    }
}