package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// fakeClock moves only when the test advances it.
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	afters  int
	waiters []waiter
}

type waiter struct {
	at time.Time
	ch chan time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afters++
	ch := make(chan time.Time, 1)
	c.waiters = append(c.waiters, waiter{at: c.now.Add(d), ch: ch})
	return ch
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	kept := c.waiters[:0]
	for _, w := range c.waiters {
		if w.at.After(c.now) {
			kept = append(kept, w)
		} else {
			w.ch <- c.now
		}
	}
	c.waiters = kept
}

// waitAfters waits until the scheduler has asked for its n-th timer, which
// means it handled everything before and is waiting again.
func (c *fakeClock) waitAfters(t *testing.T, n int) {
	t.Helper()
	require.Eventually(t, func() bool {
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.afters >= n
	}, 5*time.Second, time.Millisecond, "the scheduler never waited for timer %d", n)
}

type source struct {
	mu        sync.Mutex
	schedules []store.Schedule
}

func (s *source) EnabledSchedules(context.Context) ([]store.Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]store.Schedule(nil), s.schedules...), nil
}

func (s *source) set(schedules ...store.Schedule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules = schedules
}

type enqueued struct {
	RepositoryID int64
	PlanID       int64
	Kind         jobs.Kind
	Trigger      jobs.Trigger
}

type enqueuer struct {
	calls chan enqueued
	err   error
}

func (e *enqueuer) Enqueue(_ context.Context, repositoryID, planID int64, kind jobs.Kind, trigger jobs.Trigger) (db.Job, error) {
	e.calls <- enqueued{repositoryID, planID, kind, trigger}
	return db.Job{ID: 1}, e.err
}

func (e *enqueuer) next(t *testing.T) enqueued {
	t.Helper()
	select {
	case c := <-e.calls:
		return c
	case <-time.After(5 * time.Second):
		t.Fatal("nothing was enqueued")
		return enqueued{}
	}
}

func (e *enqueuer) none(t *testing.T) {
	t.Helper()
	select {
	case c := <-e.calls:
		t.Fatalf("unexpected enqueue %+v", c)
	default:
	}
}

type harness struct {
	clock  *fakeClock
	source *source
	enq    *enqueuer
	sched  *Scheduler
	done   chan error
	cancel context.CancelFunc
}

func start(t *testing.T, enqErr error, schedules ...store.Schedule) *harness {
	t.Helper()
	h := &harness{
		clock:  &fakeClock{now: time.Date(2026, 10, 18, 2, 30, 0, 0, time.UTC)},
		source: &source{},
		enq:    &enqueuer{calls: make(chan enqueued, 16), err: enqErr},
		done:   make(chan error, 1),
	}
	h.source.set(schedules...)
	h.sched = New(h.source, h.enq, h.clock)
	var ctx context.Context
	ctx, h.cancel = context.WithCancel(context.Background())
	go func() { h.done <- h.sched.Run(ctx) }()
	t.Cleanup(func() {
		h.cancel()
		select {
		case <-h.done:
		case <-time.After(5 * time.Second):
			t.Error("Run did not return after cancel")
		}
	})
	return h
}

