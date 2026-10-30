import { RowWidget, ValueWidget, type Widget } from '@elicdavis/node-flow';
import { useUiStore } from '../../stores/uiStore';
import { InsertRuleWidget, RemoveWidget } from './list_widgets';

function lock(widget: Widget): void {
    if (widget instanceof RowWidget) {
        widget.children().forEach(lock);
    } else if (widget instanceof ValueWidget || widget instanceof InsertRuleWidget || widget instanceof RemoveWidget) {
        widget.setDisabled(true);
    }
}

export function lockIfReadOnly<W extends Widget>(widget: W): W {
    if (!useUiStore.getState().canEdit) {
        lock(widget);
    }
    return widget;
}
