<script lang="ts">
  import { bars, linePath, niceMax, type BackupPoint } from '#lib/charts.ts';
  import { dateTime } from '#lib/format.ts';

  // A small chart of one value per backup: bars, or a line with dots.
  let {
    title,
    points,
    value,
    format,
    kind = 'bar',
    bytes = false,
    tone = 'primary',
  }: {
    title: string;
    points: BackupPoint[];
    value: (p: BackupPoint) => number | undefined;
    format: (n: number | undefined) => string;
    kind?: 'bar' | 'line';
    bytes?: boolean;
    tone?: 'primary' | 'secondary' | 'accent';
  } = $props();

  const W = 300;
  const H = 96;

  const values = $derived(points.map(value).filter((v): v is number => v !== undefined));
  const max = $derived(niceMax(Math.max(0, ...values), bytes ? 1024 : 10));
  const latest = $derived(values.at(-1));
  const barList = $derived(kind === 'bar' ? bars(points, value, W, H, max) : []);
  const path = $derived(kind === 'line' ? linePath(points, value, W, H, max) : '');
  const step = $derived(points.length > 1 ? W / (points.length - 1) : 0);

  const FILL = { primary: 'fill-primary', secondary: 'fill-secondary', accent: 'fill-accent' };
  const STROKE = { primary: 'stroke-primary', secondary: 'stroke-secondary', accent: 'stroke-accent' };

  let hover = $state<number | null>(null);
  const hovered = $derived(hover === null ? undefined : points[hover]);
</script>

<figure class="flex flex-col gap-2" data-testid="chart">
  <figcaption class="flex items-baseline justify-between gap-2">
    <span class="text-base-content/70 text-sm">{title}</span>
    <span class="tabular text-sm font-semibold">
      {#if hovered}{format(value(hovered))}{:else}{format(latest)}{/if}
    </span>
  </figcaption>
  {#if values.length === 0}
    <div class="bg-base-200 text-base-content/70 grid aspect-[300/96] place-items-center rounded-lg text-xs">No data yet</div>
  {:else}
    <!-- the numbers are in the caption and the tooltips; the drawing is decoration -->
    <svg
      viewBox="-4 -4 {W + 8} {H + 8}"
      class="h-auto w-full overflow-visible"
      aria-hidden="true"
      onmouseleave={() => (hover = null)}
    >
      <line x1="0" x2={W} y1={H} y2={H} class="stroke-base-300" stroke-width="1" />
      <line x1="0" x2={W} y1={H / 2} y2={H / 2} class="stroke-base-300" stroke-width="1" stroke-dasharray="3 4" />
      {#if kind === 'bar'}
        {#each barList as b, i (b.point.id)}
          <rect
            x={b.x}
            y={b.y}
            width={b.width}
            height={b.height}
            rx="1.5"
            class="{FILL[tone]} transition-opacity {hover === null || hover === i ? 'opacity-90' : 'opacity-40'}"
          />
          <rect x={i * (W / points.length)} y="0" width={W / points.length} height={H} fill="transparent" role="presentation" onmouseenter={() => (hover = i)}>
            <title>{dateTime(new Date(b.point.time).toISOString())}: {format(b.value)}</title>
          </rect>
        {/each}
      {:else}
        <path d={path} fill="none" class={STROKE[tone]} stroke-width="2" stroke-linejoin="round" stroke-linecap="round" />
        {#each points as p, i (p.id)}
          {@const v = value(p)}
          {#if v !== undefined}
            <circle cx={points.length > 1 ? i * step : W / 2} cy={H - (v / max) * H} r={hover === i ? 3.5 : 2} class={FILL[tone]} />
          {/if}
          <rect x={(points.length > 1 ? i * step : W / 2) - step / 2} y="0" width={Math.max(step, 8)} height={H} fill="transparent" role="presentation" onmouseenter={() => (hover = i)}>
            <title>{dateTime(new Date(p.time).toISOString())}: {format(v)}</title>
          </rect>
        {/each}
      {/if}
    </svg>
    <div class="text-base-content/70 flex justify-between text-xs">
      <time datetime={new Date(points[0]!.time).toISOString()}>{dateTime(new Date(points[0]!.time).toISOString())}</time>
      <span>max {format(max)}</span>
    </div>
  {/if}
</figure>
