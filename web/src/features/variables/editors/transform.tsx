import { useEffect, useRef, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { setVariableValue } from "@/api/variables";
import { pathInsertionPoint } from "@/lib/gizmo/point_path";
import { TRSGizmo, type TRSGizmoMode, type TRSValue } from "@/lib/gizmo/trs";
import { TRSListGizmo } from "@/lib/gizmo/trs_list";
import { eulerToQuat, identityQuat, quatEquals, quatToEuler, type Quat } from "@/lib/rotation";
import { ListEditor, type ListHandle } from "./ListEditor";
import { AxisRow, PartRow, XYZ, axes } from "./fields";
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

  return (
    <AxisRow
      labels={["X°", "Y°", "Z°"]}
      values={axes.map((a) => Number(euler[a].toFixed(4)))}
      onCommit={(i, n) => commit({ ...euler, [axes[i]]: n })}
    />
  );
}

function TRSFields({ value, onChange }: { value: TRSValue; onChange: (next: TRSValue) => void }) {
  const vec = (label: string, key: "position" | "scale") => (
    <PartRow label={label}>
      <AxisRow
        labels={XYZ}
        values={axes.map((a) => value[key][a])}
        onCommit={(i, n) => onChange({ ...value, [key]: { ...value[key], [axes[i]]: n } })}
      />
    </PartRow>
  );

  return (
    <div className="variable-inputs">
      {vec("pos", "position")}
      <PartRow label="rot">
        <QuaternionEditor value={value.rotation} onChange={(q) => onChange({ ...value, rotation: q })} />
      </PartRow>
      {vec("scale", "scale")}
    </div>
  );
}

const round = (n: number) => Number(n.toFixed(2));

function TRSItem({ value, onChange }: { value: TRSValue; onChange: (next: TRSValue) => void }) {
  const [open, setOpen] = useState(false);
  const { x, y, z } = value.position;
  return (
    <div className="variable-inputs">
      <button type="button" className={styles.itemSummary} aria-expanded={open} onClick={() => setOpen(!open)}>
        {open ? "▾" : "▸"} pos {round(x)}, {round(y)}, {round(z)}
      </button>
      {open && <TRSFields value={value} onChange={onChange} />}
    </div>
  );
}

function GizmoModeSelect({ mode, onChange }: { mode: TRSGizmoMode; onChange: (m: TRSGizmoMode) => void }) {
  return (
    <PartRow label="gizmo">
      <select value={mode} onChange={(e) => onChange(e.target.value as TRSGizmoMode)}>
        <option value="translate">Move</option>
        <option value="rotate">Rotate</option>
        <option value="scale">Scale</option>
      </select>
    </PartRow>
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
    if (!threeApp) return;
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
  }, [threeApp, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setValue(trs);
  }, [trs]);

  useEffect(() => {
    gizmoRef.current?.setMode(mode);
  }, [mode]);

  return (
    <div className="variable-inputs">
      <TRSFields value={trs} onChange={commit} />
      {threeApp && <GizmoModeSelect mode={mode} onChange={setMode} />}
    </div>
  );
}

export function TRSListEditor({
  variableKey,
  value,
  threeApp,
  open,
}: {
  variableKey: string;
  value: TRSValue[];
  threeApp?: ThreeApp;
  open: boolean;
}) {
  const [mode, setMode] = useState<TRSGizmoMode>("translate");
  const handle = useRef<ListHandle<TRSValue> | null>(null);
  const gizmoRef = useRef<TRSListGizmo | null>(null);

  useEffect(() => {
    if (!threeApp) return;
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
  }, [threeApp, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setMode(mode);
  }, [mode]);

  return (
    <ListEditor<TRSValue>
      variableKey={variableKey}
      value={value}
      open={open}
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
      renderItem={(item, update) => <TRSItem value={item} onChange={update} />}
      footer={threeApp && <GizmoModeSelect mode={mode} onChange={setMode} />}
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
    if (!threeApp) return;
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
  }, [threeApp, variableKey]);

  useEffect(() => {
    gizmoRef.current?.setValue({ ...identityTRS, rotation: current });
  }, [current]);

  return <QuaternionEditor value={current} onChange={commit} />;
}
