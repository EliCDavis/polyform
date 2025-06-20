import { BehaviorSubject, map } from "rxjs";
import { ElementConfig } from "../element";

export interface DropdownMenuItem {
    text: string;
    onclick?: (this: GlobalEventHandlers, ev: MouseEvent) => any;
}

export interface DropdownMenuConfig {
    content: Array<DropdownMenuItem>;
    buttonStyle?: Partial<CSSStyleDeclaration>
    buttonContent?: ElementConfig;
    buttonClasses?: Array<string>
}

export function DropdownMenu(config: DropdownMenuConfig): ElementConfig {
    const showDropdown$ = new BehaviorSubject<boolean>(false);
    return {
        style: { position: "relative" },
        children: [
            {
                tag: "button",
                style: config.buttonStyle,
                classList: config.buttonClasses,
                onclick: () => showDropdown$.next(!showDropdown$.value),
                children: [config.buttonContent]
            },
            {
                style: {
                    position: "absolute",
                    backgroundColor: "rgb(0 33 43)",
                    borderColor: "#003847",
                    borderWidth: "2px",
                    borderRadius: "4px",
                    borderStyle: "solid",
                    
                    flexDirection: "column",
                    display: "none",
                    right: "0",
                    zIndex: "1"
                },
                style$: showDropdown$.pipe(map(show => ({ display: show ? "flex" : "none" }))),
                children: config.content.map((item): ElementConfig => ({
                    classList: ["dropdown-item"],
                    tag: "button",
                    text: item.text,
                    onclick: (ev) => {
                        item.onclick.bind(this)(ev)
                        showDropdown$.next(!showDropdown$.value);
                    }
                }))
            }
        ]
    }
}