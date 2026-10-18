import { defineConfig } from 'vitepress';

const repo = 'https://github.com/mathieuHa/resticgui';

export default defineConfig({
  title: 'ResticGUI',
  description: 'A web interface for restic backups: schedules, retention, browse and restore, alerts and metrics.',
  lang: 'en',
  cleanUrls: true,
  lastUpdated: true,
  sitemap: { hostname: 'https://docs.resticgui.com' },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['meta', { property: 'og:title', content: 'ResticGUI' }],
    ['meta', { property: 'og:description', content: 'A web interface for restic backups.' }],
    ['meta', { property: 'og:image', content: 'https://raw.githubusercontent.com/mathieuHa/resticgui/main/images/ui-dashboard-light.png' }],
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Quick start', link: '/guide/quick-start' },
      { text: 'Features', link: '/features/storage' },
      { text: 'Reference', link: '/reference/configuration' },
      { text: 'Releases', link: `${repo}/releases` },
    ],
    sidebar: [
      {
        text: 'Get started',
        items: [
          { text: 'Quick start', link: '/guide/quick-start' },
          { text: 'On Linux', link: '/install/linux' },
          { text: 'With Docker', link: '/install/docker' },
          { text: 'On Kubernetes', link: '/install/kubernetes' },
          { text: 'Your first backup', link: '/guide/first-backup' },
        ],
      },
      {
        text: 'Features',
        items: [
          { text: 'Storage', link: '/features/storage' },
          { text: 'What and when to back up', link: '/features/backups' },
          { text: 'Retention', link: '/features/retention' },
          { text: 'Browse and restore', link: '/features/restore' },
          { text: 'Compare snapshots', link: '/features/compare' },
          { text: 'Existing repositories', link: '/features/connect' },
          { text: 'Notifications', link: '/features/notifications' },
          { text: 'Hooks', link: '/features/hooks' },
          { text: 'Metrics and Grafana', link: '/features/metrics' },
        ],
      },
      {
        text: 'Reference',
        items: [
          { text: 'Configuration', link: '/reference/configuration' },
          { text: 'Security', link: '/reference/security' },
          { text: 'API', link: '/reference/api' },
          { text: 'Upgrading from 0.1', link: '/reference/upgrading' },
        ],
      },
    ],
    socialLinks: [{ icon: 'github', link: repo }],
    editLink: { pattern: `${repo}/edit/main/docs/:path`, text: 'Edit this page on GitHub' },
    search: { provider: 'local' },
    footer: { message: 'Released under the GPL-3.0 license.', copyright: 'ResticGUI contributors' },
  },
});
