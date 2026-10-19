package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

func seed(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(ctx, filepath.Join(t.TempDir(), "resticgui.db"))
	require.NoError(t, err)
	t.Cleanup(func() { st.Close() })
	repo := func(name string) db.Repository {
		r, err := st.CreateRepository(ctx, db.CreateRepositoryParams{Name: name, Backend: "local", Location: "/srv/" + name, Options: "{}", Secrets: []byte("x")})
		require.NoError(t, err)
		return r
	}
	job := func(repoID int64, kind, status, summary string) {
		j, err := st.CreateJob(ctx, db.CreateJobParams{RepositoryID: repoID, Kind: kind, Trigger: "schedule", Params: "{}"})
		require.NoError(t, err)
		_, err = st.FinishJob(ctx, db.FinishJobParams{ID: j.ID, Status: status, Summary: summary})
		require.NoError(t, err)
	}
	photos, docs := repo("Photos"), repo("Documents")
	repo("Empty")
	job(photos.ID, "backup", "succeeded", `{"data_added":2048,"total_duration":3.5,"total_bytes_processed":10000,"total_files_processed":12}`)
	job(photos.ID, "backup", "failed", "{}")
	job(photos.ID, "check", "succeeded", "{}")
	job(docs.ID, "backup", "succeeded", `{"data_added":10,"total_duration":1}`)
	job(docs.ID, "restore", "failed", "{}")
	// still running: not a finished job
	_, err = st.CreateJob(ctx, db.CreateJobParams{RepositoryID: docs.ID, Kind: "backup", Trigger: "manual", Params: "{}"})
	require.NoError(t, err)
	require.NoError(t, st.UpsertRepositoryStats(ctx, db.UpsertRepositoryStatsParams{RepositoryID: photos.ID, Snapshots: 7, TotalSize: 123456}))
	return st
}

func scrape(t *testing.T, h http.Handler, header string) (*httptest.ResponseRecorder, map[string]*dto.MetricFamily) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return rec, nil
	}
	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(rec.Body.String()))
	require.NoError(t, err, "valid Prometheus text format")
	return rec, families
}

// value returns the metric of family name whose labels include want.
func value(t *testing.T, families map[string]*dto.MetricFamily, name string, want map[string]string) (float64, bool) {
	t.Helper()
	f, ok := families[name]
	if !ok {
		return 0, false
	}
next:
	for _, m := range f.GetMetric() {
		labels := map[string]string{}
		for _, l := range m.GetLabel() {
			labels[l.GetName()] = l.GetValue()
		}
		for k, v := range want {
			if labels[k] != v {
				continue next
			}
		}
		return m.GetGauge().GetValue(), true
	}
	return 0, false
}

func TestMetrics(t *testing.T) {
	st := seed(t)
	_, families := scrape(t, Handler(st, ""), "")
	get := func(name string, labels map[string]string) float64 {
		v, ok := value(t, families, name, labels)
		require.True(t, ok, "%s %v", name, labels)
		return v
	}
	photos := map[string]string{"repository": "Photos"}
	photosBackup := map[string]string{"repository": "Photos", "kind": "backup"}

	assert.Equal(t, 1.0, get("resticgui_repository_info", map[string]string{"repository": "Empty", "backend": "local"}))
	assert.Equal(t, 123456.0, get("resticgui_repository_size_bytes", photos))
	assert.Equal(t, 7.0, get("resticgui_repository_snapshots", photos))
	assert.Equal(t, 0.0, get("resticgui_job_last_run_success", photosBackup), "the last backup failed")
	assert.Greater(t, get("resticgui_job_last_success_timestamp_seconds", photosBackup), 1.7e9)
	assert.Equal(t, 1.0, get("resticgui_job_last_run_success", map[string]string{"repository": "Photos", "kind": "check"}))
	assert.Equal(t, 3.5, get("resticgui_backup_last_duration_seconds", photos), "from the last successful backup")
	assert.Equal(t, 2048.0, get("resticgui_backup_last_added_bytes", photos))
	assert.Equal(t, 12.0, get("resticgui_backup_last_files", photos))
	assert.Equal(t, 1.0, get("resticgui_job_last_run_success", map[string]string{"repository": "Documents", "kind": "backup"}), "a running backup is not its last run")

	_, ok := value(t, families, "resticgui_job_last_run_success", map[string]string{"kind": "restore"})
	assert.False(t, ok, "restores are not tracked")
	_, ok = value(t, families, "resticgui_repository_size_bytes", map[string]string{"repository": "Documents"})
	assert.False(t, ok, "no size before the first refresh")
	assert.Contains(t, families, "go_goroutines")
}

func TestMetricsToken(t *testing.T) {
	h := Handler(seed(t), "s3cret")
	rec, _ := scrape(t, h, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	rec, _ = scrape(t, h, "Bearer wrong")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	rec, families := scrape(t, h, "Bearer s3cret")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, families, "resticgui_repository_info")
}
