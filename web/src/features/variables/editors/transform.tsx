import { useEffect, useRef, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { setVariableValue } from "@/api/variables";
import { LabeledField } from "@/components/LabeledField";
import { NumberInput } from "@/components/NumberInput";
import { GizmoToggle } from "@/components/GizmoToggle";
import { pathInsertionPoint } from "@/lib/gizmo/point_path";
import { TRSGizmo, type TRSGizmoMode, type TRSValue } from "@/lib/gizmo/trs";
import { TRSListGizmo } from "@/lib/gizmo/trs_list";
import { eulerToQuat, identityQuat, quatEquals, quatToEuler, type Quat } from "@/lib/rotation";
import { ListEditor, type ListHandle } from "./ListEditor";
import styles from "./editors.module.css";

const identityTRS: TRSValue = {
  position: { x: 0, y: 0, z: 0 },
  rotation: identityQuat,
  scale: { x: 1, y: 1, z: 1 },
};

// Edited as Euler degrees, stored as a quaternion. The degrees are kept as
// typed rather than re-derived from the quaternion, which would fold 370
// back to 10 and flip signs past 90 on every blur.
function QuaternionEditor({ value, onChange }: { value: Quat; onChange: (q: Quat) => void }) {
  const [euler, setEuler] = useState(quatToEuler(value ?? identityQuat));
  const lastSent = useRef<Quat>(value ?? identityQuat);

  useEffect(() => {
    const incoming = value ?? identityQuat;
    if (quatEquals(incoming, lastSent.current)) return;
    lastSent.current = incoming;
    setEuler(quatToEuler(incoming));
  }, [value]);

  const commit = (next: { x: number; y: number; z: number }) => {
    setEuler(next);
    const q = eulerToQuat(next);
    lastSent.current = q;
    onChange(q);
  };

  const field = (axis: "x" | "y" | "z") => (
    <LabeledField key={axis} label={`${axis.toUpperCase()}°:`}>
      <NumberInput value={Number(euler[axis].toFixed(4))} onCommit={(n) => commit({ ...euler, [axis]: n })} />
    </LabeledField>
  );

  return <div className="variable-inputs">{(["x", "y", "z"] as const).map(field)}</div>;
}

function TRSFields({ value, onChange }: { value: TRSValue; onChange: (next: TRSValue) => void }) {
  const vec = (label: string, key: "position" | "scale") => (
    <div className={styles.pointRow}>
      <span className={styles.pointIndex}>{label}</span>
      {(["x", "y", "z"] as const).map((axis) => (
        <LabeledField key={axis} label={`${axis.toUpperCase()}:`}>
          <NumberInput
            value={value[key][axis]}
            onCommit={(n) => onChange({ ...value, [key]: { ...value[key], [axis]: n } })}
          />
        </LabeledField>
      ))}
    </div>
  );

  return (
    <>
      {vec("position", "position")}
      <div className={styles.pointRow}>
        <span className={styles.pointIndex}>rotation</span>
        <QuaternionEditor value={value.rotation} onChange={(q) => onChange({ ...value, rotation: q })} />
      </div>
      {vec("scale", "scale")}
    </>
  );
}

function GizmoModeSelect({ mode, onChange }: { mode: TRSGizmoMode; onChange: (m: TRSGizmoMode) => void }) {
  return (
    <select value={mode} onChange={(e) => onChange(e.target.value as TRSGizmoMode)}>
      <option value="translate">Move</option>
      <option value="rotate">Rotate</option>
      <option value="scale">Scale</option>
    </select>
  );
}

export function TRSEditor({
  variableKey,
  value,
  threeApp,
}: {
  variableKey: string;
  value: TRSValue;
  threeApp?: ThreeApp;
}) {
  const [trs, setTRS] = useState<TRSValue>(value ?? identityTRS);
  const [gizmoOn, setGizmoOn] = useState(false);
  const [mode, setMode] = useState<TRSGizmoMode>("translate");
  const gizmoRef = useRef<TRSGizmo | null>(null);
  const pendingWrites = useRef(0);

  useEffect(() => {
    if (pendingWrites.current > 0) return;
    setTRS(value ?? identityTRS);
  }, [value]);

  const commit = (next: TRSValue) => {
    setTRS(next);
    pendingWrites.current += 1;
    void setVariableValue(variableKey, next).finally(() => {
      pendingWrites.current -= 1;
    });
  };

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new TRSGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      initial: trs,
      mode,
    });
    gizmo.setEnabled(true);
    gizmoRef.current = gizmo;
    const sub = gizmo.changes$().subscribe(commit);
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, gizmoOn, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setValue(trs);
  }, [trs]);

  useEffect(() => {
    gizmoRef.current?.setMode(mode);
  }, [mode]);

  return (
    <div className="variable-inputs">
      <TRSFields value={trs} onChange={commit} />
      {threeApp && (
        <>
          <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />
          {gizmoOn && <GizmoModeSelect mode={mode} onChange={setMode} />}
        </>
      )}
    </div>
  );
}

