# 09 — Arsenal

**Goal:** Exploit corpus the agents can search.

**Steps**
1. Upload exploit files and scripts.
2. Import from public archives: exploitdb, nuclei templates, PayloadsAllTheThings.
3. Index by CVE and EDB id, read-only, versioned, with source and date.
4. Agent searches the arsenal before writing anything custom.

**Done when**
- Import populates and indexes; a query by CVE returns hits; empty arsenal is not an error.

**Commit:** `hive(09): arsenal upload and import`
