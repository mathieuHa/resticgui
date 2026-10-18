package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mathieuHa/resticgui/internal/hooks"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// loadHooks returns the hooks of a repository, webhook URLs decrypted.
func (m *Manager) loadHooks(ctx context.Context, repositoryID int64) ([]hooks.Hook, error) {
	rows, err := m.store.ListHooks(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	list := make([]hooks.Hook, 0, len(rows))
	for _, r := range rows {
		h := hooks.Hook{Event: r.Event, Kind: r.Kind, Command: r.Command, Timeout: time.Duration(r.TimeoutSeconds) * time.Second}
		if r.Kind == hooks.Webhook {
			url, err := m.box.Open(r.Url)
			if err != nil {
				return nil, fmt.Errorf("hook %d: decrypt the webhook URL: %w", r.ID, err)
			}
			h.URL = string(url)
		}
		list = append(list, h)
	}
	return list, nil
}

// runHooks runs the hooks of one event in order, adding their output to
// log. A failing before hook stops the ones after it; after a backup, every
// hook runs. It returns the first failure.
func (m *Manager) runHooks(ctx context.Context, list []hooks.Hook, event string, run hooks.Run, log *[]string) error {
	run.Event = event
	var first error
	n := 0
	for _, h := range list {
		if h.Event != event {
			continue
		}
		n++
		label := fmt.Sprintf("[%s hook %d]", event, n)
		lines, err := m.hooks.Do(ctx, h, run)
		for _, l := range lines {
			*log = append(*log, label+" "+l)
		}
		if err != nil {
			*log = append(*log, label+" failed: "+err.Error())
			if first == nil {
				first = fmt.Errorf("%s hook %d failed: %w", event, n, err)
			}
			if event == hooks.Before {
				break
			}
		}
	}
	return first
}

// withHooks runs a backup between the repository's hooks: before hooks
// first (a failure fails the job without backing up), then the success or
// failure hooks with the outcome. Their output goes to the job log.
func (m *Manager) withHooks(ctx context.Context, job db.Job, repo db.Repository, backup func() outcome) outcome {
	list, err := m.loadHooks(ctx, repo.ID)
	if err != nil {
		return failed(err)
	}
	if len(list) == 0 {
		return backup()
	}
	run := hooks.Run{RepositoryID: repo.ID, Repository: repo.Name, JobID: job.ID}
	var log []string
	var out outcome
	if err := m.runHooks(ctx, list, hooks.Before, run, &log); err != nil {
		out = failed(err)
	} else {
		out = backup()
	}
	before := log
	log = nil

	run.Status, run.SnapshotID = string(StatusSucceeded), out.snapshotID
	event := hooks.Success
	if out.status != StatusSucceeded || ctx.Err() != nil {
		event, run.Status = hooks.Failure, string(StatusFailed)
		if out.err != nil {
			run.Error = out.err.Error()
		}
	}
	// after hooks run even when the job was cancelled, to clean up what the
	// before hooks set up
	if err := m.runHooks(context.WithoutCancel(ctx), list, event, run, &log); err != nil && out.status == StatusSucceeded {
		// the snapshot exists: the job succeeded, and its error says what failed
		out.err = errors.Join(out.err, err)
	}
	out.stderr = append(append(before, out.stderr...), log...)
	return out
}
