import { FlowNode, GlobalWidgetFactory, type Widget } from '@elicdavis/node-flow';
import { opaque, shownAlpha, withAlpha } from '../color_alpha';

export const alphaWidgetConfig = { label: "A", min: 0, max: 1, step: 0.01, width: 64 };

export interface ColorWidgets {
    widgets: Array<Widget>;
    set: (hex: string) => void;
}

export function colorWidgets(node: FlowNode, value: string, onChange: (hex: string) => void): ColorWidgets {
    let current = value;
    const color = GlobalWidgetFactory.create(node, "color", {
        value: opaque(value),
        width: 90,
        callback: (hex) => onChange(hex + current.slice(7)),
    });
    const alpha = GlobalWidgetFactory.create(node, "number", {
        ...alphaWidgetConfig,
        value: shownAlpha(value),
        callback: (a) => onChange(withAlpha(current, a)),
    });
    return {
        widgets: [color, alpha],
        set: (hex) => {
            current = hex;
            color.setValue(opaque(hex));
            alpha.setValue(shownAlpha(hex));
        },
    };
}
