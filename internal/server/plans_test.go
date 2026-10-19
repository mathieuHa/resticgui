package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store"
)

func planInput(name string, include ...string) map[string]any {
	return map[string]any{
		"name":      name,
		"paths":     map[string]any{"include": include, "exclude": []string{}},
		"schedules": []map[string]any{{"job": "backup", "cron": "30 2 * * *"}},
	}
}

func (a *api) createPlan(repositoryID int64, in any) PlanDTO {
	a.t.Helper()
	status, body := a.do(http.MethodPost, "/repositories/"+itoa(repositoryID)+"/plans", in)
	require.Equal(a.t, http.StatusCreated, status, string(body))
	var dto PlanDTO
	a.decode(body, &dto)
	return dto
}

func TestCreateUpdateAndDeletePlans(t *testing.T) {
	a := newAPI(t)
	repo := a.create("home")
	base := "/repositories/" + itoa(repo.ID) + "/plans"

	in := planInput("Photos", "/home/me/Pictures", "/home/me/Pictures")
	in["tags"] = []string{"camera"}
	in["options"] = map[string]any{"compression": "max", "one_file_system": true, "exclude_if_present": []string{".nobackup"}, "limit_upload": 1024}
	in["schedules"] = []map[string]any{{"job": "backup", "cron": "30 2 * * *"}, {"job": "forget", "cron": "@weekly", "enabled": false}}
	in["retention"] = RetentionDTO{KeepDaily: 7, Prune: true}
	photos := a.createPlan(repo.ID, in)
	assert.Equal(t, "Photos", photos.Name)
	assert.True(t, strings.HasPrefix(photos.Tag, jobs.Tag(repo.ID)+":"), photos.Tag)
	assert.Equal(t, []string{"camera"}, photos.Tags)
	assert.Equal(t, PathsDTO{Include: []string{"/home/me/Pictures"}, Exclude: []string{}}, photos.Paths)
	assert.Equal(t, store.PlanOptions{Compression: "max", OneFileSystem: true, ExcludeIfPresent: []string{".nobackup"}, LimitUpload: 1024}, photos.Options)
	require.Len(t, photos.Schedules, 2)
	next := photos.Schedules[0].NextRun
	require.NotNil(t, next, "an enabled schedule says when it runs next")
	assert.Equal(t, 2, next.Local().Hour(), "cron runs in the server's time zone")
	assert.WithinDuration(t, time.Now(), *next, 24*time.Hour)
	assert.Nil(t, photos.Schedules[1].NextRun)
	assert.Equal(t, &RetentionDTO{KeepDaily: 7, Prune: true}, photos.Retention)
	assert.EqualValues(t, 1, a.reloads.n.Load(), "the scheduler reloads")

	docs := a.createPlan(repo.ID, planInput("Documents", "/home/me/Documents"))
	assert.NotEqual(t, photos.Tag, docs.Tag)
	assert.Equal(t, store.PlanOptions{Compression: "auto", ExcludeIfPresent: []string{}}, docs.Options, "the defaults")

	status, body := a.do(http.MethodGet, "/repositories/"+itoa(repo.ID), nil)
	require.Equal(t, http.StatusOK, status)
	var got RepositoryDTO
	a.decode(body, &got)
	assert.Equal(t, []PlanDTO{photos, docs}, got.Plans, "in the order they were made")

	// PUT replaces the whole plan, but keeps its tag
	update := planInput("Pictures", "/srv/pictures")
	update["schedules"] = []map[string]any{}
	status, body = a.do(http.MethodPut, base+"/"+itoa(photos.ID), update)
	require.Equal(t, http.StatusOK, status, string(body))
	var updated PlanDTO
	a.decode(body, &updated)
	assert.Equal(t, photos.Tag, updated.Tag)
	assert.Equal(t, []string{}, updated.Tags)
	assert.Equal(t, []ScheduleDTO{}, updated.Schedules)
	assert.Nil(t, updated.Retention)

	status, _ = a.do(http.MethodPost, base, planInput("Documents", "/x"))
	assert.Equal(t, http.StatusConflict, status, "names are unique in a repository")
	other := a.create("other")
	a.createPlan(other.ID, planInput("Documents", "/x"))
	status, _ = a.do(http.MethodPut, "/repositories/"+itoa(other.ID)+"/plans/"+itoa(photos.ID), update)
	assert.Equal(t, http.StatusNotFound, status, "a plan is reached through its repository")
	status, _ = a.do(http.MethodPost, "/repositories/999/plans", planInput("x", "/x"))
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = a.do(http.MethodPut, base+"/abc", update)
	assert.Equal(t, http.StatusBadRequest, status)

	status, _ = a.do(http.MethodDelete, base+"/"+itoa(docs.ID), nil)
	assert.Equal(t, http.StatusNoContent, status)
	status, _ = a.do(http.MethodDelete, base+"/"+itoa(docs.ID), nil)
	assert.Equal(t, http.StatusNotFound, status)
	_, body = a.do(http.MethodGet, "/repositories/"+itoa(repo.ID), nil)
	a.decode(body, &got)
	assert.Len(t, got.Plans, 1)
}

