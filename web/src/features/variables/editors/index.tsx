import type { Variable } from "@/types/schema";
import type { ThreeApp } from "@/lib/three_app";
import { VariableType } from "../variableType";
import { BoolEditor, NumberEditor, NumberListEditor, StringListEditor, TextEditor } from "./scalar";
import { ColorEditor, ColorListEditor, GradientEditor } from "./color";
import { AABBEditor, Vector2Editor, Vector3Editor, VectorEditor } from "./vector";
import { PointListEditor } from "./points";
import { QuaternionVariableEditor, TRSEditor, TRSListEditor } from "./transform";
import { FileEditor, ImageEditor } from "./file";

interface VariableValueEditorProps {
  variableKey: string;
  variable: Variable;
  threeApp?: ThreeApp;
}

const parseInteger = (s: string) => parseInt(s, 10);

export function VariableValueEditor({ variableKey, variable, threeApp }: VariableValueEditorProps) {
  switch (variable.type) {
    case VariableType.Bool:
      return <BoolEditor variableKey={variableKey} value={variable.value} />;
    case VariableType.String:
      return <TextEditor variableKey={variableKey} value={String(variable.value ?? "")} />;
    case VariableType.Color:
      return <ColorEditor variableKey={variableKey} value={String(variable.value ?? "")} />;
    case VariableType.Float:
      return <NumberEditor variableKey={variableKey} value={variable.value} step="any" parse={parseFloat} />;
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
          threeApp={threeApp}
          step={variable.type === VariableType.Int3 ? "1" : "any"}
          parse={variable.type === VariableType.Int3 ? parseInteger : parseFloat}
        />
      );
    case VariableType.Float4:
      return (
        <VectorEditor variableKey={variableKey} value={variable.value} fields={["x", "y", "z", "w"]} step="any" parse={parseFloat} />
      );
    case VariableType.AABB:
      return <AABBEditor variableKey={variableKey} value={variable.value} threeApp={threeApp} />;
    case VariableType.Quaternion:
      return <QuaternionVariableEditor variableKey={variableKey} value={variable.value} threeApp={threeApp} />;
    case VariableType.TRS:
      return <TRSEditor variableKey={variableKey} value={variable.value} threeApp={threeApp} />;
    case VariableType.TRSArray:
      return <TRSListEditor variableKey={variableKey} value={variable.value} threeApp={threeApp} />;
    case VariableType.Float3Array:
    case VariableType.Int3Array:
    case VariableType.Float2Array:
    case VariableType.Int2Array: {
      const integer = variable.type === VariableType.Int3Array || variable.type === VariableType.Int2Array;
      return (
        <PointListEditor
          variableKey={variableKey}
          value={variable.value}
          threeApp={threeApp}
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
          step={variable.type === VariableType.IntArray ? "1" : "any"}
          parse={variable.type === VariableType.IntArray ? parseInteger : parseFloat}
        />
      );
    case VariableType.StringArray:
      return <StringListEditor variableKey={variableKey} value={variable.value} />;
    case VariableType.ColorArray:
      return <ColorListEditor variableKey={variableKey} value={variable.value} />;
    case VariableType.ColorGradient:
      return <GradientEditor variableKey={variableKey} value={variable.value} />;
    case VariableType.Image:
      return <ImageEditor variableKey={variableKey} />;
    case VariableType.File:
      return <FileEditor variableKey={variableKey} size={variable.value?.size ?? 0} />;
    default:
      return <span>Unsupported type: {variable.type}</span>;
  }
}
