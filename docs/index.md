---
layout: home
hero:
  name: ResticGUI
  text: Your restic backups, in a browser
  tagline: Schedule backups, keep the snapshots you need, browse and restore files, and hear about it when something goes wrong. One small binary with restic underneath.
  image:
    src: /logo.svg
    alt: ResticGUI
  actions:
    - theme: brand
      text: Quick start
      link: /guide/quick-start
    - theme: alt
      text: Features
      link: /features/storage
    - theme: alt
      text: GitHub
      link: https://github.com/mathieuHa/resticgui
features:
  - icon: 🗄️
    title: Any restic storage
    details: Local disks, S3 and compatible, Backblaze B2, a REST server, SFTP, rclone, Azure and Google Cloud, with credentials stored encrypted.
    link: /features/storage
  - icon: ⏰
    title: Schedules and retention
    details: Daily, hourly or any cron schedule, keep-daily/weekly/monthly rules with a preview of what will be removed, and integrity checks.
    link: /features/backups
  - icon: 📂
    title: Browse and restore
    details: Walk through any snapshot, find files by name, download a file or a folder as zip, or restore in place with live progress.
    link: /features/restore
  - icon: 🔔
    title: Alerts that reach you
    details: Failed or missed backups on ntfy, Gotify, Discord, Slack, Telegram or email, plus Healthchecks pings.
    link: /features/notifications
  - icon: 🪝
    title: Hooks
    details: Dump a database or stop a service before a backup, call a webhook after it.
    link: /features/hooks
  - icon: 📈
    title: Metrics and Grafana
    details: A Prometheus endpoint and a ready-made Grafana dashboard for your backups.
    link: /features/metrics
---

<div style="max-width: 1152px; margin: 48px auto 0; padding: 0 24px;">

![The ResticGUI dashboard](../images/ui-dashboard-light.png){.screenshot}

</div>
