import { useEffect, useRef, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { setVariableValue } from "@/api/variables";
import { LabeledField } from "@/components/LabeledField";
import { NumberInput } from "@/components/NumberInput";
import { GizmoToggle } from "@/components/GizmoToggle";
import { TransformGizmo } from "@/lib/gizmo/transform";
import { BoxGizmo } from "@/lib/gizmo/box";

export type Point3 = { x: number; y: number; z: number };

export function VectorEditor({
  variableKey,
  value,
  fields,
  step,
  parse,
}: {
  variableKey: string;
  value: Record<string, number>;
  fields: string[];
  step: string;
  parse: (s: string) => number;
}) {
  const [current, setCurrent] = useState<Record<string, number>>(
    Object.fromEntries(fields.map((f) => [f, value?.[f] ?? 0])),
  );

  useEffect(() => {
    setCurrent(Object.fromEntries(fields.map((f) => [f, value?.[f] ?? 0])));
  }, [value]);

  const commit = (f: string, n: number) => {
    const next = { ...current, [f]: n };
    setCurrent(next);
    void setVariableValue(variableKey, next);
  };

  return (
    <div className="variable-inputs">
      {fields.map((f) => (
        <LabeledField key={f} label={`${f.toUpperCase()}:`}>
          <NumberInput step={step} integer={parse !== parseFloat} value={current[f]} onCommit={(n) => commit(f, n)} />
        </LabeledField>
      ))}
    </div>
  );
}

export function Vector2Editor(props: {
  variableKey: string;
  value: { x: number; y: number };
  step: string;
  parse: (s: string) => number;
}) {
  return <VectorEditor {...props} fields={["x", "y"]} />;
}

export function Vector3Editor({
  variableKey,
  value,
  threeApp,
  step,
  parse,
}: {
  variableKey: string;
  value: Point3;
  threeApp?: ThreeApp;
  step: string;
  parse: (s: string) => number;
}) {
  const [current, setCurrent] = useState<Point3>({ x: value?.x ?? 0, y: value?.y ?? 0, z: value?.z ?? 0 });
  const [gizmoOn, setGizmoOn] = useState(false);
  const gizmoRef = useRef<TransformGizmo | null>(null);

  useEffect(() => {
    setCurrent({ x: value?.x ?? 0, y: value?.y ?? 0, z: value?.z ?? 0 });
  }, [value]);

  const commit = (next: Point3) => {
    setCurrent(next);
    gizmoRef.current?.setPosition(next.x, next.y, next.z);
    void setVariableValue(variableKey, next);
  };

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new TransformGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      initialPosition: current,
    });
    gizmo.setEnabled(true);
    gizmoRef.current = gizmo;
    const sub = gizmo.position$().subscribe((pos) => {
      const next = { x: pos.x, y: pos.y, z: pos.z };
      setCurrent(next);
      void setVariableValue(variableKey, next);
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, gizmoOn, variableKey]);

  return (
    <div className="variable-inputs">
      {(["x", "y", "z"] as const).map((axis) => (
        <LabeledField key={axis} label={`${axis.toUpperCase()}:`}>
          <NumberInput
            step={step}
            integer={parse !== parseFloat}
            value={current[axis]}
            onCommit={(n) => commit({ ...current, [axis]: n })}
          />
        </LabeledField>
      ))}
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}

export function AABBEditor({
  variableKey,
  value,
  threeApp,
}: {
  variableKey: string;
  value: { center: Point3; extents: Point3 };
  threeApp?: ThreeApp;
}) {
  const [center, setCenter] = useState(value?.center ?? { x: 0, y: 0, z: 0 });
  const [extents, setExtents] = useState(value?.extents ?? { x: 1, y: 1, z: 1 });
  const [gizmoOn, setGizmoOn] = useState(false);

  const commit = (c = center, e = extents) => {
    void setVariableValue(variableKey, { center: c, extents: e });
  };

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new BoxGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      initial: { center, extents },
    });
    gizmo.setEnabled(true);
    const sub = gizmo.aabb$().subscribe((aabb) => {
      setCenter(aabb.center);
      setExtents(aabb.extents);
      void setVariableValue(variableKey, aabb);
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
    };
  }, [threeApp, gizmoOn, variableKey]);

  const num = (label: string, v: number, onCommit: (n: number) => void) => (
    <LabeledField label={label}>
      <NumberInput value={v} onCommit={onCommit} />
    </LabeledField>
  );
  const setCenterAxis = (axis: "x" | "y" | "z", n: number) => {
    const c = { ...center, [axis]: n };
    setCenter(c);
    commit(c, extents);
  };
  const setExtentsAxis = (axis: "x" | "y" | "z", n: number) => {
    const e = { ...extents, [axis]: n };
    setExtents(e);
    commit(center, e);
  };

  return (
    <div className="variable-inputs">
      <span>center</span>
      {num("X:", center.x, (n) => setCenterAxis("x", n))}
      {num("Y:", center.y, (n) => setCenterAxis("y", n))}
      {num("Z:", center.z, (n) => setCenterAxis("z", n))}
      <span>extents</span>
      {num("X:", extents.x, (n) => setExtentsAxis("x", n))}
      {num("Y:", extents.y, (n) => setExtentsAxis("y", n))}
      {num("Z:", extents.z, (n) => setExtentsAxis("z", n))}
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}
