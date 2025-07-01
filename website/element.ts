import { Observable, Subject } from "rxjs";

export type HTMLInputTypeAttribute = "button" | "checkbox" | "color" | "date" | "datetime-local" | "email" | "file" | "hidden" | "image" | "month" | "number" | "password" | "radio" | "range" | "reset" | "search" | "submit" | "tel" | "text" | "time" | "url" | "week";

export interface IChildrenManager {
    setContainer(container: HTMLElement);
}

export interface ElementConfig {
    /**
     * Defaults to `div` if unset
     */
    tag?: keyof HTMLElementTagNameMap;

    id?: string;
    classList?: Array<string>;

    style?: Partial<CSSStyleDeclaration>;
    style$?: Observable<Partial<CSSStyleDeclaration>>;

    text?: string;
    text$?: Observable<string>;

    children?: Array<ElementConfig>
    children$?: Observable<Array<ElementConfig>>
    childrenManager?: IChildrenManager

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
    value$?: Observable<string>

    step?: string,

    change?: (e: InputEvent) => void

    change$?: Subject<string>;

    size?: number;

    src?: string;
}

function replaceChildren(ele: HTMLElement, children: Array<ElementConfig>): void {
    const instantiatedChildren = new Array<HTMLElement>();
    for (let i = 0; i < children.length; i++) {
        if (!children[i]) {
            continue;
        }
        instantiatedChildren.push(Element(children[i]))
    }
    ele.replaceChildren(...instantiatedChildren);
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

    if (config.text$) {
        config.text$.subscribe(text => {
            newEle.textContent = text;
        })
    }

    if (config.style) {
        Object.assign(newEle.style, config.style);
    }

    if (config.src) {
        const img = newEle as HTMLImageElement;
        img.src = config.src;
    }

    if (config.style$) {
        config.style$.subscribe(styling => {
            Object.assign(newEle.style, styling);
        });
    }

    if (config.children) {
        replaceChildren(newEle, config.children);
    }

    if (config.children$) {
        config.children$.subscribe((newChildren) => {
            if (newChildren) {
                replaceChildren(newEle, newChildren);
            } else {
                newEle.replaceChildren();
            }
        });
    }

    if (config.childrenManager) {
        config.childrenManager.setContainer(newEle);
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

    if (config.value$) {
        const inputEle = newEle as HTMLInputElement
        config.value$.subscribe(newVal => {

            // Prevent change event from being raised when no change has occurred
            if (inputEle.value === newVal) {
                return;
            }

            inputEle.value = newVal;

            if (config.type === "checkbox") {
                inputEle.checked = newVal === "true";
            }
        })
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
