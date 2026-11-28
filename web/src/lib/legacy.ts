// Legacy ARTEX_* browser-storage support.
//
// The product was renamed from ARTEX to Hive and all client-side storage
// keys moved from the artex_ / artex- / artex. prefixes to hive_ equivalents.
// This shim copies values stored under the old names forward on first load,
// so existing installs keep their preferences across the upgrade. Old keys
// are removed after migration.
//
// This file is the only place legacy artex_* key names may appear. When the
// legacy support window closes, delete this file.

const LEGACY_RENAMED_KEYS: ReadonlyArray<readonly [legacy: string, current: string]> = [
  ["artex_chat_send_mode", "hive_chat_send_mode"],
  ["artex_finding_list_preferences", "hive_finding_list_preferences"],
  ["artex_task_findings_sort", "hive_task_findings_sort"],
  ["artex_task_list_filters", "hive_task_list_filters"],
  ["artex_task_list_sort", "hive_task_list_sort"],
  ["artex_token", "hive_token"],
  ["artex-update-dismissed", "hive-update-dismissed"],
  ["artex.chat.collapsed-agents", "hive.chat.collapsed-agents"],
];

export function migrateLegacyStorage(): void {
  if (typeof window === "undefined") return;
  for (const [legacyKey, currentKey] of LEGACY_RENAMED_KEYS) {
    const legacyValue = window.localStorage.getItem(legacyKey);
    if (legacyValue !== null && window.localStorage.getItem(currentKey) === null) {
      window.localStorage.setItem(currentKey, legacyValue);
    }
    window.localStorage.removeItem(legacyKey);
  }
}
