package jobs

import "sync"

// Event types sent to subscribers.
const (
	// EventStatus reports a job changing status (queued, running, finished).
	EventStatus = "job"
	// EventProgress reports progress of a running backup.
	EventProgress = "progress"
)

// Event is published on every job status change and, for backups, on
// progress. It is what the /events stream carries.
type Event struct {
	Type         string `json:"-"`
	JobID        int64  `json:"job_id"`
	RepositoryID int64  `json:"repository_id"`
	// PlanID is the plan of a backup or forget job.
	PlanID     *int64    `json:"plan_id,omitempty"`
	Kind       Kind      `json:"kind"`
	Status     Status    `json:"status"`
	Progress   *Progress `json:"progress,omitempty"`
	SnapshotID string    `json:"snapshot_id,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// Progress is the part of restic's backup status shown to users.
type Progress struct {
	PercentDone      float64  `json:"percent_done"`
	FilesDone        uint64   `json:"files_done"`
	TotalFiles       uint64   `json:"total_files"`
	BytesDone        uint64   `json:"bytes_done"`
	TotalBytes       uint64   `json:"total_bytes"`
	SecondsRemaining uint64   `json:"seconds_remaining"`
	CurrentFiles     []string `json:"current_files,omitempty"`
}

// Broker fans events out to subscribers without ever blocking a job.
type Broker struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: map[chan Event]struct{}{}}
}

// Subscribe returns a channel of events and a function that ends the
// subscription. A subscriber too slow to keep up loses progress events; if
// it would lose a status event, its channel is closed instead so that it
// reconnects and reloads the job list rather than show a stale status.
func (b *Broker) Subscribe(buffer int) (<-chan Event, func()) {
	ch := make(chan Event, buffer)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
	}
}

func (b *Broker) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- e:
		default:
			if e.Type != EventProgress {
				delete(b.subs, ch)
				close(ch)
			}
		}
	}
}
