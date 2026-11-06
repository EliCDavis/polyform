import { useEffect, useRef, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { setVariableValue } from "@/api/variables";
import { TransformGizmo } from "@/lib/gizmo/transform";
import { BoxGizmo } from "@/lib/gizmo/box";
import { AxisRow, PartRow, XYZ, axes } from "./fields";

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
    <AxisRow
      labels={fields.map((f) => f.toUpperCase())}
      values={fields.map((f) => current[f])}
      step={step}
      integer={parse !== parseFloat}
      onCommit={(i, n) => commit(fields[i], n)}
    />
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
  const gizmoRef = useRef<TransformGizmo | null>(null);

  useEffect(() => {
    setCurrent({ x: value?.x ?? 0, y: value?.y ?? 0, z: value?.z ?? 0 });
  }, [value]);

  const commit = (next: Point3) => {
    setCurrent(next);
    void setVariableValue(variableKey, next);
  };

  useEffect(() => {
    gizmoRef.current?.setPosition(current.x, current.y, current.z);
  }, [current]);

  useEffect(() => {
    if (!threeApp) return;
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
  }, [threeApp, variableKey]);

  return (
    <AxisRow
      labels={XYZ}
      values={axes.map((a) => current[a])}
      step={step}
      integer={parse !== parseFloat}
      onCommit={(i, n) => commit({ ...current, [axes[i]]: n })}
    />
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
  const gizmoRef = useRef<BoxGizmo | null>(null);

  useEffect(() => {
    setCenter(value?.center ?? { x: 0, y: 0, z: 0 });
    setExtents(value?.extents ?? { x: 1, y: 1, z: 1 });
  }, [value]);

  useEffect(() => {
    gizmoRef.current?.set({ center, extents });
  }, [center, extents]);

  const commit = (c = center, e = extents) => {
    void setVariableValue(variableKey, { center: c, extents: e });
  };

  useEffect(() => {
    if (!threeApp) return;
    const gizmo = new BoxGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      initial: { center, extents },
    });
    gizmo.setEnabled(true);
    gizmoRef.current = gizmo;
    const sub = gizmo.aabb$().subscribe((aabb) => {
      setCenter(aabb.center);
      setExtents(aabb.extents);
      void setVariableValue(variableKey, aabb);
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, variableKey]);

  const setCenterAxis = (i: number, n: number) => {
    const c = { ...center, [axes[i]]: n };
    setCenter(c);
    commit(c, extents);
  };
  const setExtentsAxis = (i: number, n: number) => {
    const e = { ...extents, [axes[i]]: n };
    setExtents(e);
    commit(center, e);
  };

  return (
    <div className="variable-inputs">
      <PartRow label="center">
        <AxisRow labels={XYZ} values={axes.map((a) => center[a])} onCommit={setCenterAxis} />
      </PartRow>
      <PartRow label="extents">
        <AxisRow labels={XYZ} values={axes.map((a) => extents[a])} onCommit={setExtentsAxis} />
      </PartRow>
    </div>
  );
}
