# 19 — Chat tabs and offline queue

**Goal:** Two channels, nothing lost.

**Steps**
1. Tabs `AI` and `Users`.
2. Children see only the admin; admin sees the roster with unread counts.
3. Messages to offline users queue and deliver on return with the send time.

**Done when**
- A message sent to an offline user arrives on their next login, in order.

**Commit:** `hive(19): chat tabs and offline queue`
