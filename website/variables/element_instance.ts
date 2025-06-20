import { Subscription } from "rxjs";
import { Element, ElementConfig } from "../element";

export interface IElementInstance<T> {
    element(): HTMLElement;
    set(data: T): void;
    cleanup(): void;
}

export abstract class ElementInstance<T> implements IElementInstance<T> {

    eleInstance: HTMLElement;

    subscriptions: Array<Subscription>;

    constructor() {
        this.subscriptions = new Array<Subscription>();
    }

    element(): HTMLElement {
        if (this.eleInstance) {
            return this.eleInstance
        }
        this.eleInstance = Element(this.build());
        return this.eleInstance;
    }

    abstract set(data: T): void;
    abstract onDestroy(): void;
    abstract build(): ElementConfig;

    addSubscription(subscription: Subscription): void {
        this.subscriptions.push(subscription);
    }

    cleanup(): void {
        for (let i = 0; i < this.subscriptions.length; i++) {
            this.subscriptions[i].unsubscribe();
        }
        this.onDestroy();
    }
}