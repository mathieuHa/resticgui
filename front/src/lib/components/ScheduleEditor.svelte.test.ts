import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { flushSync } from 'svelte';
import { describe, expect, it } from 'vitest';

import ScheduleEditor from './ScheduleEditor.svelte';

// ScheduleEditor binds value; tests read it back through the props object
function setup(value: string | null) {
  const props = $state({ value, label: 'Back up' });
  render(ScheduleEditor, { props });
  return props;
}

describe('ScheduleEditor', () => {
  it('shows a stored cron expression as a preset', () => {
    setup('0 23 * * 5');
    expect(screen.getByLabelText('Back up: frequency')).toHaveValue('weekly');
    expect(screen.getByLabelText('Back up: weekday')).toHaveValue('5');
    expect(screen.getByText('Every Friday at 23:00')).toBeInTheDocument();
  });

  it('writes the cron expression of the chosen preset', async () => {
    const user = userEvent.setup();
    const props = setup('30 2 * * *');
    await user.selectOptions(screen.getByLabelText('Back up: frequency'), 'monthly');
    flushSync();
    expect(props.value).toBe('30 2 1 * *');
    expect(screen.getByText('On the 1st of every month at 02:30')).toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText('Back up: frequency'), 'custom');
    const cron = screen.getByLabelText('Back up: cron');
    await user.clear(cron);
    await user.type(cron, '*/15 * * * *');
    flushSync();
    expect(props.value).toBe('*/15 * * * *');
  });

  it('flags an invalid custom expression', async () => {
    const user = userEvent.setup();
    setup('*/15 * * * *');
    const cron = screen.getByLabelText('Back up: cron');
    await user.clear(cron);
    await user.type(cron, 'sometimes');
    expect(screen.getByText(/five fields/)).toBeInTheDocument();
  });

  it('turns the schedule off and on', async () => {
    const user = userEvent.setup();
    const props = setup(null);
    expect(screen.queryByLabelText('Back up: frequency')).not.toBeInTheDocument();
    await user.click(screen.getByRole('checkbox'));
    flushSync();
    expect(props.value).toBe('30 2 * * *');
    await user.click(screen.getByRole('checkbox'));
    flushSync();
    expect(props.value).toBeNull();
  });
});
