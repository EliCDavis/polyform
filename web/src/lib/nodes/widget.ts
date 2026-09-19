import { FlowNode } from '@elicdavis/node-flow';

// HACK: TODO: Remove once Widget is exported in node-flow
export type Widget = Parameters<FlowNode["addWidget"]>[0];
