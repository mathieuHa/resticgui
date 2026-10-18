// Package metrics serves Prometheus metrics about the repositories: when
// each kind of job last ran and succeeded, the last backup's duration and
// data added, and each repository's size and snapshot count.
package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store"
)

var (
	repoLabels = []string{"repository"}
	jobLabels  = []string{"repository", "kind"}

	repositoryInfo = prometheus.NewDesc("resticgui_repository_info",
		"A repository, with its storage type.", []string{"repository", "backend"}, nil)
	repositorySize = prometheus.NewDesc("resticgui_repository_size_bytes",
		"Size of the repository's data, compressed, as of its last backup, retention or prune.", repoLabels, nil)
	repositorySnapshots = prometheus.NewDesc("resticgui_repository_snapshots",
		"Number of snapshots in the repository, as of its last backup, retention or prune.", repoLabels, nil)
	lastRun = prometheus.NewDesc("resticgui_job_last_run_timestamp_seconds",
		"When the last job of this kind finished, whatever its outcome.", jobLabels, nil)
	lastRunSuccess = prometheus.NewDesc("resticgui_job_last_run_success",
		"1 when the last job of this kind succeeded, 0 when it failed or was interrupted.", jobLabels, nil)
	lastSuccess = prometheus.NewDesc("resticgui_job_last_success_timestamp_seconds",
		"When the last successful job of this kind finished.", jobLabels, nil)
	backupDuration = prometheus.NewDesc("resticgui_backup_last_duration_seconds",
		"How long the last successful backup took.", repoLabels, nil)
	backupAdded = prometheus.NewDesc("resticgui_backup_last_added_bytes",
		"Data the last successful backup added to the repository, before compression.", repoLabels, nil)
	backupProcessed = prometheus.NewDesc("resticgui_backup_last_processed_bytes",
		"Size of the files the last successful backup read.", repoLabels, nil)
	backupFiles = prometheus.NewDesc("resticgui_backup_last_files",
		"Files the last successful backup read.", repoLabels, nil)
)

// Collector reads the metrics from the store on every scrape.
type Collector struct {
	Store *store.Store
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{repositoryInfo, repositorySize, repositorySnapshots, lastRun, lastRunSuccess, lastSuccess, backupDuration, backupAdded, backupProcessed, backupFiles} {
		ch <- d
	}
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repos, err := c.Store.ListRepositories(ctx)
	if err != nil {
		log.Error().Err(err).Msg("metrics: list the repositories")
		return
	}
	names := map[int64]string{}
	for _, r := range repos {
		names[r.ID] = r.Name
		ch <- prometheus.MustNewConstMetric(repositoryInfo, prometheus.GaugeValue, 1, r.Name, r.Backend)
	}
	gauge := func(d *prometheus.Desc, v float64, labels ...string) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v, labels...)
	}

	if stats, err := c.Store.ListRepositoryStats(ctx); err == nil {
		for _, s := range stats {
			if name, ok := names[s.RepositoryID]; ok {
				gauge(repositorySize, float64(s.TotalSize), name)
				gauge(repositorySnapshots, float64(s.Snapshots), name)
			}
		}
	}

	last, err := c.Store.LastJobs(ctx)
	if err != nil {
		log.Error().Err(err).Msg("metrics: read the last jobs")
		return
	}
	type key struct {
		repo int64
		kind string
	}
	latest := map[key]bool{}
	// LastJobs is ordered by id within a repository and kind: the last row
	// of a group is its latest finished job
	for i, j := range last {
		name, ok := names[j.RepositoryID]
		if !ok || j.FinishedAt == nil || j.Kind == "restore" {
			continue
		}
		k := key{j.RepositoryID, j.Kind}
		if i+1 == len(last) || last[i+1].RepositoryID != j.RepositoryID || last[i+1].Kind != j.Kind {
			latest[k] = true
			ok := 0.0
			if j.Status == "succeeded" {
				ok = 1
			}
			gauge(lastRun, unix(*j.FinishedAt), name, j.Kind)
			gauge(lastRunSuccess, ok, name, j.Kind)
		}
		if j.Status != "succeeded" {
			continue
		}
		gauge(lastSuccess, unix(*j.FinishedAt), name, j.Kind)
		var sum restic.BackupSummary
		if j.Kind == "backup" && json.Unmarshal([]byte(j.Summary), &sum) == nil {
			gauge(backupDuration, sum.TotalDuration, name)
			gauge(backupAdded, float64(sum.DataAdded), name)
			gauge(backupProcessed, float64(sum.TotalBytesProcessed), name)
			gauge(backupFiles, float64(sum.TotalFilesProcessed), name)
		}
	}
}

func unix(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}

// Handler serves the metrics in the Prometheus text format, with the Go
// runtime and process metrics. A non-empty token is required as a bearer
// token.
func Handler(st *store.Store, token string) http.Handler {
	reg := prometheus.NewRegistry()
	reg.MustRegister(Collector{Store: st}, collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	if token == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			w.Header().Set("WWW-Authenticate", `Bearer realm="resticgui metrics"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
