package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mathieuHa/resticgui/internal/backends"
	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/scheduler"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// RepositoryDTO is a repository as the API returns it. It never carries the
// password or the values of secret variables, only which ones are set.
type RepositoryDTO struct {
	ID                int64             `json:"id"`
	Name              string            `json:"name"`
	Backend           string            `json:"backend"`
	Location          string            `json:"location"`
	Env               map[string]string `json:"env"`
	Secrets           []string          `json:"secrets"`
	HasPassword       bool              `json:"has_password"`
	CACert            string            `json:"cacert,omitempty"`
	Extended          map[string]string `json:"extended"`
	Host              string            `json:"host"`
	Compression       string            `json:"compression"`
	ExcludeLargerThan string            `json:"exclude_larger_than"`
	InitializedAt     *time.Time        `json:"initialized_at"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Paths             PathsDTO          `json:"paths"`
	Retention         *RetentionDTO     `json:"retention"`
	Schedules         []ScheduleDTO     `json:"schedules"`
}

type PathsDTO struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

type RetentionDTO struct {
	KeepLast    int64  `json:"keep_last"`
	KeepHourly  int64  `json:"keep_hourly"`
	KeepDaily   int64  `json:"keep_daily"`
	KeepWeekly  int64  `json:"keep_weekly"`
	KeepMonthly int64  `json:"keep_monthly"`
	KeepYearly  int64  `json:"keep_yearly"`
	KeepWithin  string `json:"keep_within"`
	Prune       bool   `json:"prune"`
}

type ScheduleDTO struct {
	Job     string `json:"job"`
	Cron    string `json:"cron"`
	Enabled bool   `json:"enabled"`
}

// RepositoryInput creates or replaces a repository's settings.
type RepositoryInput struct {
	Name     string `json:"name"`
	Backend  string `json:"backend"`
	Location string `json:"location"`
	// Password is required on create; on update, empty keeps the current one.
	Password string            `json:"password"`
	Env      map[string]string `json:"env"`
	// Secrets replaces the secret variables; omitted (null) on update keeps them.
	Secrets           *map[string]string `json:"secrets"`
	CACert            string             `json:"cacert"`
	Extended          map[string]string  `json:"extended"`
	Host              string             `json:"host"`
	Compression       string             `json:"compression"`
	ExcludeLargerThan string             `json:"exclude_larger_than"`
}

var (
	sizePattern     = regexp.MustCompile(`^[0-9]+[kKmMgGtT]?$`)
	durationPattern = regexp.MustCompile(`^([0-9]+[ymdh])+$`)
	extendedKey     = regexp.MustCompile(`^[a-z0-9]+\.[a-z0-9.-]+$`)
)

// validate checks the input against the secrets that will be stored.
func (in *RepositoryInput) validate(secretEnv map[string]string, creating bool, roots Roots) error {
	var errs []error
	if in.Backend == "local" && filepath.IsAbs(in.Location) {
		if err := roots.Check(in.Location); err != nil {
			errs = append(errs, fmt.Errorf("location: %w", err))
		}
	}
	if strings.TrimSpace(in.Name) == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if creating && in.Password == "" {
		errs = append(errs, errors.New("password is required"))
	}
	if err := backends.Validate(in.Backend, in.Location, in.Env, secretEnv); err != nil {
		errs = append(errs, err)
	}
	if in.Compression == "" {
		in.Compression = "auto"
	}
	if in.Compression != "auto" && in.Compression != "off" && in.Compression != "max" {
		errs = append(errs, errors.New("compression is one of auto, off, max"))
	}
	if in.ExcludeLargerThan != "" && !sizePattern.MatchString(in.ExcludeLargerThan) {
		errs = append(errs, errors.New("exclude_larger_than is a size such as 500M or 2G"))
	}
	if in.CACert != "" && !filepath.IsAbs(in.CACert) {
		errs = append(errs, errors.New("cacert is an absolute path"))
	}
	for _, k := range sortedKeys(in.Extended) {
		if !extendedKey.MatchString(k) {
			errs = append(errs, fmt.Errorf("extended option %q is not a restic option name such as s3.storage-class", k))
		}
	}
	return errors.Join(errs...)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *Server) listBackends(c *gin.Context) {
	out := make([]backends.Backend, 0, len(backends.Registry))
	for _, name := range backends.Names() {
		out = append(out, backends.Registry[name])
	}
	c.JSON(http.StatusOK, out)
}

// repositoryDTO loads everything a repository response shows.
func (s *Server) repositoryDTO(ctx context.Context, r db.Repository) (RepositoryDTO, error) {
	opts, err := store.DecodeOptions(r.Options)
	if err != nil {
		return RepositoryDTO{}, err
	}
	sec, err := store.OpenSecrets(s.Box, r.Secrets)
	if err != nil {
		return RepositoryDTO{}, err
	}
	dto := RepositoryDTO{
		ID: r.ID, Name: r.Name, Backend: r.Backend, Location: r.Location,
		Env: opts.Env, Secrets: sortedKeys(sec.Env), HasPassword: sec.Password != "",
		CACert: opts.CACert, Extended: opts.Extended,
		Host: r.Host, Compression: r.Compression, ExcludeLargerThan: r.ExcludeLargerThan,
		InitializedAt: r.InitializedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Paths:     PathsDTO{Include: []string{}, Exclude: []string{}},
		Schedules: []ScheduleDTO{},
	}
	if dto.Env == nil {
		dto.Env = map[string]string{}
	}
	if dto.Extended == nil {
		dto.Extended = map[string]string{}
	}

	paths, err := s.Store.ListPaths(ctx, r.ID)
	if err != nil {
		return RepositoryDTO{}, err
	}
	for _, p := range paths {
		if p.Kind == "exclude" {
			dto.Paths.Exclude = append(dto.Paths.Exclude, p.Path)
		} else {
			dto.Paths.Include = append(dto.Paths.Include, p.Path)
		}
	}
	ret, err := s.Store.GetRetention(ctx, r.ID)
	switch {
	case err == nil:
		dto.Retention = &RetentionDTO{
			KeepLast: ret.KeepLast, KeepHourly: ret.KeepHourly, KeepDaily: ret.KeepDaily, KeepWeekly: ret.KeepWeekly,
			KeepMonthly: ret.KeepMonthly, KeepYearly: ret.KeepYearly, KeepWithin: ret.KeepWithin, Prune: ret.Prune,
		}
	case !errors.Is(err, sql.ErrNoRows):
		return RepositoryDTO{}, err
	}
	schedules, err := s.Store.ListSchedules(ctx, r.ID)
	if err != nil {
		return RepositoryDTO{}, err
	}
	for _, sc := range schedules {
		dto.Schedules = append(dto.Schedules, ScheduleDTO{Job: sc.Job, Cron: sc.Cron, Enabled: sc.Enabled})
	}
	return dto, nil
}

func (s *Server) listRepositories(c *gin.Context) {
	repos, err := s.Store.ListRepositories(c.Request.Context())
	if err != nil {
		internal(c, err)
		return
	}
	out := make([]RepositoryDTO, 0, len(repos))
	for _, r := range repos {
		dto, err := s.repositoryDTO(c.Request.Context(), r)
		if err != nil {
			internal(c, err)
			return
		}
		out = append(out, dto)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) getRepository(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

func (s *Server) respondRepository(c *gin.Context, status int, id int64) {
	r, err := s.Store.GetRepository(c.Request.Context(), id)
	if err != nil {
		notFound(c, err, "repository")
		return
	}
	dto, err := s.repositoryDTO(c.Request.Context(), r)
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(status, dto)
}

func (s *Server) createRepository(c *gin.Context) {
	var in RepositoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	var secretEnv map[string]string
	if in.Secrets != nil {
		secretEnv = *in.Secrets
	}
	if err := in.validate(secretEnv, true, s.Roots); err != nil {
		invalid(c, err)
		return
	}
	sealed, err := store.SealSecrets(s.Box, store.RepositorySecrets{Password: in.Password, Env: secretEnv})
	if err != nil {
		internal(c, err)
		return
	}
	options, err := store.EncodeOptions(store.RepositoryOptions{Env: in.Env, CACert: in.CACert, Extended: in.Extended})
	if err != nil {
		internal(c, err)
		return
	}
	r, err := s.Store.CreateRepository(c.Request.Context(), db.CreateRepositoryParams{
		Name: in.Name, Backend: in.Backend, Location: in.Location, Options: options, Secrets: sealed,
		Host: in.Host, Compression: in.Compression, ExcludeLargerThan: in.ExcludeLargerThan,
	})
	if isUniqueViolation(err) {
		fail(c, http.StatusConflict, fmt.Sprintf("a repository named %q already exists", in.Name))
		return
	}
	if err != nil {
		internal(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("%s/repositories/%d", BasePath, r.ID))
	s.respondRepository(c, http.StatusCreated, r.ID)
}

func (s *Server) updateRepository(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in RepositoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	ctx := c.Request.Context()
	current, err := s.Store.GetRepository(ctx, id)
	if err != nil {
		notFound(c, err, "repository")
		return
	}
	sec, err := store.OpenSecrets(s.Box, current.Secrets)
	if err != nil {
		internal(c, err)
		return
	}
	if in.Secrets != nil {
		sec.Env = *in.Secrets
	}
	if in.Password != "" {
		sec.Password = in.Password
	}
	if err := in.validate(sec.Env, false, s.Roots); err != nil {
		invalid(c, err)
		return
	}
	sealed, err := store.SealSecrets(s.Box, sec)
	if err != nil {
		internal(c, err)
		return
	}
	options, err := store.EncodeOptions(store.RepositoryOptions{Env: in.Env, CACert: in.CACert, Extended: in.Extended})
	if err != nil {
		internal(c, err)
		return
	}
	err = s.Store.InTx(ctx, func(q *db.Queries) error {
		if _, err := q.UpdateRepository(ctx, db.UpdateRepositoryParams{
			ID: id, Name: in.Name, Backend: in.Backend, Location: in.Location, Options: options,
			Host: in.Host, Compression: in.Compression, ExcludeLargerThan: in.ExcludeLargerThan,
		}); err != nil {
			return err
		}
		_, err := q.UpdateRepositorySecrets(ctx, db.UpdateRepositorySecretsParams{ID: id, Secrets: sealed})
		return err
	})
	if isUniqueViolation(err) {
		fail(c, http.StatusConflict, fmt.Sprintf("a repository named %q already exists", in.Name))
		return
	}
	if err != nil {
		internal(c, err)
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

func (s *Server) deleteRepository(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	if s.Jobs.Busy(id) {
		fail(c, http.StatusConflict, "a job is queued or running for this repository; cancel it first")
		return
	}
	n, err := s.Store.DeleteRepository(c.Request.Context(), id)
	if err != nil {
		internal(c, err)
		return
	}
	if n == 0 {
		fail(c, http.StatusNotFound, "repository not found")
		return
	}
	s.Scheduler.Reload()
	// the snapshots stay in the restic repository; only resticgui forgets it
	c.Status(http.StatusNoContent)
}

// exists answers 404 and returns false when the repository is missing.
func (s *Server) exists(c *gin.Context, id int64) bool {
	if _, err := s.Store.GetRepository(c.Request.Context(), id); err != nil {
		notFound(c, err, "repository")
		return false
	}
	return true
}

func (s *Server) putPaths(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in PathsDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	var errs []error
	for _, p := range in.Include {
		if err := s.Roots.Check(p); err != nil {
			errs = append(errs, err)
		}
	}
	for _, p := range in.Exclude {
		if strings.TrimSpace(p) == "" {
			errs = append(errs, errors.New("an exclude pattern is empty"))
		}
	}
	if err := errors.Join(errs...); err != nil {
		invalid(c, err)
		return
	}
	if !s.exists(c, id) {
		return
	}
	ctx := c.Request.Context()
	err := s.Store.InTx(ctx, func(q *db.Queries) error {
		if err := q.DeletePaths(ctx, id); err != nil {
			return err
		}
		for kind, paths := range map[string][]string{"include": in.Include, "exclude": in.Exclude} {
			for _, p := range dedupe(paths) {
				if _, err := q.AddPath(ctx, db.AddPathParams{RepositoryID: id, Path: p, Kind: kind}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		internal(c, err)
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

func dedupe(list []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range list {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func (s *Server) putRetention(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in RetentionDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	var errs []error
	for name, v := range map[string]int64{
		"keep_last": in.KeepLast, "keep_hourly": in.KeepHourly, "keep_daily": in.KeepDaily,
		"keep_weekly": in.KeepWeekly, "keep_monthly": in.KeepMonthly, "keep_yearly": in.KeepYearly,
	} {
		if v < 0 {
			errs = append(errs, fmt.Errorf("%s cannot be negative", name))
		}
	}
	if in.KeepWithin != "" && !durationPattern.MatchString(in.KeepWithin) {
		errs = append(errs, errors.New("keep_within is a duration such as 1y6m or 30d"))
	}
	policy := restic.Policy{
		KeepLast: int(in.KeepLast), KeepHourly: int(in.KeepHourly), KeepDaily: int(in.KeepDaily),
		KeepWeekly: int(in.KeepWeekly), KeepMonthly: int(in.KeepMonthly), KeepYearly: int(in.KeepYearly), KeepWithin: in.KeepWithin,
	}
	if policy.IsEmpty() {
		errs = append(errs, errors.New("the policy must keep something; delete the retention instead"))
	}
	if err := errors.Join(errs...); err != nil {
		invalid(c, err)
		return
	}
	if !s.exists(c, id) {
		return
	}
	if _, err := s.Store.UpsertRetention(c.Request.Context(), db.UpsertRetentionParams{
		RepositoryID: id, KeepLast: in.KeepLast, KeepHourly: in.KeepHourly, KeepDaily: in.KeepDaily,
		KeepWeekly: in.KeepWeekly, KeepMonthly: in.KeepMonthly, KeepYearly: in.KeepYearly,
		KeepWithin: in.KeepWithin, Prune: in.Prune,
	}); err != nil {
		internal(c, err)
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

func (s *Server) deleteRetention(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok || !s.exists(c, id) {
		return
	}
	if _, err := s.Store.DeleteRetention(c.Request.Context(), id); err != nil {
		internal(c, err)
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

func scheduleJob(c *gin.Context) (string, bool) {
	job := c.Param("job")
	switch jobs.Kind(job) {
	case jobs.KindBackup, jobs.KindForget, jobs.KindCheck:
		return job, true
	}
	fail(c, http.StatusBadRequest, "a schedule runs backup, forget or check")
	return "", false
}

func (s *Server) putSchedule(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	job, ok := scheduleJob(c)
	if !ok {
		return
	}
	var in struct {
		Cron    string `json:"cron"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	if _, err := scheduler.Parse(in.Cron); err != nil {
		invalid(c, err)
		return
	}
	if !s.exists(c, id) {
		return
	}
	enabled := in.Enabled == nil || *in.Enabled
	if _, err := s.Store.UpsertSchedule(c.Request.Context(), db.UpsertScheduleParams{RepositoryID: id, Job: job, Cron: in.Cron, Enabled: enabled}); err != nil {
		internal(c, err)
		return
	}
	s.Scheduler.Reload()
	s.respondRepository(c, http.StatusOK, id)
}

