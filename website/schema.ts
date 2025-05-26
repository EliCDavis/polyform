export interface NodeOutput {
    type: string;
    description?: string;
}

export interface NodeInput {
    type: string;
    isArray: boolean;
    description?: string;
}

export interface NodeType {
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
}

export interface Variable {
    type: string;
    name: string;
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