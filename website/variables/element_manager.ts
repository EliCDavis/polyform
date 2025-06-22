import { Observable } from "rxjs";
import { IChildrenManager } from "../element";
import { IElementInstance } from "./element_instance";

export interface ListItemEntry<T> {
    key: string;
    data: T
}

export class ElementList<T> implements IChildrenManager {

    // CONFIG =================================================================
    builder: (key: string, t: T) => IElementInstance<T>;

    // RUNTIME ================================================================
    container: Element;
    built: Map<string, IElementInstance<T>>;
    cached: Array<ListItemEntry<T>>;

    constructor(data$: Observable<Array<ListItemEntry<T>>>, builder: (key: string, t: T) => IElementInstance<T>) {
        this.builder = builder;
        this.built = new Map<string, IElementInstance<T>>();
        data$.subscribe((data) => {
            if (this.container) {
                this.set(data);
            } else {
                this.cached = data;
            }
        })
    }

    setContainer(container: HTMLElement) {
        this.container = container;
        if (this.cached) {
            this.set(this.cached)
        }
    }

    private set(data: Array<ListItemEntry<T>>): void {

        // TODO: Figure out how re-ordering should behave

        const keep = new Map<string, boolean>();
        for (let i = 0; i < data.length; i++) {
            const entry = data[i];
            const key = entry.key;
            keep.set(key, true);
            if (this.built.has(key)) {
                this.built.get(key).set(entry.data);
                continue;
            }

            const instance = this.builder(key, entry.data);
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