func TestPlanValidation(t *testing.T) {
	a := newAPI(t)
	repo := a.create("home")
	status, body := a.do(http.MethodPost, "/repositories/"+itoa(repo.ID)+"/plans", map[string]any{
		"name":  " ",
		"paths": map[string]any{"include": []string{"relative/dir", "/ok\n/etc"}, "exclude": []string{" ", "*.tmp\n/etc"}},
		"tags":  []string{"a,b", "resticgui:9", "with space"},
		"options": map[string]any{
			"compression": "fast", "exclude_larger_than": "big", "exclude_if_present": []string{""}, "limit_download": -1,
		},
		"schedules": []map[string]any{{"job": "check", "cron": "@daily"}, {"job": "backup", "cron": "every day"}, {"job": "forget", "cron": "@weekly"}},
	})
	require.Equal(t, http.StatusBadRequest, status)
	var e apiError
	a.decode(body, &e)
	details := strings.Join(e.Details, "\n")
	for _, want := range []string{
		"name is required", `"relative/dir" is not an absolute path`, "control character",
		`exclude pattern " " is empty`, `exclude pattern "*.tmp\n/etc"`,
		`tag "a,b"`, `tag "resticgui:9": tags starting with resticgui: are ResticGUI's own`, `tag "with space"`,
		"compression is one of", "exclude_larger_than is a size", "exclude_if_present", "limit_upload and limit_download",
		`not "check"; a check is scheduled on the repository`, "invalid schedule", "the forget schedule applies the retention policy",
	} {
		assert.Contains(t, details, want)
	}

	status, body = a.do(http.MethodPost, "/repositories/"+itoa(repo.ID)+"/plans", map[string]any{"name": "Empty"})
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "add at least one folder")
	in := planInput("Twice", "/x")
	in["schedules"] = []map[string]any{{"job": "backup", "cron": "@daily"}, {"job": "backup", "cron": "@hourly"}}
	in["retention"] = RetentionDTO{}
	status, body = a.do(http.MethodPost, "/repositories/"+itoa(repo.ID)+"/plans", in)
	require.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "given twice")
	assert.Contains(t, string(body), "must keep something")
	assert.EqualValues(t, 0, a.reloads.n.Load())
}

