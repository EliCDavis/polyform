import { Element, ElementConfig } from "../element";

const NewGraphPopupStyle = {
    // "position": "fixed",
    "justify-content": "center",
    width: "100%"
}

export enum PopupButtonType {
    Primary = "primary",
    Secondary = "secondary",
    Destructive = "destructive"
}

function buttonTypeToClass(type?: PopupButtonType): Array<string> {
    switch (type) {
        case PopupButtonType.Primary:
            return [];

        case PopupButtonType.Secondary:
            return ["secondary"];

        case PopupButtonType.Destructive:
            return ["destructive"];

        default:
            return ["secondary"];
    }
}

export interface PopupButton {
    text: string;
    click?: () => void;
    type?: PopupButtonType
}

export interface PopupConfig {
    title?: string
    buttons?: Array<PopupButton>
    content?: Array<ElementConfig>
}

export function Popup(config: PopupConfig): HTMLElement {
    let titleEle: ElementConfig = undefined;
    if (config.title) {
        titleEle = {
            tag: "h2",
            text: config.title
        }
    }

    let buttonConfig: ElementConfig = undefined;
    if (config.buttons) {
        buttonConfig = {
            style: {
                marginTop: "20px",
                flexDirection: "row",
                display: "flex",
                justifyContent: "space-between",
                width: "100%"
            },
            children: config.buttons.map(buttonConfig => ({
                tag: "button",
                text: buttonConfig.text,
                onclick: buttonConfig.click,
                classList: buttonTypeToClass(buttonConfig.type)
            }))
        }
    }

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
            classList: ["popup"],
            children: [
                titleEle,
                {
                    style: NewGraphPopupStyle,
                    children: config.content
                },
                buttonConfig
            ]
        }]
    });
}