import { NodeParameter } from "../types/parameter";

export interface NodeOutput {
  type: string;
  description?: string;
  dynamic?: boolean;
}

export interface NodeInput {
  type: string;
  isArray: boolean;
  description?: string;
  dynamic?: boolean;

  // Set when the port takes more than one type, because `type` is only the
  // one it happens to be showing.
  acceptedTypes?: Array<string>;
}

export interface GraphHistory {
  undo: Array<string>;
  redo: Array<string>;
  applied?: string;
}

export interface RegisteredTypes {
  nodeTypes: Array<NodeDefinition>;
  serializableOutputTypes: Array<string>;
  portTypes: Array<string>;
}

export interface NodeDefinition {
  displayName: string;
  info: string;
  type: string;
  path: string;
  keywords?: Array<string>;
  outputs?: { [key: string]: NodeOutput };
  inputs?: { [key: string]: NodeInput };
  parameter?: NodeParameter;
}

export interface PortReference {
  id: string;
  port: string;
}

export interface NodeInstanceOutputPort {
  version: number;

  // Sent only when it differs from the node type's declaration: a lifted
  // output becomes an array once one is wired in.
  type?: string;
}

export interface NodeInstanceOutput {
  [key: string]: NodeInstanceOutputPort;
}

export interface NodeInstanceAssignedInput {
  [key: string]: PortReference;
}

export enum BoundaryType {
  Input = "input",
  Output = "output",
}

export interface SubGraphInputBoundary {
  portName: string;
  portType: string;
}

export interface SubGraphOutputBoundary {
  portName: string;
  portType: string;
}

export function subGraphBoundaryKind(
  node: NodeInstance,
): BoundaryType | undefined {
  if (node.subGraphInputBoundary) return BoundaryType.Input;
  if (node.subGraphOutputBoundary) return BoundaryType.Output;
  return undefined;
}

export function subGraphBoundaryInfo(
  node: NodeInstance,
): SubGraphInputBoundary | SubGraphOutputBoundary | undefined {
  return node.subGraphInputBoundary ?? node.subGraphOutputBoundary;
}

export interface NodeInstance {
  type: string;
  name: string;
  assignedInput: NodeInstanceAssignedInput;
  output: NodeInstanceOutput;
  parameter?: NodeParameter;
  variable?: any;
  metadata?: { [key: string]: any };
  subGraphInputBoundary?: SubGraphInputBoundary;
  subGraphOutputBoundary?: SubGraphOutputBoundary;
  subGraphId?: string;

  // For a node with dynamic ports, what each of its type variables settled
  // on, keyed by the variable the ports were declared with.
  dynamicTypes?: { [variable: string]: string };

  variablePath?: string;
}

export interface RuntimeSubGraphDefinition {
  name: string;
  description?: string;
  nodes: GraphInstanceNodes;
  notes?: { [key: string]: any };
  variables?: VariableGroup;
}

export interface GraphInstance {
  producers: { [key: string]: any };
  nodes: GraphInstanceNodes;
  notes: { [key: string]: any };
  variables: VariableGroup;
  profiles: Array<string>;
  subGraphs?: { [key: string]: RuntimeSubGraphDefinition };
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
  nodeType: NodeDefinition;
}

export interface CreateSubGraphResponse {
  nodeType: NodeDefinition;
}

export interface ImportedSubGraphEntry {
  id: string;
  name: string;
  originalId?: string;
  nodeType: NodeDefinition;
}

export interface ImportSubGraphsResult {
  imported: ImportedSubGraphEntry[];
}

export interface ConvertSelectionToSubGraphRequest {
  scope: string;
  nodeIds: string[];
  name: string;
  description?: string;
}

export interface ConvertSelectionToSubGraphResponse {
  subGraphId: string;
  name: string;
  runtimeNodeId: string;
  nodeType: NodeDefinition;
}

export interface GraphInstanceNodes {
  [key: string]: NodeInstance;
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
  nodes: { [key: string]: NodeExecutionReport };
}

export interface NodeExecutionReport {
  output: { [key: string]: ExecutionReport };
}
