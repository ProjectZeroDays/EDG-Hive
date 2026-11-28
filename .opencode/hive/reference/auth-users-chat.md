# Reference — Auth, registration, user management (PHASE 1 §9/§10/§11)

## Account model
Exactly one admin — the first account, unique-constrained. Admin never minted
by invite; transfer CLI-only (`hive admin transfer`) behind current admin's passkey.

## Registration modes
`REGISTRATION_MODE`: `open` (localhost default) / `invite` (hosted default,
auto-flip on non-loopback bind) / `closed`. Mode shown on login screen.

## Invite tokens
- `hive_inv_` + 32 random bytes, base32. SHA-256 at rest. Plaintext shown once.
- Fields: label, max_uses (default 1), expires_at (default 7d), assigned role,
  revoke flag, created-by, timestamps.
- Redemption `/register?token=...`: validate + decrement in ONE transaction.
- Timing-safe compare, generic "invalid or expired" error, 5 attempts/min/IP, audit.
- Warning when REGISTRATION_MODE=open while unused tokens exist.

## Passkeys
Token gates creation only; passkey/FIDO2 enrollment follows immediately,
mandatory on hosted. Admin can reset a child's passkey.

## User states
active, suspended, disabled, pending. Disable kills live sessions immediately.
Actions: disable, suspend, reinstate, reset passkey, force re-enroll, delete,
extend invite.

## Usage
Per-child requests, tokens in/out, model, cost, daily rollup, date filter,
filtered logs. Flag over configurable daily spend threshold.

## Key source
`byok` (default, encrypted at rest) or `hosted` (admin grant, monthly cap).
Neither → child cannot send a prompt. Tag every request with its source.

## Teaming
Per-child toggle, default off. Joined admin sees screen + task state, messages
in-band, takes over. Persistent "admin is watching" indicator; child sees who
joined and when. Server enforces 4 concurrent joins max.

## Chat
Two tabs: `AI`, `Users`. Children see only admin; admin sees roster + unread.
Offline messages queue, deliver on return with send time. Unread badges.
Admin DMs pinned above AI in child's view when unread.

## Telegram
Per-user bot link, optional. Mirrors DMs both directions. `!ai` prefix routes
inbound to the agent; replies return to Telegram. One-click unlink.

## Remote
QR opens browser client, can start WebRTC screen share. Android foreground
service, reconnect on resume. View-only unless control granted; revocable
mid-session.

## Console/logs
Level filter, search, follow, export, clear, admin scoping. Never log secrets.

## Docs
README, TROUBLESHOOTING, HELP, CONTRIBUTING, TUTORIALS, API_DOCS, CHANGELOG, LICENSE.
