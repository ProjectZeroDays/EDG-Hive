# Reference — Design tokens (PHASE 1 §3)

```css
:root[data-theme="dark"] {
 --sidebar:#141319; --pane:#1a1920; --bg:#1c1b22;
 --card:#434042; --panel:#262530; --border:#4e4b4e;
 --accent:#ff3e3e; --accent-hover:#ff5c5c; --accent-tint:rgba(255,62,62,.12);
 --info:#6b8cff; --text:#eceaf0; --muted:#c3bfc7;
}
:root[data-theme="light"] {
 --sidebar:#ececf0; --pane:#f4f4f7; --bg:#ffffff;
 --card:#ffffff; --panel:#fafafc; --border:#dcdce4;
 --accent:#1a56db; --accent-hover:#1445b3; --accent-tint:rgba(26,86,219,.1);
 --info:#1a56db; --text:#1c1b22; --muted:#5c5a63;
}
```

Accent on: primary button fill, secondary outline, active nav rail, selected row,
focus rings, tab rules, non-zero counters, toggles, charts, destructive text. Nothing else.
Seal is a fixed asset, not recolored per theme.
