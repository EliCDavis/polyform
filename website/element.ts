

export interface ElementConfig {
    /**
     * Defaults to `div` if unset
     */
    tag?: keyof HTMLElementTagNameMap;

    id?: string;
    classList?: Array<string>;
    style?: {
        [name: string]: string
    }

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

    change?: (e: InputEvent) => void
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
        for (const key in config.style) {
            newEle.style[key] = config.style[key];
        }
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

    if (config.onclick) {
        newEle.onclick = config.onclick;
    }

    return newEle;
}