func TestFiresWhenDue(t *testing.T) {
	h := start(t, nil,
		store.Schedule{RepositoryID: 1, Job: "backup", Cron: "0 3 * * *"},
		store.Schedule{RepositoryID: 2, Job: "check", Cron: "@every 1h"},
	)
	h.clock.waitAfters(t, 1)

	h.clock.Advance(29 * time.Minute)
	h.clock.waitAfters(t, 1)
	h.enq.none(t)

	h.clock.Advance(time.Minute) // 03:00
	assert.Equal(t, enqueued{1, 0, jobs.KindBackup, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 2)
	h.enq.none(t)

	h.clock.Advance(30 * time.Minute) // 03:30
	assert.Equal(t, enqueued{2, 0, jobs.KindCheck, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 3)

	h.clock.Advance(time.Hour) // 04:30: check again, backup only tomorrow
	assert.Equal(t, enqueued{2, 0, jobs.KindCheck, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 4)
	h.enq.none(t)

	h.clock.Advance(22*time.Hour + 30*time.Minute) // 03:00 the next day
	assert.Equal(t, enqueued{1, 0, jobs.KindBackup, jobs.TriggerSchedule}, h.enq.next(t))
}

func TestSchedulesDueTogether(t *testing.T) {
	h := start(t, nil,
		store.Schedule{RepositoryID: 1, Job: "backup", Cron: "0 3 * * *"},
		store.Schedule{RepositoryID: 1, Job: "forget", Cron: "0 3 * * *"},
	)
	h.clock.waitAfters(t, 1)
	h.clock.Advance(30 * time.Minute)
	got := []enqueued{h.enq.next(t), h.enq.next(t)}
	assert.ElementsMatch(t, []enqueued{
		{1, 0, jobs.KindBackup, jobs.TriggerSchedule},
		{1, 0, jobs.KindForget, jobs.TriggerSchedule},
	}, got)
}

func TestInvalidScheduleIsSkipped(t *testing.T) {
	h := start(t, nil,
		store.Schedule{RepositoryID: 1, Job: "backup", Cron: "not a cron"},
		store.Schedule{RepositoryID: 2, Job: "backup", Cron: "0 3 * * *"},
	)
	h.clock.waitAfters(t, 1)
	h.clock.Advance(30 * time.Minute)
	assert.Equal(t, enqueued{2, 0, jobs.KindBackup, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 2)
	h.enq.none(t)
}

func TestReload(t *testing.T) {
	yearly := store.Schedule{RepositoryID: 9, Job: "check", Cron: "0 0 1 1 *"}
	h := start(t, nil, yearly)
	h.clock.waitAfters(t, 1)

	// a schedule added later fires without a restart
	h.source.set(yearly, store.Schedule{RepositoryID: 3, Job: "forget", Cron: "45 3 * * *"})
	h.sched.Reload()
	h.clock.waitAfters(t, 2)
	h.clock.Advance(75 * time.Minute) // 03:45
	assert.Equal(t, enqueued{3, 0, jobs.KindForget, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 3)

	// a replaced schedule stops firing
	h.source.set(yearly, store.Schedule{RepositoryID: 4, Job: "check", Cron: "0 12 * * *"})
	h.sched.Reload()
	h.clock.waitAfters(t, 4)
	h.clock.Advance(24 * time.Hour) // past 12:00 and past 03:45 the next day
	assert.Equal(t, enqueued{4, 0, jobs.KindCheck, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 5)
	h.enq.none(t)
}

func TestNoSchedules(t *testing.T) {
	h := start(t, nil)
	h.clock.Advance(48 * time.Hour)
	time.Sleep(20 * time.Millisecond)
	h.enq.none(t)
}

func TestKeepsRunningWhenAJobIsSkipped(t *testing.T) {
	h := start(t, jobs.ErrAlreadyQueued, store.Schedule{RepositoryID: 1, Job: "backup", Cron: "@every 1h"})
	for i := 1; i <= 3; i++ {
		h.clock.waitAfters(t, i)
		h.clock.Advance(time.Hour)
		h.enq.next(t)
	}
}

func TestParse(t *testing.T) {
	for _, spec := range []string{"0 3 * * *", "*/15 * * * 1-5", "@daily", "@every 6h"} {
		_, err := Parse(spec)
		assert.NoError(t, err, spec)
	}
	for _, spec := range []string{"", "0 3 * *", "0 0 3 * * *", "@sometimes", "61 * * * *"} {
		_, err := Parse(spec)
		assert.Error(t, err, spec)
	}
}

func TestTwoPlansOfOneRepositoryFireOnTheirOwn(t *testing.T) {
	h := start(t, nil,
		store.Schedule{RepositoryID: 1, PlanID: 10, Job: "backup", Cron: "0 3 * * *"},
		store.Schedule{RepositoryID: 1, PlanID: 11, Job: "backup", Cron: "0 * * * *"},
		store.Schedule{RepositoryID: 1, PlanID: 11, Job: "forget", Cron: "0 3 * * *"},
	)
	h.clock.waitAfters(t, 1)
	h.clock.Advance(30 * time.Minute) // 03:00
	got := []enqueued{h.enq.next(t), h.enq.next(t), h.enq.next(t)}
	assert.ElementsMatch(t, []enqueued{
		{1, 10, jobs.KindBackup, jobs.TriggerSchedule},
		{1, 11, jobs.KindBackup, jobs.TriggerSchedule},
		{1, 11, jobs.KindForget, jobs.TriggerSchedule},
	}, got)
	h.clock.waitAfters(t, 2)
	h.clock.Advance(time.Hour) // 04:00: only the hourly plan
	assert.Equal(t, enqueued{1, 11, jobs.KindBackup, jobs.TriggerSchedule}, h.enq.next(t))
	h.clock.waitAfters(t, 3)
	h.enq.none(t)
}
