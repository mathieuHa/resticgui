package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

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
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	Backend       string            `json:"backend"`
	Location      string            `json:"location"`
	Env           map[string]string `json:"env"`
	Secrets       []string          `json:"secrets"`
	HasPassword   bool              `json:"has_password"`
	CACert        string            `json:"cacert,omitempty"`
	Extended      map[string]string `json:"extended"`
	Host          string            `json:"host"`
	InitializedAt *time.Time        `json:"initialized_at"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	// Plans say what is backed up into the repository, in the order they
	// were made.
	Plans []PlanDTO `json:"plans"`
	// Schedules holds the integrity check's; backups and retention are
	// scheduled by plan.
	Schedules []ScheduleDTO `json:"schedules"`
	// Healthcheck is the Healthchecks check pinged by backups, null when unset.
	Healthcheck *HealthcheckDTO `json:"healthcheck"`
	// Hooks run around backups, in order.
	Hooks []HookDTO `json:"hooks"`
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
	// NextRun is when the schedule fires next; null when it is disabled.
	NextRun *time.Time `json:"next_run"`
	// ReadDataSubset is how much data a scheduled check reads, as restic's
	// --read-data-subset; empty reads none.
	ReadDataSubset string `json:"read_data_subset,omitempty"`
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
	Secrets  *map[string]string `json:"secrets"`
	CACert   string             `json:"cacert"`
	Extended map[string]string  `json:"extended"`
	Host     string             `json:"host"`
}

var (
	sizePattern     = regexp.MustCompile(`^[0-9]+[kKmMgGtT]?$`)
	durationPattern = regexp.MustCompile(`^([0-9]+[ymdh])+$`)
	extendedKey     = regexp.MustCompile(`^[a-z0-9]+\.[a-z0-9.-]+$`)
)

// validate checks the input against the secrets that will be stored.
func (in *RepositoryInput) validate(secretEnv map[string]string, creating bool, roots Roots, commands bool) error {
	var errs []error
	if in.Backend == "local" && filepath.IsAbs(in.Location) {
		if err := roots.checkLinks(in.Location); err != nil {
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
	if in.CACert != "" && (!filepath.IsAbs(in.CACert) || strings.ContainsFunc(in.CACert, unicode.IsControl)) {
		errs = append(errs, errors.New("cacert is an absolute path"))
	}
	for _, k := range sortedKeys(in.Extended) {
		if !extendedKey.MatchString(k) {
			errs = append(errs, fmt.Errorf("extended option %q is not a restic option name such as s3.storage-class", k))
		}
	}
	env := map[string]string{}
	for _, m := range []map[string]string{in.Env, secretEnv} {
		for k, v := range m {
			env[k] = v
		}
	}
	if found := restic.CommandOptions(in.Location, in.Extended, env); len(found) > 0 && !commands {
		errs = append(errs, fmt.Errorf("%s would run a program with ResticGUI's rights, which needs hooks.commands: true in the configuration", strings.Join(found, ", ")))
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
		b := backends.Registry[name]
		// lists, never null, as the API promises
		b.Env, b.Secrets, b.Required = orEmpty(b.Env), orEmpty(b.Secrets), orEmpty(b.Required)
		out = append(out, b)
	}
	c.JSON(http.StatusOK, out)
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
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
		Host:          r.Host,
		InitializedAt: r.InitializedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Plans: []PlanDTO{}, Schedules: []ScheduleDTO{},
	}
	if dto.Env == nil {
		dto.Env = map[string]string{}
	}
	if dto.Extended == nil {
		dto.Extended = map[string]string{}
	}

	plans, err := s.Store.ListPlans(ctx, r.ID)
	if err != nil {
		return RepositoryDTO{}, err
	}
	for _, p := range plans {
		plan, err := s.planDTO(ctx, p)
		if err != nil {
			return RepositoryDTO{}, err
		}
		dto.Plans = append(dto.Plans, plan)
	}
	schedules, err := s.Store.ListSchedules(ctx, r.ID)
	if err != nil {
		return RepositoryDTO{}, err
	}
	for _, sc := range schedules {
		sdto := scheduleDTO(sc.Job, sc.Cron, sc.Enabled)
		sdto.ReadDataSubset = sc.ReadDataSubset
		dto.Schedules = append(dto.Schedules, sdto)
	}
	if err := s.monitoringDTO(ctx, &dto); err != nil {
		return RepositoryDTO{}, err
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
	if err := in.validate(secretEnv, true, s.Roots, s.HookCommands); err != nil {
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
		Host: in.Host,
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
	if err := in.validate(sec.Env, false, s.Roots, s.HookCommands); err != nil {
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
			Host: in.Host,
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

// policy validates the retention rules and returns them as restic's.
func (in RetentionDTO) policy() (restic.Policy, error) {
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
	return policy, errors.Join(errs...)
}

// RetentionPreview is what a policy would keep and remove.
type RetentionPreview struct {
	Keep   []KeptSnapshot    `json:"keep"`
	Remove []restic.Snapshot `json:"remove"`
}

// KeptSnapshot is a snapshot a policy keeps and the rules that keep it.
type KeptSnapshot struct {
	restic.Snapshot
	Matches []string `json:"matches"`
}

// scheduleJob reads the :job of a repository's schedule: only the check,
// since backups and retention are scheduled by plan.
func scheduleJob(c *gin.Context) (string, bool) {
	job := c.Param("job")
	if jobs.Kind(job) == jobs.KindCheck {
		return job, true
	}
	fail(c, http.StatusBadRequest, "a repository schedules its check; backups and retention are scheduled in its plans")
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
		Cron           string `json:"cron"`
		Enabled        *bool  `json:"enabled"`
		ReadDataSubset string `json:"read_data_subset"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	if _, err := scheduler.Parse(in.Cron); err != nil {
		invalid(c, err)
		return
	}
	if err := jobs.ValidSubset(in.ReadDataSubset); err != nil {
		invalid(c, err)
		return
	}
	if !s.exists(c, id) {
		return
	}
	enabled := in.Enabled == nil || *in.Enabled
	if _, err := s.Store.UpsertSchedule(c.Request.Context(), db.UpsertScheduleParams{RepositoryID: id, Job: job, Cron: in.Cron, Enabled: enabled, ReadDataSubset: in.ReadDataSubset}); err != nil {
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
		// PlanID is the plan a backup or forget runs; it can be left out
		// when the repository has only one.
		PlanID int64 `json:"plan_id"`
		// ReadDataSubset is how much data a check reads.
		ReadDataSubset string `json:"read_data_subset"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	if in.ReadDataSubset != "" && in.Kind != jobs.KindCheck {
		fail(c, http.StatusBadRequest, "read_data_subset is only for a check")
		return
	}
	if err := jobs.ValidSubset(in.ReadDataSubset); err != nil {
		invalid(c, err)
		return
	}
	var job db.Job
	var err error
	if in.Kind == jobs.KindCheck {
		job, err = s.Jobs.EnqueueCheck(c.Request.Context(), id, jobs.CheckParams{ReadDataSubset: in.ReadDataSubset})
	} else {
		job, err = s.Jobs.Enqueue(c.Request.Context(), id, in.PlanID, in.Kind, jobs.TriggerManual)
	}
	s.respondJob(c, job, err)
}

// changePassword starts a job changing the repository's password, in
// restic and in ResticGUI.
func (s *Server) changePassword(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	if in.Password == "" {
		fail(c, http.StatusBadRequest, "password is required")
		return
	}
	job, err := s.Jobs.EnqueuePasswordChange(c.Request.Context(), id, in.Password)
	s.respondJob(c, job, err)
}

func (s *Server) respondJob(c *gin.Context, job db.Job, err error) {
	switch {
	case errors.Is(err, jobs.ErrPlanRequired):
		fail(c, http.StatusBadRequest, "plan_id is required: "+strings.TrimPrefix(err.Error(), jobs.ErrPlanRequired.Error()+": "))
	case errors.Is(err, jobs.ErrPlanNotFound):
		fail(c, http.StatusNotFound, "plan not found")
	case errors.Is(err, jobs.ErrUnknownKind):
		fail(c, http.StatusBadRequest, "kind is one of init, backup, forget, prune, check, unlock, repair_index, repair_snapshots, migrate")
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
	all := c.Query("all") == "true"
	filter := restic.SnapshotFilter{Host: r.Host}
	if all {
		// also snapshots made outside resticgui, such as by restic on the command line
		filter = restic.SnapshotFilter{}
	}
	snaps, err := s.Restic.Snapshots(ctx, target, filter)
	if err != nil {
		resticFailure(c, err)
		return
	}
	if !all {
		// the snapshots of every plan, also deleted ones
		mine := snaps[:0]
		for _, snap := range snaps {
			if jobs.OfRepository(r.ID, snap.Tags) {
				mine = append(mine, snap)
			}
		}
		snaps = mine
	}
	if r.InitializedAt == nil {
		if err := s.Store.MarkRepositoryInitialized(ctx, r.ID); err != nil {
			internal(c, err)
			return
		}
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
