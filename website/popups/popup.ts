import { Element, ElementConfig } from "../element";

const NewGraphPopupStyle = {
    // "position": "fixed",
    "justify-content": "center",
}

export function Popup(children: Array<ElementConfig>): HTMLElement {
    return Element({
        style: {
            position: "absolute",
            width: "100%",
            height: "100%",
            backgroundColor: "rgba(0,0,0,0.5)",
            top: "0",
            left: "0",
            display: "none",
            justifyContent: "center",
            alignItems: "center"
        },
        children: [{
            style: {
                backgroundColor: "#00000069",
                backdropFilter: "blur(10px)",
                padding: "24px",
                borderRadius: "24px",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
            },
            children: [
                {
                    style: NewGraphPopupStyle,
                    children: children
                }
            ]
        }]
    });
}