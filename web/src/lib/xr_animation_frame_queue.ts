// Quest pauses window.requestAnimationFrame in immersive sessions; install before entering so earlier callbacks flush too.
export class AnimationFrameQueue {
    #pending = new Map<number, { callback: FrameRequestCallback, nativeId: number }>();
    // Above native ids so ids issued before the wrap still cancel natively.
    #nextId = 1_000_000_000;
    #cancel: (id: number) => void;

    constructor() {
        const request = window.requestAnimationFrame.bind(window);
        this.#cancel = window.cancelAnimationFrame.bind(window);

        window.requestAnimationFrame = (callback) => {
            const id = this.#nextId++;
            const nativeId = request((time) => this.#run(id, time));
            this.#pending.set(id, { callback, nativeId });
            return id;
        };
        window.cancelAnimationFrame = (id) => {
            const entry = this.#pending.get(id);
            if (entry === undefined) {
                this.#cancel(id);
                return;
            }
            this.#pending.delete(id);
            this.#cancel(entry.nativeId);
        };
    }

    #run(id: number, time: number): void {
        const entry = this.#pending.get(id);
        if (entry === undefined) {
            return;
        }
        this.#pending.delete(id);
        entry.callback(time);
    }

    flush(): boolean {
        if (this.#pending.size === 0) {
            return false;
        }
        const entries = [...this.#pending.values()];
        this.#pending.clear();
        const now = performance.now();
        entries.forEach((entry) => {
            this.#cancel(entry.nativeId);
            entry.callback(now);
        });
        return true;
    }
}
