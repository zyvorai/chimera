# Settings

## Purpose

Show the configured listen address (read-only) and change Command Center admin username/password live — no Chimera restart required.

## When to use it

- After first login in a shared lab — rotate off default **admin** / **admin**
- Confirm which address/port Chimera binds (`CHIMERA_LISTEN`) before sharing URLs
- Same panel as **Users & Auth** — Chimera has one shared admin login, not a multi-user directory

## How to get there

- Drawer id: `#settingsDrawer`
- Nav: **Administration → Settings** or **Users & Auth** (sidebar), or top bar **gear** icon

## Operate from the console (UX)

1. Sign in (required to save credential changes).
2. Open Settings from sidebar **Settings**, **Users & Auth**, or the top bar gear — all three open this drawer.
3. Read **Listen address** at top — reflects live bind (for example `0.0.0.0:8989`); changing it needs `CHIMERA_LISTEN` in env/config and a process restart.
4. Enter **New admin username** and **New admin password** (both required).
5. Click **Save login** — toast **Login updated**; fields clear.
6. Log out via the top-right user chip and sign in with the new credentials on next session.
7. Optional: set `CHIMERA_ADMIN_USERNAME` / `CHIMERA_ADMIN_PASSWORD` at startup for immutable deploys — Settings overrides persist in memory until restart.

**Empty / fail:** **Missing fields** toast → supply both username and password. **Save failed** while logged out → login gate appears; authenticate first. Listen shows **—** → bootstrap could not read bind metadata; check config file.

**Success:** New credentials work on next login; old password rejected; listen address matches what you use in `http(s)://<host>:8989/__chimera/`.

## Related pages

- [Login](login.md)
- [Fault Studio](fault-studio.md) (separate drawer — bolt icon, not gear)
- [Admin basics](../../admin-basics.md)
- [Getting Started](../../getting-started.md)
- [Page index](../../PAGE_INDEX.md)
