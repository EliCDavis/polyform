import type { EnumDefinition } from "./schema";

export enum GraphScopeKind {
  Root = "root",
  SubGraph = "subgraph",
}

export type GraphScope =
  | { kind: GraphScopeKind.Root }
  | { kind: GraphScopeKind.SubGraph; id: string };

export const ROOT_SCOPE: GraphScope = { kind: GraphScopeKind.Root };

export function subgraphScope(id: string): GraphScope {
  return { kind: GraphScopeKind.SubGraph, id };
}

export function scopeToApiPath(scope: GraphScope): string | null {
  if (scope.kind === GraphScopeKind.Root) return null;
  return `graph/subgraph/${scope.id}`;
}

export const ENUM_TYPE_PREFIX = "enum/";

export function isEnumType(type: string): boolean {
  return type.startsWith(ENUM_TYPE_PREFIX);
}

// Fed from /node-types as well as /schema: node types register before the first schema arrives.
const enumNames = new Map<string, string>();

export function rememberEnumNames(names: Iterable<[string, string]>): void {
  for (const [type, name] of names) {
    enumNames.set(type, name);
  }
}

export interface EnumSources {
  // The graph's own enums, keyed by id; current after every schema refresh.
  graphEnums?: { [id: string]: EnumDefinition };
  // Go enums from /node-types, keyed by variable type.
  goEnums?: { [type: string]: EnumDefinition };
}

/**
 * The enum behind a port, variable or parameter type, or undefined for any
 * other type. A port names a Go enum by its import path while the variable
 * type keys it by package-qualified name, so the lookup drops the path.
 */
export function enumOf(type: string, sources: EnumSources): EnumDefinition | undefined {
  let element = type;
  while (element.startsWith("[]")) {
    element = element.substring(2);
  }
  if (isEnumType(element)) {
    return sources.graphEnums?.[element.substring(ENUM_TYPE_PREFIX.length)];
  }
  return sources.goEnums?.[element.split("/").pop() ?? element];
}

export function enumMembersOf(type: string, sources: EnumSources): string[] | undefined {
  return enumOf(type, sources)?.members.map((m) => m.name);
}

export function formatPortTypeLabel(type: string): string {
  let prefix = "";
  let element = type;
  while (element.startsWith("[]")) {
    prefix += "[]";
    element = element.substring(2);
  }
  const enumName = enumNames.get(element);
  if (enumName !== undefined) {
    return prefix + enumName;
  }
  return type.split("github.com/EliCDavis/").join("");
}

export function isSubGraphRuntimeType(type: string): boolean {
  return type.startsWith("subgraph/");
}

/**
 * Undefined while the variable is unbound. A pattern is the variable with a
 * slice prefix per level of array: `pkg.Element`, `[]pkg.Element`.
 */
export function resolveDynamicType(
  pattern: string,
  bound: { [variable: string]: string },
): string | undefined {
  let prefix = "";
  let variable = pattern;
  while (variable.startsWith("[]")) {
    prefix += "[]";
    variable = variable.substring(2);
  }

  const concrete = bound[variable];
  return concrete === undefined ? undefined : prefix + concrete;
}

/** Drops the import path: `[]github.com/.../arrays.Element` reads as `[]arrays.Element`. */
export function shortDynamicPattern(pattern: string): string {
  let prefix = "";
  let variable = pattern;
  while (variable.startsWith("[]")) {
    prefix += "[]";
    variable = variable.substring(2);
  }

  return prefix + (variable.split("/").pop() ?? variable);
}

export function subGraphRuntimeType(id: string): string {
  return `subgraph/${id}`;
}

export const SUBGRAPH_INPUT_TYPE =
  "github.com/EliCDavis/polyform/generator/subgraph.InputNode";
export const SUBGRAPH_OUTPUT_TYPE =
  "github.com/EliCDavis/polyform/generator/subgraph.OutputNode";
