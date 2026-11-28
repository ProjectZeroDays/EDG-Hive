# 10 — Remote access and mobile persistence

**Goal:** Browser remote view that survives app switching.

**Steps**
1. QR opens the browser client.
2. WebRTC screen share, view-only unless control granted, revocable mid-session.
3. Android foreground service with persistent notification and reconnect on resume.

**Done when**
- Session survives backgrounding; control revokes live.

**Commit:** `hive(10): remote access and persistence`