export function TRSListEditor({
  variableKey,
  value,
  threeApp,
}: {
  variableKey: string;
  value: TRSValue[];
  threeApp?: ThreeApp;
}) {
  const [gizmoOn, setGizmoOn] = useState(false);
  const [mode, setMode] = useState<TRSGizmoMode>("translate");
  const handle = useRef<ListHandle<TRSValue> | null>(null);
  const gizmoRef = useRef<TRSListGizmo | null>(null);

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new TRSListGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      items: handle.current?.items() ?? [],
      mode,
    });
    gizmoRef.current = gizmo;
    const sub = gizmo.changes$().subscribe(({ index, value: next }) => {
      const list = handle.current;
      if (list) list.commit(list.items().map((it, i) => (i === index ? next : it)));
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, gizmoOn, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setMode(mode);
  }, [mode]);

  return (
    <ListEditor<TRSValue>
      variableKey={variableKey}
      value={value}
      handle={handle}
      onItems={(items) => gizmoRef.current?.setItems(items)}
      onFocus={(i) => gizmoRef.current?.setFocus(i)}
      onInsertPreview={(i) => gizmoRef.current?.setInsertPreview(i)}
      // Copies the neighbour's rotation and scale but steps the position on,
      // so the new item does not land on top of the one it was made from.
      blank={(items, index) => ({
        ...structuredClone(items[index - 1] ?? identityTRS),
        position: pathInsertionPoint(items.map((it) => it.position), index),
      })}
      renderItem={(item, update) => <TRSFields value={item} onChange={update} />}
      footer={
        threeApp && (
          <>
            <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />
            {gizmoOn && <GizmoModeSelect mode={mode} onChange={setMode} />}
          </>
        )
      }
    />
  );
}

// A rotate-only transform gizmo parked at the origin.
export function QuaternionVariableEditor({
  variableKey,
  value,
  threeApp,
}: {
  variableKey: string;
  value: Quat;
  threeApp?: ThreeApp;
}) {
  const [gizmoOn, setGizmoOn] = useState(false);
  const [current, setCurrent] = useState<Quat>(value ?? identityQuat);
  const gizmoRef = useRef<TRSGizmo | null>(null);

  useEffect(() => {
    setCurrent(value ?? identityQuat);
  }, [value]);

  const commit = (q: Quat) => {
    setCurrent(q);
    void setVariableValue(variableKey, q);
  };

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new TRSGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      initial: { ...identityTRS, rotation: current },
      mode: "rotate",
    });
    gizmo.setEnabled(true);
    gizmoRef.current = gizmo;
    const sub = gizmo.changes$().subscribe((v) => commit(v.rotation));
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
      gizmoRef.current = null;
    };
  }, [threeApp, gizmoOn, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setValue({ ...identityTRS, rotation: current });
  }, [current]);

  return (
    <div className="variable-inputs">
      <QuaternionEditor value={current} onChange={commit} />
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}
