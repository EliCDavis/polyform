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

export type Vector3ArrayNodeParameter = NodeParameterBase<
  "[]vector3.Vector[float64]" | "[]vector3.Vector[float32]",
  Vec3[]
>;

export type ImageNodeParameter = NodeParameterBase<"image.Image">;

export type FileNodeParameter = NodeParameterBase<"[]uint8">;

export type AABBNodeParameter = NodeParameterBase<"geometry.AABB", AABB>;

export type NodeParameter =
  | ScalarNodeParameter
  | Vector2NodeParameter
  | Vector3NodeParameter
  | Vector3ArrayNodeParameter
  | ImageNodeParameter
  | FileNodeParameter
  | AABBNodeParameter;
