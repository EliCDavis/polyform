import { IElementInstance } from "./element_instance";

export class ElementManager<T> {

    // CONFIG =================================================================
    builder: (key: string, t: T) => IElementInstance<T>;
    container: Element;

    // RUNTIME ================================================================
    built: Map<string, IElementInstance<T>>;

    constructor(container: Element, builder: (key: string, t: T) => IElementInstance<T>) {
        this.container = container;
        this.builder = builder;
        this.built = new Map<string, IElementInstance<T>>();
    }

    set(data: { [key: string]: T }): void {
        const keep = new Map<string, boolean>();
        let i = -1;
        for (const key in data) {
            keep.set(key, true);
            i++;
            if (this.built.has(key)) {
                this.built.get(key).set(data[key]);
                continue;
            }

            const instance = this.builder(key, data[key]);
            this.built.set(key, instance);
            if (i === 0) {
                this.container.append(instance.element());
            } else {
                this.container.insertBefore(instance.element(), this.container.children[i]);
            }
        }

        this.built.forEach((val, key) => {
            if (keep.has(key)) {
                return;
            }
            val.cleanup();
            this.container.removeChild(val.element());
            this.built.delete(key);
        })
    }
}
