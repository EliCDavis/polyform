export interface NodeOutput {
    type: string;
    description?: string;
}

export interface NodeInput {
    type: string;
    isArray: boolean;
    description?: string;
}

export interface NodeDefinition {
    displayName: string;
    info: string;
    type: string;
    path: string;
    outputs?: { [key: string]: NodeOutput };
    inputs?: { [key: string]: NodeInput };
    parameter?: any;
}

export interface PortReference {
    id: string;
    port: string;
}

export interface NodeInstanceOutputPort {
    version: number;
    report?: ExecutionReport;
}

export interface NodeInstanceOutput {
    [key: string]: NodeInstanceOutputPort
}

export interface NodeInstanceAssignedInput {
    [key: string]: PortReference
}


export interface NodeInstance {
    type: string;
    name: string;
    assignedInput: NodeInstanceAssignedInput;
    output: NodeInstanceOutput;
    parameter?: any;
    variable?: any;
    metadata?: { [key: string]: any };
}

export interface GraphInstanceNodes {
    [key: string]: NodeInstance
}

export interface GraphInstance {
    producers: { [key: string]: any };
    nodes: GraphInstanceNodes;
    notes: { [key: string]: any };
    variables: VariableGroup;
    profiles: Array<string>
}

export interface Variable {
    type: string;
    // name: string;
    description: string;
    value: any;
}

export interface VariableGroup {
    subgroups: VariableGroup;
    variables: { [key: string]: Variable };
}

export interface Entry {
    metadata: { [key: string]: any };
}

export interface Manifest {
    main: string;
    entries: { [key: string]: Entry };
}

export interface CreateVariableResponse {
    nodeType: NodeDefinition
}

export interface StepTiming {
    label?: string;
    duration: number;
    steps?: StepTiming[];
}

export interface ExecutionReport {
    errors?: string[];
    logs?: string[];
    totalTime: number;
    selfTime?: number;
    steps?: StepTiming[];
}

export interface GraphExecutionReport {
    nodes: { [key: string]: NodeExecutionReport};
}

export interface NodeExecutionReport {
    output: { [key: string]: ExecutionReport};
}