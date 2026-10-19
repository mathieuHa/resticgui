# Upgrading from 0.1

ResticGUI 0.2 is a rewrite with a new database. The first time it starts,
when `store.path` does not exist and the 0.1 database is at
`store.legacy_path` (by default `resticgui.db`, next to the configuration),
it imports it:

- repositories, with their passwords and credentials, now encrypted;
- the folders to back up and the schedules, as each repository's Default plan;
- the retention policies.

The 0.1 database is left as it was. The import is logged; if it fails,
ResticGUI does not start, and nothing is changed.

To import by hand, for example into a fresh install elsewhere:

```sh
resticgui migrate-legacy -from resticgui.db -to /var/lib/resticgui/resticgui.sqlite \
  -key-file /var/lib/resticgui/resticgui.key
```

## What changed

- The configuration file has new sections; compare yours with the
  [reference](/reference/configuration). The 0.1 API (`/api/v1`) is gone.
- The images are on `ghcr.io/mathieuha/resticgui` and listen on port 3030,
  with the database in `/opt/db`.
- Snapshots made by 0.1 are listed when you turn on **Also snapshots made
  outside ResticGUI**.

## Backup plans

An install upgraded to backup plans turns each repository's folders,
options, schedules and retention policy into one plan named Default. Its
snapshots keep their tag, so the retention policy goes on applying to the
snapshots made before the upgrade, and backups run as before.
