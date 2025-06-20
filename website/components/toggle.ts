import { BehaviorSubject, map, Subject } from "rxjs";
import { ElementConfig } from "../element";

export interface ToggleConfig {
    initialValue?: boolean;
    change: Subject<boolean>;
}

export function Toggle(config?: ToggleConfig): ElementConfig {
    let currentValue = config?.initialValue ? config?.initialValue : false;
    const display = new BehaviorSubject<boolean>(currentValue);
    return {
        tag: "button",
        classList: ["toggle"],
        style$: display.pipe(map((turnedOn) => ({
            flexDirection: turnedOn ? "row-reverse" : "row",
            backgroundColor: turnedOn ? "#196d6d" : "#0a2e3d"
        }))),
        children: [
            {
                classList: ["toggle-slider"]
            }
        ],
        onclick: () => {
            currentValue = !currentValue;
            display.next(currentValue);
            if (config?.change) {
                config.change.next(currentValue);
            }
        }
    }
}