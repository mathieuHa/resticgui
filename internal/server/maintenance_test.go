package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
)

// finished waits for a job to end and returns it.
func (a *api) finished(id int64) jobs.JobDTO {
	a.t.Helper()
	var job jobs.JobDTO
	require.Eventually(a.t, func() bool {
		_, body := a.do(http.MethodGet, "/jobs/"+itoa(id), nil)
		a.decode(body, &job)
		return job.FinishedAt != nil
	}, 5*time.Second, 10*time.Millisecond)
	return job
}

func (a *api) start(path string, in map[string]any) jobs.JobDTO {
	a.t.Helper()
	status, body := a.do(http.MethodPost, path, in)
	require.Equal(a.t, http.StatusAccepted, status, string(body))
	var job jobs.JobDTO
	a.decode(body, &job)
	return a.finished(job.ID)
}

func TestMaintenanceJobs(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/jobs"
	a.restic.modified = 1

	job := a.start(path, map[string]any{"kind": "repair_index"})
	assert.Equal(t, jobs.StatusSucceeded, job.Status)
	job = a.start(path, map[string]any{"kind": "repair_snapshots"})
	assert.Equal(t, jobs.StatusSucceeded, job.Status)
	assert.JSONEq(t, `{"modified":1}`, string(job.Summary))
	job = a.start(path, map[string]any{"kind": "migrate"})
	assert.Equal(t, jobs.StatusSucceeded, job.Status)
	assert.JSONEq(t, `{"applied":["upgrade_repo_v2"]}`, string(job.Summary))
	// repairing snapshots changes the size: the stats are read again after it
	assert.Equal(t, []string{"repair index", "repair snapshots", "stats raw-data ", "migrate upgrade_repo_v2"}, a.restic.calls)

	// a password changes only through its own route
	status, body := a.do(http.MethodPost, path, map[string]any{"kind": "change_password"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "repair_index")
}

func TestCheckWithADataSubset(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/jobs"

	job := a.start(path, map[string]any{"kind": "check", "read_data_subset": "10%"})
	assert.Equal(t, jobs.StatusSucceeded, job.Status)
	assert.JSONEq(t, `{"read_data_subset":"10%"}`, string(job.Params))
	a.start(path, map[string]any{"kind": "check"})
	assert.Equal(t, []restic.CheckOptions{{ReadDataSubset: "10%"}, {}}, a.restic.checks)

	status, body := a.do(http.MethodPost, path, map[string]any{"kind": "check", "read_data_subset": "--help"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "read_data_subset")
	status, body = a.do(http.MethodPost, path, map[string]any{"kind": "prune", "read_data_subset": "10%"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "only for a check")

	schedule := "/repositories/" + itoa(repo.ID) + "/schedules/check"
	status, body = a.do(http.MethodPut, schedule, map[string]any{"cron": "@weekly", "read_data_subset": "1/5"})
	require.Equal(t, http.StatusOK, status, string(body))
	var dto RepositoryDTO
	a.decode(body, &dto)
	require.Len(t, dto.Schedules, 1)
	assert.Equal(t, "1/5", dto.Schedules[0].ReadDataSubset)
	status, _ = a.do(http.MethodPut, schedule, map[string]any{"cron": "@weekly", "read_data_subset": "a lot"})
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangePassword(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	path := "/repositories/" + itoa(repo.ID) + "/password"
	a.restic.keys = []string{"key of " + password}

	status, _ := a.do(http.MethodPost, path, map[string]any{"password": ""})
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = a.do(http.MethodPost, "/repositories/999/password", map[string]any{"password": "x"})
	assert.Equal(t, http.StatusNotFound, status)

	status, body := a.do(http.MethodPost, path, map[string]any{"password": "a brand new password"})
	require.Equal(t, http.StatusAccepted, status, string(body))
	assert.NotContains(t, string(body), "a brand new password")
	var job jobs.JobDTO
	a.decode(body, &job)
	assert.Equal(t, jobs.KindChangePassword, job.Kind)
	job = a.finished(job.ID)
	assert.Equal(t, jobs.StatusSucceeded, job.Status, job.Error)

	assert.Equal(t, "a brand new password", a.secrets(repo.ID).Password)
	assert.Equal(t, []string{"key of a brand new password"}, a.restic.keys)
	_, body = a.do(http.MethodGet, "/jobs?repository="+itoa(repo.ID), nil)
	assert.NotContains(t, string(body), "a brand new password")
}
