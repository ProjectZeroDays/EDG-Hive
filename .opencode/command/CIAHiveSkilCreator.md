---
description: Execute the next queued Hive build part, or a named part by number
agent: build
---

You are executing one part of a staged build for the Hive app.

$ARGUMENTS may be: empty (next pending part), a number, `status`, or `skip <n>`.

.opencode/hive/QUEUE.md is the single source of truth. One line per part:
07 | pending | extensibility install surface
States: pending, active, done, blocked, skipped. Only one part may be active.

Rules:
1. One part per invocation. Never start the next.
2. Read only the part file plus what it names. No out-of-scope refactors.
3. State the plan in three lines or fewer, then execute.
4. Run the suite and the build after editing. Both green or the part stays active.
5. Commit as hive(<part>): <summary>.
6. Set the part done, append a one-line note to .opencode/hive/NOTES.md with anything the next part must know.
7. If blocked, set blocked, write the reason, stop. Do not improvise around it.

Never: batch two parts in one commit, mark done on a red build, skip tests as trivial, cross the phase boundary.

Part files live at .opencode/hive/parts/NN-slug.md. If missing, stop and say so.
