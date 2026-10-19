# What and when to back up

## Backup plans

A plan says what to back up and when: its folders, exclusions, options,
schedule and retention. A repository has one plan, named Default, and can
have more: back up your documents every hour and your photos every night
into the same repository, where data shared by both is stored once.

Each plan tags its snapshots with its own tag, so its retention policy only
removes its own snapshots, and the **Snapshots** tab shows which plan made
each one. Add, change or delete plans in the repository's **Settings**,
under **Backup plans**. Deleting a plan keeps its snapshots; no retention
policy removes them any more, so forget them by hand when you no longer
need them.

## Folders and exclusions

Each plan has a list of folders and files to back up, and exclude
patterns. Pick folders with the browser or type their paths. Patterns follow
restic's `--exclude` rules, one per line:

```text
*.tmp
/home/*/.cache
/home/*/Downloads
node_modules
```

The browser and the API only accept paths under the folders the
configuration allows, `paths.roots`. Leave it empty to allow every path, for
example on a desktop where ResticGUI runs as your user.

Change them later in the plan's settings. The snapshots stay in the same
series: ResticGUI groups them by host and plan, not by paths, so retention
keeps working when you add a folder.

## Backup options

**Backup options**, in each plan, hold the restic options of its backups:

| Option | restic |
| --- | --- |
| Tags | `--tag`, more tags for the plan's snapshots |
| Compression: `auto`, `max` or `off` | `--compression` |
| Skip files larger than, such as `2G` | `--exclude-larger-than` |
| Upload and download limits, in KiB/s | `--limit-upload`, `--limit-download` |
| Stay on one file system | `--one-file-system` |
| Skip cache folders | `--exclude-caches` |
| Skip folders that contain a file named | `--exclude-if-present` |

## Schedules

Each plan has two schedules, and the repository one, all optional:

| Schedule | Of | Runs | Typical |
| --- | --- | --- | --- |
| Back up automatically | A plan | `restic backup` | Every day at 02:30, or every hour |
| Apply the retention policy | A plan | `restic forget`, and `prune` if chosen | Every Sunday at 04:00 |
| Check the repository's integrity | The repository | `restic check` | The first day of the month |

Choose every hour, every few hours, every day, some days of the week, a day
of the month, or type a cron expression:

| Cron | Means |
| --- | --- |
| `30 2 * * *` | Every day at 02:30 |
| `0 */6 * * *` | Every 6 hours |
| `0 21 * * 1-5` | Weekdays at 21:00 |
| `0 3 1 * *` | The 1st of the month at 03:00 |

Times are in the server's time zone: set `TZ` for the service or the
container, such as `TZ=Europe/Paris`. The repository page shows the next
run. Jobs of one repository wait for each other; a schedule whose previous run
is still going skips that time. A backup missed because the server was off
runs at its next time, and a [notifier](/features/notifications) can tell
you about it.

![Schedules and retention](../../images/ui-wizard-when.png){.screenshot}

## Run now

**Back up now** starts a backup whatever the schedule: of the plan, or,
when the repository has several, of one plan or all of them. **More** holds
the other jobs: apply the retention policy, prune, check, unlock a repository
left locked by a crashed restic, and create the repository if it does not
exist yet.

Jobs run one at a time per repository, and up to `jobs.workers` at once
across repositories. A running job shows its progress live, files and
bytes, and can be stopped; restic leaves no lock behind.

## Runs

The **Runs** tab lists every job by day: what started it, how long it took,
what the backup added, and its log. A backup where some files could not be
read finishes as **succeeded** with a warning that lists them.

![The runs of a repository](../../images/ui-runs-light.png){.screenshot}
