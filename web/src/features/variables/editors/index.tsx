import type { Variable } from "@/types/schema";
import type { ThreeApp } from "@/lib/three_app";
import { useNodeTypes, useSchema } from "@/api/hooks";
import { enumMembersOf } from "@/lib/portTypes";
import { VariableType } from "../variableType";
import { BoolEditor, NormEditor, NumberEditor, NumberListEditor, StringListEditor, TextEditor } from "./scalar";
import { ColorEditor, ColorListEditor, GradientEditor } from "./color";
import { AABBEditor, Vector2Editor, Vector3Editor, VectorEditor } from "./vector";
import { PointListEditor } from "./points";
import { QuaternionVariableEditor, TRSEditor, TRSListEditor } from "./transform";
import { FileEditor, ImageEditor } from "./file";
import { EnumEditor } from "./enum";

export interface EditorChrome {
  block: boolean;
  gizmo: boolean;
  count?: (value: any) => number;
  countNoun?: string;
}

const blockTypes = new Set<string>([
  VariableType.Float2,
  VariableType.Float3,
  VariableType.Float4,
  VariableType.Int2,
  VariableType.Int3,
  VariableType.AABB,
  VariableType.Quaternion,
  VariableType.TRS,
]);

const gizmoTypes = new Set<string>([
  VariableType.Float3,
  VariableType.Int3,
  VariableType.AABB,
  VariableType.Quaternion,
  VariableType.TRS,
  VariableType.TRSArray,
  VariableType.Float3Array,
  VariableType.Int3Array,
  VariableType.Float2Array,
  VariableType.Int2Array,
]);

const arrayLength = (value: unknown[] | undefined) => value?.length ?? 0;

const listTypes: Record<string, (value: any) => number> = {
  [VariableType.FloatArray]: arrayLength,
  [VariableType.IntArray]: arrayLength,
  [VariableType.Float2Array]: arrayLength,
  [VariableType.Float3Array]: arrayLength,
  [VariableType.Int2Array]: arrayLength,
  [VariableType.Int3Array]: arrayLength,
  [VariableType.StringArray]: arrayLength,
  [VariableType.ColorArray]: arrayLength,
  [VariableType.TRSArray]: arrayLength,
  [VariableType.ColorGradient]: (value) => value?.keys?.length ?? 0,
};

export function editorChrome(type: string): EditorChrome {
  const count = listTypes[type];
  return {
    block: blockTypes.has(type) || count !== undefined,
    gizmo: gizmoTypes.has(type),
    count,
    countNoun: type === VariableType.ColorGradient ? "stops" : "items",
  };
}

interface VariableValueEditorProps {
  variableKey: string;
  variable: Variable;
  threeApp?: ThreeApp;
  gizmo: boolean;
  open: boolean;
}

const parseInteger = (s: string) => parseInt(s, 10);

export function VariableValueEditor({ variableKey, variable, threeApp, gizmo, open }: VariableValueEditorProps) {
  const { data: graph } = useSchema();
  const { data: registered } = useNodeTypes();
  const members = enumMembersOf(variable.type, { graphEnums: graph?.enums, goEnums: registered?.enums });
  if (members) {
    return <EnumEditor variableKey={variableKey} value={String(variable.value ?? "")} values={members} />;
  }
  const gizmoApp = gizmo ? threeApp : undefined;
  switch (variable.type) {
    case VariableType.Bool:
      return <BoolEditor variableKey={variableKey} value={variable.value} />;
    case VariableType.String:
      return <TextEditor variableKey={variableKey} value={String(variable.value ?? "")} />;
    case VariableType.Color:
      return <ColorEditor variableKey={variableKey} value={String(variable.value ?? "")} />;
    case VariableType.Float:
      return <NumberEditor variableKey={variableKey} value={variable.value} step="any" parse={parseFloat} />;
    case VariableType.UNorm:
    case VariableType.SNorm:
      return <NormEditor variableKey={variableKey} value={variable.value} min={variable.type === VariableType.SNorm ? -1 : 0} />;
    case VariableType.Int:
      return <NumberEditor variableKey={variableKey} value={variable.value} step="1" parse={parseInteger} />;
    case VariableType.Float2:
    case VariableType.Int2:
      return (
        <Vector2Editor
          variableKey={variableKey}
          value={variable.value}
          step={variable.type === VariableType.Int2 ? "1" : "any"}
          parse={variable.type === VariableType.Int2 ? parseInteger : parseFloat}
        />
      );
    case VariableType.Float3:
    case VariableType.Int3:
      return (
        <Vector3Editor
          variableKey={variableKey}
          value={variable.value}
          threeApp={gizmoApp}
          step={variable.type === VariableType.Int3 ? "1" : "any"}
          parse={variable.type === VariableType.Int3 ? parseInteger : parseFloat}
        />
      );
    case VariableType.Float4:
      return (
        <VectorEditor variableKey={variableKey} value={variable.value} fields={["x", "y", "z", "w"]} step="any" parse={parseFloat} />
      );
    case VariableType.AABB:
      return <AABBEditor variableKey={variableKey} value={variable.value} threeApp={gizmoApp} />;
    case VariableType.Quaternion:
      return <QuaternionVariableEditor variableKey={variableKey} value={variable.value} threeApp={gizmoApp} />;
    case VariableType.TRS:
      return <TRSEditor variableKey={variableKey} value={variable.value} threeApp={gizmoApp} />;
    case VariableType.TRSArray:
      return <TRSListEditor variableKey={variableKey} value={variable.value} threeApp={gizmoApp} open={open} />;
    case VariableType.Float3Array:
    case VariableType.Int3Array:
    case VariableType.Float2Array:
    case VariableType.Int2Array: {
      const integer = variable.type === VariableType.Int3Array || variable.type === VariableType.Int2Array;
      return (
        <PointListEditor
          variableKey={variableKey}
          value={variable.value}
          threeApp={gizmoApp}
          open={open}
          planar={variable.type === VariableType.Float2Array || variable.type === VariableType.Int2Array}
          step={integer ? "1" : "any"}
          parse={integer ? parseInteger : parseFloat}
        />
      );
    }
    case VariableType.FloatArray:
    case VariableType.IntArray:
      return (
        <NumberListEditor
          variableKey={variableKey}
          value={variable.value}
          open={open}
          step={variable.type === VariableType.IntArray ? "1" : "any"}
          parse={variable.type === VariableType.IntArray ? parseInteger : parseFloat}
        />
      );
    case VariableType.StringArray:
      return <StringListEditor variableKey={variableKey} value={variable.value} open={open} />;
    case VariableType.ColorArray:
      return <ColorListEditor variableKey={variableKey} value={variable.value} open={open} />;
    case VariableType.ColorGradient:
      return <GradientEditor variableKey={variableKey} value={variable.value} open={open} />;
    case VariableType.Image:
      return <ImageEditor variableKey={variableKey} />;
    case VariableType.File:
      return <FileEditor variableKey={variableKey} size={variable.value?.size ?? 0} />;
    default:
      return <span>Unsupported type: {variable.type}</span>;
  }
}
