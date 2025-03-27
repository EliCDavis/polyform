import { Clock } from 'three';

export interface UpdateEntry {
    name: string;
    loop: (delta: number) => void;
}

export class UpdateManager {

    clock: Clock;

    funcs: Array<UpdateEntry>;

    constructor() {
        this.clock = new Clock();
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

    run() {
        const delta = this.clock.getDelta();
        this.funcs.forEach(f => f.loop(delta));
    }
}