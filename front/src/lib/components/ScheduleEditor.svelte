<script lang="ts">
  import { checkCron, defaults, describe, parse, toCron, WEEKDAYS, type Preset, type PresetKind } from '#lib/schedule.ts';

  // Edits a cron expression as a preset. null means no schedule.
  let { value = $bindable(null), label }: { value?: string | null; label: string } = $props();

  let preset: Preset = $state(parse(value ?? '30 2 * * *'));
  let enabled = $state(value !== null);

  $effect(() => {
    value = enabled ? toCron(preset) : null;
  });

  const kinds: [PresetKind, string][] = [
    ['hourly', 'Every hour'],
    ['every', 'Every few hours'],
    ['daily', 'Every day'],
    ['weekly', 'Every week'],
    ['monthly', 'Every month'],
    ['custom', 'Cron expression'],
  ];

  const time = (p: { hour: number; minute: number }) => `${String(p.hour).padStart(2, '0')}:${String(p.minute).padStart(2, '0')}`;
  function setTime(p: { hour: number; minute: number }, v: string) {
    const [h, m] = v.split(':').map(Number);
    if (h !== undefined && m !== undefined && !Number.isNaN(h) && !Number.isNaN(m)) {
      p.hour = h;
      p.minute = m;
    }
  }
  const cronError = $derived(preset.kind === 'custom' ? checkCron(preset.cron) : null);
</script>

<fieldset class="fieldset">
  <label class="label cursor-pointer gap-3">
    <input type="checkbox" class="toggle toggle-sm" bind:checked={enabled} />
    <span class="text-base-content font-medium">{label}</span>
  </label>
  {#if enabled}
    <div class="flex flex-wrap items-end gap-2">
      <select
        class="select select-sm w-44"
        aria-label="{label}: frequency"
        value={preset.kind}
        onchange={(e) => (preset = defaults(e.currentTarget.value as PresetKind, preset))}
      >
        {#each kinds as [kind, text] (kind)}
          <option value={kind}>{text}</option>
        {/each}
      </select>

      {#if preset.kind === 'hourly'}
        <label class="input input-sm w-36">
          <span class="label">minute</span>
          <input type="number" min="0" max="59" bind:value={preset.minute} />
        </label>
      {:else if preset.kind === 'every'}
        <label class="input input-sm w-36">
          <span class="label">hours</span>
          <input type="number" min="1" max="168" bind:value={preset.hours} />
        </label>
      {:else if preset.kind === 'custom'}
        <input class="input input-sm w-56 font-mono" aria-label="{label}: cron" placeholder="30 2 * * *" bind:value={preset.cron} />
      {:else}
        {#if preset.kind === 'weekly'}
          <select class="select select-sm w-36" aria-label="{label}: weekday" bind:value={preset.weekday}>
            {#each WEEKDAYS as day, i (i)}
              <option value={i}>{day}</option>
            {/each}
          </select>
        {:else if preset.kind === 'monthly'}
          <label class="input input-sm w-32">
            <span class="label">day</span>
            <input type="number" min="1" max="28" bind:value={preset.day} />
          </label>
        {/if}
        <input
          type="time"
          class="input input-sm w-32"
          aria-label="{label}: time"
          value={time(preset)}
          onchange={(e) => preset.kind !== 'custom' && 'hour' in preset && setTime(preset, e.currentTarget.value)}
        />
      {/if}
    </div>
    <p class="label {cronError ? 'text-error' : ''}">{cronError ?? describe(preset)}</p>
  {/if}
</fieldset>
