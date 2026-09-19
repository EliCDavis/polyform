import { Fragment, useEffect, useRef, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { setVariableValue } from "@/api/variables";
import { LabeledField } from "@/components/LabeledField";
import { NumberInput } from "@/components/NumberInput";
import { GizmoToggle } from "@/components/GizmoToggle";
import { PointPathGizmo, pathInsertionPoint } from "@/lib/gizmo/point_path";
import type { Point3 } from "./vector";
import styles from "./editors.module.css";

// Points on the ground plane are stored as {x, y} and shown as (x, 0, y), the
// way the rest of the library lays 2D shapes out.
const liftPlanar = (p: Record<string, number>): Point3 => ({ x: p.x, y: 0, z: p.y });
const dropPlanar = (p: Point3): Record<string, number> => ({ x: p.x, y: p.z });

export function PointListEditor({
  variableKey,
  value,
  threeApp,
  planar,
  parse,
  step,
}: {
  variableKey: string;
  value: Array<Record<string, number>>;
  threeApp?: ThreeApp;
  planar: boolean;
  parse: (s: string) => number;
  step: string;
}) {
  const snap = parse === parseFloat ? (n: number) => n : Math.round;
  const lift = planar ? liftPlanar : (p: Record<string, number>) => p as Point3;
  const drop = planar ? dropPlanar : (p: Point3) => p as Record<string, number>;
  const fields = planar ? (["x", "y"] as const) : (["x", "y", "z"] as const);

  const [items, setItems] = useState<Array<Point3>>((value ?? []).map(lift));
  const [gizmoOn, setGizmoOn] = useState(false);
  const [focused, setFocused] = useState<number | null>(null);
  const [insertAt, setInsertAt] = useState<number | null>(null);
  const gizmoRef = useRef<PointPathGizmo | null>(null);
  const itemsRef = useRef(items);
  itemsRef.current = items;
  const pendingWrites = useRef(0);

  useEffect(() => {
    // A refetch that raced one of our own writes carries a stale array; taking
    // it would snap the points back to where they were before the edit.
    if (pendingWrites.current > 0) return;
    setItems((value ?? []).map(lift));
  }, [value]);

  const commit = (next: Array<Point3>) => {
    const snapped = next.map((p) => ({ x: snap(p.x), y: snap(p.y), z: snap(p.z) }));
    itemsRef.current = snapped;
    setItems(snapped);
    pendingWrites.current += 1;
    void setVariableValue(variableKey, snapped.map(drop)).finally(() => {
      pendingWrites.current -= 1;
    });
  };

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new PointPathGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      points: itemsRef.current,
      planar,
      closed: planar,
    });
    gizmoRef.current = gizmo;
    const sub = gizmo.changes$().subscribe(({ index, position }) => {
      commit(itemsRef.current.map((p, i) => (i === index ? position : p)));
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, gizmoOn, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setPoints(items);
  }, [items]);

  useEffect(() => {
    gizmoRef.current?.setFocus(focused);
  }, [focused]);

  useEffect(() => {
    gizmoRef.current?.setInsertPreview(insertAt);
  }, [insertAt, items]);

  // The visible field is the stored one: planar y is world z.
  const fieldOf = (p: Point3, f: "x" | "y" | "z") => (planar && f === "y" ? p.z : p[f]);
  const updateItem = (index: number, field: "x" | "y" | "z", val: number) => {
    if (Number.isNaN(val)) return;
    const key = planar && field === "y" ? "z" : field;
    commit(items.map((p, i) => (i === index ? { ...p, [key]: val } : p)));
  };

  const insertItem = (index: number) => {
    const next = [...items];
    next.splice(index, 0, pathInsertionPoint(items, index));
    setInsertAt(null);
    setFocused(index);
    commit(next);
  };

  const insertControl = (index: number, label: string) => (
    <button
      type="button"
      className={styles.insert}
      title={label}
      onClick={() => insertItem(index)}
      onPointerEnter={() => setInsertAt(index)}
      onPointerLeave={() => setInsertAt((at) => (at === index ? null : at))}
    >
      <span className={styles.insertRule} />
      <span className={styles.insertIcon}>+</span>
      <span className={styles.insertRule} />
    </button>
  );

  return (
    <div className="variable-inputs">
      <span>{items.length} items</span>
      {items.map((item, i) => (
        <Fragment key={i}>
          {insertControl(i, `Insert a point before ${i}`)}
          <div
            className={`${styles.pointRow} ${focused === i ? styles.pointRowFocused : ""}`}
            onPointerEnter={() => setFocused(i)}
            onPointerLeave={() => setFocused((f) => (f === i ? null : f))}
          >
            <span className={styles.pointIndex}>{i}</span>
            {fields.map((f) => (
              <LabeledField key={f} label={`${f.toUpperCase()}:`}>
                <NumberInput
                  step={step}
                  integer={parse !== parseFloat}
                  value={fieldOf(item, f)}
                  onCommit={(n) => updateItem(i, f, n)}
                />
              </LabeledField>
            ))}
            <button
              type="button"
              onClick={() => {
                setFocused(null);
                commit(items.filter((_, j) => j !== i));
              }}
            >
              Delete
            </button>
          </div>
        </Fragment>
      ))}
      <button
        type="button"
        onClick={() => insertItem(items.length)}
        onPointerEnter={() => setInsertAt(items.length)}
        onPointerLeave={() => setInsertAt((at) => (at === items.length ? null : at))}
      >
        Add
      </button>
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}