func (s *Server) deleteSchedule(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	job, ok := scheduleJob(c)
	if !ok || !s.exists(c, id) {
		return
	}
	if _, err := s.Store.DeleteSchedule(c.Request.Context(), db.DeleteScheduleParams{RepositoryID: id, Job: job}); err != nil {
		internal(c, err)
		return
	}
	s.Scheduler.Reload()
	s.respondRepository(c, http.StatusOK, id)
}

func (s *Server) startJob(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in struct {
		Kind jobs.Kind `json:"kind"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	job, err := s.Jobs.Enqueue(c.Request.Context(), id, in.Kind, jobs.TriggerManual)
	switch {
	case errors.Is(err, jobs.ErrUnknownKind):
		fail(c, http.StatusBadRequest, "kind is one of init, backup, forget, prune, check, unlock")
	case errors.Is(err, jobs.ErrRepositoryNotFound):
		fail(c, http.StatusNotFound, "repository not found")
	case errors.Is(err, jobs.ErrClosed):
		fail(c, http.StatusServiceUnavailable, "shutting down")
	case err != nil:
		internal(c, err)
	default:
		c.JSON(http.StatusAccepted, jobs.NewJobDTO(job))
	}
}

func (s *Server) listSnapshots(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	r, err := s.Store.GetRepository(ctx, id)
	if err != nil {
		notFound(c, err, "repository")
		return
	}
	target, err := s.resticRepo(r)
	if err != nil {
		internal(c, err)
		return
	}
	filter := restic.SnapshotFilter{Host: r.Host, Tags: []string{jobs.Tag(r.ID)}}
	if c.Query("all") == "true" {
		// also snapshots made outside resticgui, such as by restic on the command line
		filter = restic.SnapshotFilter{}
	}
	snaps, err := s.Restic.Snapshots(ctx, target, filter)
	if err != nil {
		resticFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, snaps)
}

func (s *Server) resticRepo(r db.Repository) (*restic.Repo, error) {
	return jobs.ResticRepo(s.Box, r)
}

// resticFailure maps a restic error to a response a user can act on.
func resticFailure(c *gin.Context, err error) {
	switch {
	case errors.Is(err, restic.ErrRepositoryNotFound):
		fail(c, http.StatusConflict, "the repository is not initialized; run an init job first")
	case errors.Is(err, restic.ErrWrongPassword):
		fail(c, http.StatusConflict, "the password does not open the repository")
	case errors.Is(err, restic.ErrLocked):
		fail(c, http.StatusConflict, "the repository is locked by another process")
	case errors.Is(err, context.Canceled):
		c.Abort()
	default:
		_ = c.Error(err)
		fail(c, http.StatusBadGateway, err.Error())
	}
}
