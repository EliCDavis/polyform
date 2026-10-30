import { useEffect, useRef, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { LabeledField } from "@/components/LabeledField";
import { NumberInput } from "@/components/NumberInput";
import { GizmoToggle } from "@/components/GizmoToggle";
import { PointPathGizmo, pathInsertionPoint } from "@/lib/gizmo/point_path";
import type { Point3 } from "./vector";
import { ListEditor, type ListHandle } from "./ListEditor";

type Point = Record<string, number>;

// Points on the ground plane are stored as {x, y} and shown as (x, 0, y), the
// way the rest of the library lays 2D shapes out.
const liftPlanar = (p: Point): Point3 => ({ x: p.x, y: 0, z: p.y });
const dropPlanar = (p: Point3): Point => ({ x: p.x, y: p.z });

export function PointListEditor({
  variableKey,
  value,
  threeApp,
  planar,
  parse,
  step,
}: {
  variableKey: string;
  value: Array<Point>;
  threeApp?: ThreeApp;
  planar: boolean;
  parse: (s: string) => number;
  step: string;
}) {
  const snap = parse === parseFloat ? (n: number) => n : Math.round;
  const lift = planar ? liftPlanar : (p: Point) => p as Point3;
  const drop = planar ? dropPlanar : (p: Point3) => p as Point;
  const fields = planar ? (["x", "y"] as const) : (["x", "y", "z"] as const);

  const [gizmoOn, setGizmoOn] = useState(false);
  const handle = useRef<ListHandle<Point> | null>(null);
  const gizmoRef = useRef<PointPathGizmo | null>(null);

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new PointPathGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      points: (handle.current?.items() ?? []).map(lift),
      planar,
      closed: planar,
    });
    gizmoRef.current = gizmo;
    const sub = gizmo.changes$().subscribe(({ index, position }) => {
      const list = handle.current;
      if (!list) return;
      const moved = drop({ x: snap(position.x), y: snap(position.y), z: snap(position.z) });
      list.commit(list.items().map((p, i) => (i === index ? moved : p)));
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, gizmoOn, variableKey]);

  return (
    <ListEditor<Point>
      variableKey={variableKey}
      value={value}
      handle={handle}
      onItems={(items) => gizmoRef.current?.setPoints(items.map(lift))}
      onFocus={(i) => gizmoRef.current?.setFocus(i)}
      onInsertPreview={(i) => gizmoRef.current?.setInsertPreview(i)}
      blank={(items, index) => drop(pathInsertionPoint(items.map(lift), index))}
      renderItem={(item, update) =>
        fields.map((f) => (
          <LabeledField key={f} label={`${f.toUpperCase()}:`}>
            <NumberInput
              step={step}
              integer={parse !== parseFloat}
              value={item[f]}
              onCommit={(n) => update({ ...item, [f]: snap(n) })}
            />
          </LabeledField>
        ))
      }
      footer={threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    />
  );
}
