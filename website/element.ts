import { BehaviorSubject, Subject } from "rxjs";

export type HTMLInputTypeAttribute = "button" | "checkbox" | "color" | "date" | "datetime-local" | "email" | "file" | "hidden" | "image" | "month" | "number" | "password" | "radio" | "range" | "reset" | "search" | "submit" | "tel" | "text" | "time" | "url" | "week";

export interface ElementConfig {
    /**
     * Defaults to `div` if unset
     */
    tag?: keyof HTMLElementTagNameMap;

    id?: string;
    classList?: Array<string>;

    style?: Partial<CSSStyleDeclaration>;

    text?: string;

    children?: Array<ElementConfig>

    onclick?: (this: GlobalEventHandlers, ev: MouseEvent) => any;

    /**
     * Name of the object.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/HTMLInputElement/name)
     */
    name?: string

    /**
     * Content type of the object.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/HTMLInputElement/type)
     * 
     * https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/input#input_types
     */
    type?: HTMLInputTypeAttribute;

    value?: string

    step?: string,

    change?: (e: InputEvent) => void

    change$?: Subject<string>;

    size?: number;
}

export function Element(config: ElementConfig): HTMLElement {
    let tag = "div"
    if (config.tag) {
        tag = config.tag;
    } else if (config.name || config.type || config.change) {
        tag = "input";
    }

    const newEle = document.createElement(tag);

    if (config.id) {
        newEle.id = config.id;
    }

    if (config.classList) {
        for (let i = 0; i < config.classList.length; i++) {
            newEle.classList.add(config.classList[i]);
        }
    }

    if (config.text) {
        newEle.textContent = config.text;
    }

    if (config.style) {
        Object.assign(newEle.style, config.style);
    }

    if (config.children) {
        const instantiatedChildren = new Array<HTMLElement>();
        for (let i = 0; i < config.children.length; i++) {
            instantiatedChildren.push(Element(config.children[i]))
        }
        newEle.replaceChildren(...instantiatedChildren);
    }

    if (config.name) {
        const inputEle = newEle as HTMLInputElement
        inputEle.name = config.name;
    }

    if (config.type) {
        const inputEle = newEle as HTMLInputElement
        inputEle.type = config.type;
    }

    if (config.change) {
        const inputEle = newEle as HTMLInputElement
        inputEle.addEventListener("change", config.change);
    }

    if (config.value) {
        const inputEle = newEle as HTMLInputElement
        inputEle.value = config.value;

        if (config.type === "checkbox") {
            inputEle.checked = config.value === "true";
        }
    }

    if (config.step) {
        const inputEle = newEle as HTMLInputElement
        inputEle.step = config.step;
    }

    if (config.size) {
        const inputEle = newEle as HTMLInputElement
        inputEle.size = config.size;
    }

    if (config.change$) {
        const inputEle = newEle as HTMLInputElement
        inputEle.addEventListener("change", (ev: InputEvent) => {
            if (config.type === "checkbox") {
                config.change$.next("" + inputEle.checked);
            } else {
                config.change$.next(inputEle.value);
            }
        });
    }

    if (config.onclick) {
        newEle.onclick = config.onclick;
    }

    return newEle;
}
