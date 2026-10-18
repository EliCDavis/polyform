import { useState } from "react";
import { Box, ChevronDown, ChevronRight, CircleDot, Eye, EyeOff, Layers } from "lucide-react";
import { useModelStatsStore } from "@/stores/modelStatsStore";
import { useUiStore } from "@/stores/uiStore";
import { useEditorOptional } from "@/features/editor/EditorContext";
import type { HierarchyEntry, HierarchyKind } from "@/lib/ProducerView/model_hierarchy";
import styles from "./ModelStats.module.css";

const KIND_ICONS: Record<HierarchyKind, typeof Box> = {
  mesh: Box,
  points: CircleDot,
  group: Layers,
};

interface RowProps {
  entry: HierarchyEntry;
  depth: number;
  ancestorHidden: boolean;
  expanded: Set<string>;
  onToggleExpanded: (path: string) => void;
  onToggleVisible: (entry: HierarchyEntry) => void;
  onHover: (path: string | null) => void;
}

function HierarchyRow({ entry, depth, ancestorHidden, expanded, onToggleExpanded, onToggleVisible, onHover }: RowProps) {
  const open = expanded.has(entry.path);
  const hasChildren = entry.children.length > 0;
  const KindIcon = KIND_ICONS[entry.kind];
  const VisibleIcon = entry.visible ? Eye : EyeOff;

  return (
    <>
      <div
        className={`${styles.row} ${entry.visible && !ancestorHidden ? "" : styles.hidden}`}
        style={{ paddingLeft: 4 + depth * 14 }}
        onMouseEnter={() => onHover(entry.path)}
      >
        <button
          className={styles.iconButton}
          style={{ visibility: hasChildren ? "visible" : "hidden" }}
          onClick={() => onToggleExpanded(entry.path)}
          aria-label={open ? "Collapse" : "Expand"}
        >
          {open ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        </button>
        <button
          className={styles.iconButton}
          onClick={() => onToggleVisible(entry)}
          aria-label={entry.visible ? "Hide" : "Show"}
          title={entry.visible ? "Hide" : "Show"}
        >
          <VisibleIcon size={12} />
        </button>
        <KindIcon size={12} className={styles.kind} />
        <span className={styles.name} title={entry.name}>
          {entry.name}
          {entry.instances > 1 && <span className={styles.unit}>×{entry.instances.toLocaleString()}</span>}
        </span>
        {entry.vertices > 0 && (
          <span className={styles.count}>
            {(entry.triangles > 0 ? entry.triangles : entry.vertices).toLocaleString()}
            <span className={styles.unit}>{entry.triangles > 0 ? "tris" : "verts"}</span>
          </span>
        )}
      </div>
      {open &&
        entry.children.map((child) => (
          <HierarchyRow
            key={child.path}
            entry={child}
            depth={depth + 1}
            ancestorHidden={ancestorHidden || !entry.visible}
            expanded={expanded}
            onToggleExpanded={onToggleExpanded}
            onToggleVisible={onToggleVisible}
            onHover={onHover}
          />
        ))}
    </>
  );
}

export function ModelStats() {
  const stats = useModelStatsStore((s) => s.stats);
  const hierarchy = useModelStatsStore((s) => s.hierarchy);
  const hideStats = useUiStore((s) => s.hideStats);
  const editor = useEditorOptional();
  const [open, setOpen] = useState(false);
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set());

  if (hideStats || !stats) {
    return null;
  }

  const toggleExpanded = (path: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  const toggleVisible = (entry: HierarchyEntry) => {
    editor?.producerViewManager.SetPartVisible(entry.path, !entry.visible);
  };

  const hover = (path: string | null) => {
    editor?.producerViewManager.HighlightPart(path);
  };

  const toggleOpen = () => {
    if (open) {
      hover(null);
    }
    setOpen(!open);
  };

  return (
    <div className={styles.root}>
      <button className={styles.stats} onClick={toggleOpen} aria-expanded={open} title="Model hierarchy">
        {stats.triangles > 0 && (
          <span>
            <span className={styles.value}>{stats.triangles.toLocaleString()}</span>
            <span className={styles.unit}>tris</span>
          </span>
        )}
        <span>
          <span className={styles.value}>{stats.vertices.toLocaleString()}</span>
          <span className={styles.unit}>verts</span>
        </span>
        <span>
          <span className={styles.value}>{stats.draws.toLocaleString()}</span>
          <span className={styles.unit}>draws</span>
        </span>
        <ChevronDown size={12} className={`${styles.arrow} ${open ? styles.arrowOpen : ""}`} />
      </button>
      {open && (
        <div className={styles.panel} onMouseLeave={() => hover(null)}>
          {hierarchy.length === 0 && <div className={styles.empty}>No parts</div>}
          {hierarchy.map((entry) => (
            <HierarchyRow
              key={entry.path}
              entry={entry}
              depth={0}
              ancestorHidden={false}
              expanded={expanded}
              onToggleExpanded={toggleExpanded}
              onToggleVisible={toggleVisible}
              onHover={hover}
            />
          ))}
        </div>
      )}
    </div>
  );
}
