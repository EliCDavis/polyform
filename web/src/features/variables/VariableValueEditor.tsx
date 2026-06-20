import { useEffect, useState } from "react";
import type { Variable } from "@/types/schema";
import type { ThreeApp } from "@/lib/three_app";
import { setBinaryVariableValue, setVariableValue } from "@/api/variables";
import { LabeledField } from "@/components/LabeledField";
import { GizmoToggle } from "@/components/GizmoToggle";
import { VariableType } from "./variableType";
import { TransformGizmo } from "@/lib/gizmo/transform";
import { BoxGizmo } from "@/lib/gizmo/box";

function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const size = bytes / Math.pow(k, i);
  return `${size.toFixed(size < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
}

interface VariableValueEditorProps {
  variableKey: string;
  variable: Variable;
  threeApp?: ThreeApp;
}

export function VariableValueEditor({ variableKey, variable, threeApp }: VariableValueEditorProps) {
  switch (variable.type) {
    case VariableType.Bool:
      return <BoolEditor variableKey={variableKey} value={variable.value} />;
    case VariableType.String:
      return <TextEditor variableKey={variableKey} value={String(variable.value ?? "")} />;
    case VariableType.Color:
      return <ColorEditor variableKey={variableKey} value={String(variable.value ?? "")} />;
    case VariableType.Float:
      return (
        <NumberEditor variableKey={variableKey} value={variable.value} step="any" parse={parseFloat} />
      );
    case VariableType.Int:
      return (
        <NumberEditor variableKey={variableKey} value={variable.value} step="1" parse={(s) => parseInt(s, 10)} />
      );
    case VariableType.Float2:
    case VariableType.Int2:
      return (
        <Vector2Editor
          variableKey={variableKey}
          value={variable.value}
          step={variable.type === VariableType.Int2 ? "1" : "any"}
          parse={variable.type === VariableType.Int2 ? (s) => parseInt(s, 10) : parseFloat}
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
          parse={variable.type === VariableType.Int3 ? (s) => parseInt(s, 10) : parseFloat}
        />
      );
    case VariableType.AABB:
      return <AABBEditor variableKey={variableKey} value={variable.value} threeApp={threeApp} />;
    case VariableType.Float3Array:
      return <Vector3ArrayEditor variableKey={variableKey} value={variable.value} threeApp={threeApp} />;
    case VariableType.Image:
      return <ImageEditor variableKey={variableKey} />;
    case VariableType.File:
      return <FileEditor variableKey={variableKey} size={variable.value?.size ?? 0} />;
    default:
      return <span>Unsupported type: {variable.type}</span>;
  }
}

function BoolEditor({ variableKey, value }: { variableKey: string; value: boolean }) {
  const [checked, setChecked] = useState(!!value);
  return (
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => {
        setChecked(e.target.checked);
        void setVariableValue(variableKey, e.target.checked);
      }}
    />
  );
}

function TextEditor({ variableKey, value }: { variableKey: string; value: string }) {
  const [text, setText] = useState(value);
  return (
    <input
      type="text"
      value={text}
      onChange={(e) => setText(e.target.value)}
      onBlur={() => void setVariableValue(variableKey, text)}
    />
  );
}

function ColorEditor({ variableKey, value }: { variableKey: string; value: string }) {
  const [color, setColor] = useState(value);
  return (
    <div style={{ display: "flex", flexDirection: "row", gap: 16, alignItems: "center" }}>
      <input
        type="color"
        value={color}
        style={{ minHeight: 25, width: 25, maxWidth: 25, padding: 0, cursor: "pointer" }}
        onChange={(e) => {
          setColor(e.target.value);
          void setVariableValue(variableKey, e.target.value);
        }}
      />
      <span>{color}</span>
    </div>
  );
}

function NumberEditor({
  variableKey,
  value,
  step,
  parse,
}: {
  variableKey: string;
  value: number;
  step: string;
  parse: (s: string) => number;
}) {
  const [num, setNum] = useState(String(value ?? 0));
  return (
    <input
      type="number"
      step={step}
      value={num}
      onChange={(e) => setNum(e.target.value)}
      onBlur={() => void setVariableValue(variableKey, parse(num))}
    />
  );
}

function Vector2Editor({
  variableKey,
  value,
  step,
  parse,
}: {
  variableKey: string;
  value: { x: number; y: number };
  step: string;
  parse: (s: string) => number;
}) {
  const [x, setX] = useState(String(value?.x ?? 0));
  const [y, setY] = useState(String(value?.y ?? 0));

  const commit = (nx: string, ny: string) => {
    void setVariableValue(variableKey, { x: parse(nx), y: parse(ny) });
  };

  return (
    <div className="variable-inputs">
      <LabeledField label="X:">
        <input type="number" step={step} value={x} onChange={(e) => setX(e.target.value)} onBlur={() => commit(x, y)} />
      </LabeledField>
      <LabeledField label="Y:">
        <input type="number" step={step} value={y} onChange={(e) => setY(e.target.value)} onBlur={() => commit(x, y)} />
      </LabeledField>
    </div>
  );
}

function Vector3Editor({
  variableKey,
  value,
  threeApp,
  step,
  parse,
}: {
  variableKey: string;
  value: { x: number; y: number; z: number };
  threeApp?: ThreeApp;
  step: string;
  parse: (s: string) => number;
}) {
  const [x, setX] = useState(String(value?.x ?? 0));
  const [y, setY] = useState(String(value?.y ?? 0));
  const [z, setZ] = useState(String(value?.z ?? 0));
  const [gizmoOn, setGizmoOn] = useState(false);

  const commit = (nx: string, ny: string, nz: string) => {
    void setVariableValue(variableKey, { x: parse(nx), y: parse(ny), z: parse(nz) });
  };

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmo = new TransformGizmo({
      camera: threeApp.Camera,
      domElement: threeApp.Renderer.domElement,
      orbitControls: threeApp.OrbitControls,
      parent: threeApp.ViewerScene,
      scene: threeApp.Scene,
      initialPosition: { x: parse(x), y: parse(y), z: parse(z) },
    });
    gizmo.setEnabled(true);
    const sub = gizmo.position$().subscribe((pos) => {
      setX(String(pos.x));
      setY(String(pos.y));
      setZ(String(pos.z));
      void setVariableValue(variableKey, { x: pos.x, y: pos.y, z: pos.z });
    });
    return () => {
      sub.unsubscribe();
      gizmo.dispose();
    };
  }, [threeApp, gizmoOn, variableKey]);

  return (
    <div className="variable-inputs">
      <LabeledField label="X:">
        <input type="number" step={step} value={x} onChange={(e) => setX(e.target.value)} onBlur={() => commit(x, y, z)} />
      </LabeledField>
      <LabeledField label="Y:">
        <input type="number" step={step} value={y} onChange={(e) => setY(e.target.value)} onBlur={() => commit(x, y, z)} />
      </LabeledField>
      <LabeledField label="Z:">
        <input type="number" step={step} value={z} onChange={(e) => setZ(e.target.value)} onBlur={() => commit(x, y, z)} />
      </LabeledField>
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}

function AABBEditor({
  variableKey,
  value,
  threeApp,
}: {
  variableKey: string;
  value: { center: { x: number; y: number; z: number }; extents: { x: number; y: number; z: number } };
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

  const num = (label: string, v: number, onChange: (n: number) => void) => (
    <LabeledField label={label}>
      <input
        type="number"
        value={v}
        onChange={(e) => onChange(parseFloat(e.target.value))}
        onBlur={() => commit()}
      />
    </LabeledField>
  );

  return (
    <div className="variable-inputs">
      <span>center</span>
      {num("X:", center.x, (n) => setCenter((c) => ({ ...c, x: n })))}
      {num("Y:", center.y, (n) => setCenter((c) => ({ ...c, y: n })))}
      {num("Z:", center.z, (n) => setCenter((c) => ({ ...c, z: n })))}
      <span>extents</span>
      {num("X:", extents.x, (n) => setExtents((e) => ({ ...e, x: n })))}
      {num("Y:", extents.y, (n) => setExtents((e) => ({ ...e, y: n })))}
      {num("Z:", extents.z, (n) => setExtents((e) => ({ ...e, z: n })))}
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}

function Vector3ArrayEditor({
  variableKey,
  value,
  threeApp,
}: {
  variableKey: string;
  value: Array<{ x: number; y: number; z: number }>;
  threeApp?: ThreeApp;
}) {
  const [items, setItems] = useState<Array<{ x: number; y: number; z: number }>>(value ?? []);
  const [gizmoOn, setGizmoOn] = useState(false);

  useEffect(() => {
    setItems(value ?? []);
  }, [value]);

  useEffect(() => {
    if (!threeApp || !gizmoOn) return;
    const gizmos = items.map((item, i) => {
      const gizmo = new TransformGizmo({
        camera: threeApp.Camera,
        domElement: threeApp.Renderer.domElement,
        orbitControls: threeApp.OrbitControls,
        parent: threeApp.ViewerScene,
        scene: threeApp.Scene,
        initialPosition: item,
      });
      gizmo.setEnabled(true);
      const sub = gizmo.position$().subscribe((pos) => {
        setItems((prev) => {
          const next = [...prev];
          next[i] = { x: pos.x, y: pos.y, z: pos.z };
          void setVariableValue(variableKey, next);
          return next;
        });
      });
      return { gizmo, sub };
    });
    return () => {
      for (const { gizmo, sub } of gizmos) {
        sub.unsubscribe();
        gizmo.dispose();
      }
    };
  }, [threeApp, gizmoOn, items.length, variableKey]);

  const updateItem = (index: number, field: "x" | "y" | "z", val: number) => {
    setItems((prev) => {
      const next = [...prev];
      next[index] = { ...next[index], [field]: val };
      void setVariableValue(variableKey, next);
      return next;
    });
  };

  return (
    <div className="variable-inputs">
      <span>{items.length} items</span>
      {items.map((item, i) => (
        <div key={i}>
          <LabeledField label="X:">
            <input type="number" value={item.x} onChange={(e) => updateItem(i, "x", parseFloat(e.target.value))} />
          </LabeledField>
          <LabeledField label="Y:">
            <input type="number" value={item.y} onChange={(e) => updateItem(i, "y", parseFloat(e.target.value))} />
          </LabeledField>
          <LabeledField label="Z:">
            <input type="number" value={item.z} onChange={(e) => updateItem(i, "z", parseFloat(e.target.value))} />
          </LabeledField>
          <button
            type="button"
            onClick={() => {
              const next = items.filter((_, j) => j !== i);
              setItems(next);
              void setVariableValue(variableKey, next);
            }}
          >
            Delete
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() => {
          const next = [...items, { x: 0, y: 0, z: 0 }];
          setItems(next);
          void setVariableValue(variableKey, next);
        }}
      >
        Add
      </button>
      {threeApp && <GizmoToggle value={gizmoOn} onChange={setGizmoOn} />}
    </div>
  );
}

function ImageEditor({ variableKey }: { variableKey: string }) {
  const [src, setSrc] = useState(`./variable/value/${variableKey}?t=${Date.now()}`);
  return (
    <div className="variable-inputs">
      <img src={src} alt={variableKey} style={{ maxWidth: "100%" }} />
      <button type="button" onClick={() => setBinaryVariableValue(variableKey, () => setSrc(`./variable/value/${variableKey}?t=${Date.now()}`))}>
        Set Image
      </button>
    </div>
  );
}

function FileEditor({ variableKey, size }: { variableKey: string; size: number }) {
  const [displaySize, setDisplaySize] = useState(formatFileSize(size));
  return (
    <div className="variable-inputs">
      <span>{displaySize}</span>
      <button
        type="button"
        onClick={() =>
          setBinaryVariableValue(variableKey, () => {
            /* size updates on schema refresh */
          })
        }
      >
        Set File
      </button>
    </div>
  );
}
