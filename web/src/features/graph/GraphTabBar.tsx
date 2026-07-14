import { useGraphTabStore } from "@/stores/graphTabStore";
import styles from "./GraphTabBar.module.css";

export function GraphTabBar() {
  const tabs = useGraphTabStore((s) => s.tabs);
  const activeTabId = useGraphTabStore((s) => s.activeTabId);
  const setActiveTab = useGraphTabStore((s) => s.setActiveTab);
  const closeTab = useGraphTabStore((s) => s.closeTab);

  return (
    <div className={styles.tabBar}>
      {tabs.map((tab) => {
        const isActive = tab.id === activeTabId;
        return (
          <div
            key={tab.id}
            className={`${styles.tab} ${isActive ? styles.tabActive : ""}`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span>{tab.label}</span>
            {tab.id !== "root" && (
              <button
                type="button"
                className={styles.closeButton}
                aria-label={`Close ${tab.label}`}
                onClick={(e) => {
                  e.stopPropagation();
                  closeTab(tab.id);
                }}
              >
                ×
              </button>
            )}
          </div>
        );
      })}
    </div>
  );
}