func TestRepositorySchedulesAreForTheCheck(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	base := "/repositories/" + itoa(repo.ID) + "/schedules/"

	for _, job := range []string{"backup", "forget", "restore"} {
		status, body := a.do(http.MethodPut, base+job, map[string]any{"cron": "@daily"})
		assert.Equal(t, http.StatusBadRequest, status, job)
		assert.Contains(t, string(body), "scheduled in its plans", job)
	}
	status, body := a.do(http.MethodPut, base+"check", map[string]any{"cron": "every day"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "invalid schedule")

	status, body = a.do(http.MethodPut, base+"check", map[string]any{"cron": "@monthly", "enabled": false})
	require.Equal(t, http.StatusOK, status, string(body))
	var dto RepositoryDTO
	a.decode(body, &dto)
	assert.Equal(t, []ScheduleDTO{{Job: "check", Cron: "@monthly", Enabled: false}}, dto.Schedules)
	status, body = a.do(http.MethodDelete, base+"check", nil)
	require.Equal(t, http.StatusOK, status)
	a.decode(body, &dto)
	assert.Empty(t, dto.Schedules)
	assert.EqualValues(t, 2, a.reloads.n.Load())
}

func TestRetentionPreviewOfAPlan(t *testing.T) {
	a := newAPI(t)
	repo := a.create("photos")
	plan := a.createPlan(repo.ID, planInput("Photos", "/photos"))
	path := "/repositories/" + itoa(repo.ID) + "/plans/" + itoa(plan.ID) + "/retention/preview"
	day := func(d int) time.Time { return time.Date(2026, 10, d, 2, 30, 0, 0, time.UTC) }
	a.restic.groups = []restic.ForgetGroup{{
		Keep:    []restic.Snapshot{{ID: "aaaa1111", Time: day(1)}, {ID: "cccc3333", Time: day(3)}},
		Remove:  []restic.Snapshot{{ID: "bbbb2222", Time: day(2)}},
		Reasons: []restic.KeepReason{{Snapshot: restic.Snapshot{ID: "cccc3333"}, Matches: []string{"daily snapshot"}}},
	}}

	status, body := a.do(http.MethodPost, path, RetentionDTO{})
	assert.Equal(t, http.StatusBadRequest, status, "an empty policy is refused before restic runs")
	assert.Contains(t, string(body), "must keep something")
	assert.Empty(t, a.restic.forgets)

	status, body = a.do(http.MethodPost, path, RetentionDTO{KeepDaily: 2})
	require.Equal(t, http.StatusOK, status, string(body))
	var preview RetentionPreview
	a.decode(body, &preview)
	require.Len(t, preview.Keep, 2)
	assert.Equal(t, "cccc3333", preview.Keep[0].ID, "newest first")
	assert.Equal(t, []string{"daily snapshot"}, preview.Keep[0].Matches)
	require.Len(t, preview.Remove, 1)
	assert.Equal(t, "bbbb2222", preview.Remove[0].ID)

	require.Len(t, a.restic.forgets, 1)
	opts := a.restic.forgets[0]
	assert.True(t, opts.DryRun, "a preview never forgets anything")
	assert.False(t, opts.Prune)
	assert.Equal(t, restic.Policy{KeepDaily: 2}, opts.Policy)
	assert.Equal(t, restic.SnapshotFilter{Host: "nas", Tags: []string{plan.Tag}}, opts.Filter, "same snapshots as a forget job of the plan")
	assert.Equal(t, "host", opts.GroupBy)

	a.restic.err = restic.ErrRepositoryNotFound
	status, _ = a.do(http.MethodPost, path, RetentionDTO{KeepDaily: 2})
	assert.Equal(t, http.StatusConflict, status)
	status, _ = a.do(http.MethodPost, "/repositories/"+itoa(repo.ID)+"/plans/999/retention/preview", RetentionDTO{KeepDaily: 2})
	assert.Equal(t, http.StatusNotFound, status)
}

func TestBackupOfAPlan(t *testing.T) {
	a := newAPI(t)
	repo := a.create("home")
	jobsPath := "/repositories/" + itoa(repo.ID) + "/jobs"

	status, body := a.do(http.MethodPost, jobsPath, map[string]any{"kind": "backup"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "plan_id is required: the repository has 0 plans")

	photos := a.createPlan(repo.ID, planInput("Photos", "/photos"))
	status, body = a.do(http.MethodPost, jobsPath, map[string]any{"kind": "backup"})
	require.Equal(t, http.StatusAccepted, status, string(body))
	var job jobs.JobDTO
	a.decode(body, &job)
	require.NotNil(t, job.PlanID, "the only plan")
	assert.Equal(t, photos.ID, *job.PlanID)

	docs := a.createPlan(repo.ID, planInput("Documents", "/docs"))
	status, body = a.do(http.MethodPost, jobsPath, map[string]any{"kind": "backup"})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(body), "plan_id is required: the repository has 2 plans")
	status, body = a.do(http.MethodPost, jobsPath, map[string]any{"kind": "backup", "plan_id": docs.ID})
	require.Equal(t, http.StatusAccepted, status, string(body))
	a.decode(body, &job)
	assert.Equal(t, docs.ID, *job.PlanID)

	other := a.create("other")
	elsewhere := a.createPlan(other.ID, planInput("Elsewhere", "/x"))
	status, _ = a.do(http.MethodPost, jobsPath, map[string]any{"kind": "backup", "plan_id": elsewhere.ID})
	assert.Equal(t, http.StatusNotFound, status)
	status, body = a.do(http.MethodPost, jobsPath, map[string]any{"kind": "check", "plan_id": docs.ID})
	require.Equal(t, http.StatusAccepted, status, string(body))
	a.decode(body, &job)
	assert.Nil(t, job.PlanID, "a check is of the whole repository")
}
