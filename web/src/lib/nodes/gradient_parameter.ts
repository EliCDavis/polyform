import { FlowNode, GlobalWidgetFactory, RowWidget, type Widget } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { InsertRuleWidget, RemoveWidget } from './list_widgets';
import { lockIfReadOnly } from './read_only';
import type { ColorGradientNodeParameter, GradientKey } from '../../types/parameter';

export class GradientParameterNodeController {

    flowNode: FlowNode;

    nodeManager: NodeManager;

    id: string;

    keys: Array<GradientKey>;

    widgets: Array<Widget>;

    private slots: Array<(key: GradientKey) => void> = [];

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: ColorGradientNodeParameter) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.keys = parameterData.currentValue?.keys ?? [];
        this.widgets = [];

        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        this.buildWidgets();
    }

    private setKeys(keys: Array<GradientKey>): void {
        this.keys = keys;
        if (keys.length === this.slots.length) {
            keys.forEach((key, index) => this.slots[index](key));
        } else {
            this.buildWidgets();
        }
    }

    private buildWidgets(): void {
        for (const w of this.widgets) {
            this.flowNode.removeWidget(w);
        }
        this.widgets = [];
        this.slots = [];

        if (this.keys.length === 0) {
            this.add(new InsertRuleWidget(() => this.commit([{ time: 0, value: "#000000" }, { time: 1, value: "#ffffff" }])));
            return;
        }

        const pinned = (i: number) => this.keys.length > 1 && (i === 0 || i === this.keys.length - 1);
        const replace = (index: number, patch: Partial<GradientKey>) =>
            this.commit(this.keys.map((k, i) => (i === index ? { ...k, ...patch } : k)));

        this.keys.forEach((key, index) => {
            if (index > 0) {
                this.add(new InsertRuleWidget(() => this.insertBefore(index)));
            }

            const heading = GlobalWidgetFactory.create(this.flowNode, "text", {
                value: pinned(index) ? `t = ${key.time}` : String(index),
            });
            this.add(new RowWidget([
                heading,
                new RemoveWidget(() => this.commit(this.keys.filter((_, i) => i !== index))),
            ], { grow: [1, 0] }));

            const color = GlobalWidgetFactory.create(this.flowNode, "color", {
                value: key.value,
                callback: (value) => replace(index, { value }),
            });
            if (pinned(index)) {
                this.add(color);
                this.slots.push((next) => {
                    heading.setText(`t = ${next.time}`);
                    color.setValue(next.value);
                });
                return;
            }

            const time = GlobalWidgetFactory.create(this.flowNode, "number", {
                value: key.time,
                label: "t",
                width: 72,
                min: 0,
                max: 1,
                step: 0.005,
                callback: (t) => replace(index, { time: t }),
            });
            this.add(new RowWidget([color, time], { grow: [1, 0] }));
            this.slots.push((next) => {
                color.setValue(next.value);
                time.setValue(next.time);
            });
        });
    }

    // A new stop halfway between its neighbours, in the earlier one's colour.
    private insertBefore(index: number): void {
        const left = this.keys[index - 1];
        const right = this.keys[index];
        const next = [...this.keys];
        next.splice(index, 0, { time: (left.time + right.time) / 2, value: left.value });
        this.commit(next);
    }

    private add(widget: Widget): void {
        this.widgets.push(widget);
        this.flowNode.addWidget(lockIfReadOnly(widget));
    }

    private commit(keys: Array<GradientKey>): void {
        this.setKeys([...keys].sort((a, b) => a.time - b.time));
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: { keys: this.keys },
            binary: false
        });
    }

    update(parameterData: ColorGradientNodeParameter): void {
        this.setKeys(parameterData.currentValue?.keys ?? []);
    }
}
