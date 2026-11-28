# Reference — Skill format (PHASE 1 §7)

```yaml
---
name: port-sweep
description: Fast TCP port sweep of an in-scope host before enumeration.
version: 1.0.0
tags: [recon, tcp]
---

## When to use
Before service fingerprinting, on any host in scope.

## Steps
1. Run nmap -Pn -p- --min-rate 2000 against the target.
2. Feed open ports to the fingerprint step.

## Notes
Back off to 500 if the target rate-limits.
```

Skill upload validates this format and rejects on mismatch.
