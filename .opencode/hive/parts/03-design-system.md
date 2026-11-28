# 03 — Design system

**Goal:** Themed tokens, no hardcoded colors.

**Steps**
1. Tokens, light: accent `#1a56db`, hover `#1445b3`, bg `#ffffff`, pane `#f4f4f7`, sidebar `#ececf0`, border `#dcdce4`, text `#1c1b22`, muted `#5c5a63`.
2. Tokens, dark: accent `#ff3e3e`, hover `#ff5c5c`, tint `rgba(255,62,62,.12)`, bg `#1c1b22`, pane `#1a1920`, sidebar `#141319`, card `#434042`, panel `#262530`, border `#4e4b4e`, text `#eceaf0`, muted `#c3bfc7`.
3. `data-theme` on root, persisted, honoring system preference.
4. Accent only on: primary fill, secondary outline, active nav rail, selected row, focus rings, tab rules, non-zero counters, toggles, chart series, destructive text. Nothing else.
5. Tooltips on every icon-only control, nav item, card and widget: 400 ms delay, card bubble, `aria-describedby`, suppressed on touch.
6. Typography: UI Inter, code IBM Plex Mono, terminal JetBrainsMono Nerd Font Mono, bundled with fallbacks.
7. Seal asset fixed across themes.

**Done when**
- Both themes render with zero hardcoded colors; accent coverage under 10% of pixels; tooltips reachable by keyboard.

**Commit:** `hive(03): design system and tokens`
