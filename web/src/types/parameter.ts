import type { AABB } from "@/lib/gizmo/box";

export interface Vec2 {
  x: number;
  y: number;
}

export interface Vec3 {
  x: number;
  y: number;
  z: number;
}

export interface Vec4 {
  x: number;
  y: number;
  z: number;
  w: number;
}

/** Common fields on parameter nodes returned by the schema API. */
export interface NodeParameterBase<TType extends string, TCurrentValue = unknown> {
  name: string;
  description?: string;
  type: TType;
  currentValue?: TCurrentValue;
}

export type ScalarParameterType =
  | "float64"
  | "float32"
  | "int"
  | "bool"
  | "string"
  | "coloring.Color";

export type ScalarNodeParameter = NodeParameterBase<
  ScalarParameterType,
  number | boolean | string
>;

export type Vector2NodeParameter = NodeParameterBase<
  "vector2.Vector[float64]" | "vector2.Vector[float32]",
  Vec2
>;

export type Vector3NodeParameter = NodeParameterBase<
  "vector3.Vector[float64]" | "vector3.Vector[float32]",
  Vec3
>;

export type Vector4NodeParameter = NodeParameterBase<
  "vector4.Vector[float64]" | "vector4.Vector[float32]",
  Vec4
>;

export type Vector3ArrayNodeParameter = NodeParameterBase<
  "[]vector3.Vector[float64]" | "[]vector3.Vector[float32]",
  Vec3[]
>;

export type NumberArrayNodeParameter = NodeParameterBase<"[]float64" | "[]int", number[]>;

export type StringArrayNodeParameter = NodeParameterBase<"[]string", string[]>;

export type Vector2ArrayNodeParameter = NodeParameterBase<
  "[]vector2.Vector[float64]" | "[]vector2.Vector[int]",
  Vec2[]
>;

export type Vector3IntArrayNodeParameter = NodeParameterBase<"[]vector3.Vector[int]", Vec3[]>;

export type ImageNodeParameter = NodeParameterBase<"image.Image">;

export type FileNodeParameter = NodeParameterBase<"[]uint8">;

export type AABBNodeParameter = NodeParameterBase<"geometry.AABB", AABB>;

export type QuaternionNodeParameter = NodeParameterBase<"quaternion.Quaternion", Vec4>;

export interface TRS {
  position: Vec3;
  rotation: Vec4;
  scale: Vec3;
}

export type TRSNodeParameter = NodeParameterBase<"trs.TRS", TRS>;

export type TRSArrayNodeParameter = NodeParameterBase<"[]trs.TRS", TRS[]>;

export type ColorArrayNodeParameter = NodeParameterBase<"[]coloring.Color", string[]>;

export interface GradientKey {
  time: number;
  value: string;
}

export type ColorGradientNodeParameter = NodeParameterBase<
  "coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color]",
  { keys: GradientKey[] }
>;

export type NodeParameter =
  | ScalarNodeParameter
  | Vector2NodeParameter
  | Vector3NodeParameter
  | Vector4NodeParameter
  | Vector3ArrayNodeParameter
  | NumberArrayNodeParameter
  | StringArrayNodeParameter
  | Vector2ArrayNodeParameter
  | Vector3IntArrayNodeParameter
  | ImageNodeParameter
  | FileNodeParameter
  | AABBNodeParameter
  | QuaternionNodeParameter
  | TRSNodeParameter
  | TRSArrayNodeParameter
  | ColorArrayNodeParameter
  | ColorGradientNodeParameter;
