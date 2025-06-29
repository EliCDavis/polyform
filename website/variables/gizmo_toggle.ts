import { Subject } from "rxjs";
import { Toggle } from "../components/toggle";
import { ElementConfig } from "../element";

export function GizmoToggle(onChange: Subject<boolean>): ElementConfig {
    return {
        style: {
            display: "flex",
            flexDirection: "row",
            alignItems: "center",
            gap: "8px"
        },
        children: [
            {
                tag: "i",
                classList: ["fa-solid", "fa-eye"],
                style: { color: "#196d6d" }
            },
            { text: "Gizmo" },
            { style: { flex: "1" } },
            Toggle({ initialValue: false, change: onChange })
        ]
    };
}