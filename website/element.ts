import { Observable } from "./observable";


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
     */
    type?: string

    value?: string

    change?: (e: InputEvent) => void

    value$?: Observable<string>
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
        for (let i = 0; i < config.children.length; i++) {
            newEle.appendChild(Element(config.children[i]));
        }
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
    }

    if (config.value$) {
        if (config.value) {
            config.value$.set(config.value);
        }
        
        const inputEle = newEle as HTMLInputElement
        inputEle.addEventListener("input", (ev: InputEvent) => {
            console.log(ev);
            config.value$.set(inputEle.value);
        });
        inputEle.value = config.value$.value();
    }

    if (config.onclick) {
        newEle.onclick = config.onclick;
    }

    return newEle;
}
