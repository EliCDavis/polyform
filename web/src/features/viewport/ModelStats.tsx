import { useModelStatsStore } from "@/stores/modelStatsStore";
import { useUiStore } from "@/stores/uiStore";
import styles from "./ModelStats.module.css";

export function ModelStats() {
  const stats = useModelStatsStore((s) => s.stats);
  const hideStats = useUiStore((s) => s.hideStats);

  if (hideStats || !stats || stats.vertices === 0) {
    return null;
  }

  return (
    <div className={styles.stats}>
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
    </div>
  );
}
