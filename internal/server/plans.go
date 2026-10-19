package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/mathieuHa/resticgui/internal/jobs"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/scheduler"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// PlanDTO is a backup plan as the API returns it: what to back up into a
// repository, how, when, and how long to keep it.
type PlanDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Tag is put on every snapshot of the plan; its retention only applies
	// to snapshots with this tag.
	Tag       string            `json:"tag"`
	Tags      []string          `json:"tags"`
	Paths     PathsDTO          `json:"paths"`
	Options   store.PlanOptions `json:"options"`
	Schedules []ScheduleDTO     `json:"schedules"`
	Retention *RetentionDTO     `json:"retention"`
	CreatedAt time.Time         `json:"created_at"`
}

// PlanInput creates or replaces a plan.
type PlanInput struct {
	Name      string            `json:"name"`
	Tags      []string          `json:"tags"`
	Paths     PathsDTO          `json:"paths"`
	Options   store.PlanOptions `json:"options"`
	Schedules []ScheduleInput   `json:"schedules"`
	// Retention is null for a plan that keeps every snapshot.
	Retention *RetentionDTO `json:"retention"`
}

// ScheduleInput sets when a job runs.
type ScheduleInput struct {
	Job  string `json:"job"`
	Cron string `json:"cron"`
	// Enabled defaults to true.
	Enabled *bool `json:"enabled"`
}

const maxTags = 20

func hasControl(s string) bool {
	return strings.ContainsFunc(s, unicode.IsControl)
}

// validate checks the input and returns it as the store saves it.
func (in *PlanInput) validate(roots Roots) (store.Plan, error) {
	var errs []error
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || hasControl(in.Name) {
		errs = append(errs, errors.New("name is required, on one line"))
	}
	if len(in.Paths.Include) == 0 {
		errs = append(errs, errors.New("add at least one folder or file to back up"))
	}
	for _, p := range in.Paths.Include {
		if err := roots.checkLinks(p); err != nil {
			errs = append(errs, err)
		}
	}
	for _, p := range in.Paths.Exclude {
		if strings.TrimSpace(p) == "" || hasControl(p) {
			errs = append(errs, fmt.Errorf("exclude pattern %q is empty or not on one line", p))
		}
	}
	if len(in.Tags) > maxTags {
		errs = append(errs, fmt.Errorf("a plan has at most %d tags", maxTags))
	}
	for _, t := range in.Tags {
		switch {
		case t == "" || hasControl(t) || strings.ContainsAny(t, ", "):
			errs = append(errs, fmt.Errorf("tag %q is empty, or has a comma, a space or a line break", t))
		case strings.HasPrefix(t, "resticgui:"):
			errs = append(errs, fmt.Errorf("tag %q: tags starting with resticgui: are ResticGUI's own", t))
		}
	}
	o := &in.Options
	if o.Compression == "" {
		o.Compression = "auto"
	}
	if o.Compression != "auto" && o.Compression != "off" && o.Compression != "max" {
		errs = append(errs, errors.New("compression is one of auto, off, max"))
	}
	if o.ExcludeLargerThan != "" && !sizePattern.MatchString(o.ExcludeLargerThan) {
		errs = append(errs, errors.New("exclude_larger_than is a size such as 500M or 2G"))
	}
	for _, name := range o.ExcludeIfPresent {
		if strings.TrimSpace(name) == "" || hasControl(name) {
			errs = append(errs, fmt.Errorf("exclude_if_present %q is empty or not on one line", name))
		}
	}
	if o.LimitUpload < 0 || o.LimitDownload < 0 {
		errs = append(errs, errors.New("limit_upload and limit_download are KiB/s, 0 for no limit"))
	}

	plan := store.Plan{Name: in.Name, Tags: in.Tags, Options: *o, Include: in.Paths.Include, Exclude: in.Paths.Exclude}
	seen := map[string]bool{}
	for _, s := range in.Schedules {
		if s.Job != string(jobs.KindBackup) && s.Job != string(jobs.KindForget) {
			errs = append(errs, fmt.Errorf("a plan schedules backup or forget, not %q; a check is scheduled on the repository", s.Job))
			continue
		}
		if seen[s.Job] {
			errs = append(errs, fmt.Errorf("the %s schedule is given twice", s.Job))
		}
		seen[s.Job] = true
		if _, err := scheduler.Parse(s.Cron); err != nil {
			errs = append(errs, err)
		}
		plan.Schedules = append(plan.Schedules, db.PlanSchedule{Job: s.Job, Cron: s.Cron, Enabled: s.Enabled == nil || *s.Enabled})
	}
	if r := in.Retention; r != nil {
		if _, err := r.policy(); err != nil {
			errs = append(errs, err)
		}
		plan.Retention = &db.PlanRetention{
			KeepLast: r.KeepLast, KeepHourly: r.KeepHourly, KeepDaily: r.KeepDaily, KeepWeekly: r.KeepWeekly,
			KeepMonthly: r.KeepMonthly, KeepYearly: r.KeepYearly, KeepWithin: r.KeepWithin, Prune: r.Prune,
		}
	} else if seen[string(jobs.KindForget)] {
		errs = append(errs, errors.New("the forget schedule applies the retention policy: set one, or remove the schedule"))
	}
	return plan, errors.Join(errs...)
}

// planDTO loads everything a plan response shows.
func (s *Server) planDTO(ctx context.Context, p db.Plan) (PlanDTO, error) {
	dto := PlanDTO{
		ID: p.ID, Name: p.Name, Tag: p.Tag, CreatedAt: p.CreatedAt,
		Paths: PathsDTO{Include: []string{}, Exclude: []string{}}, Schedules: []ScheduleDTO{},
	}
	var err error
	if dto.Tags, err = store.DecodeTags(p.Tags); err != nil {
		return PlanDTO{}, err
	}
	if dto.Options, err = store.DecodePlanOptions(p.Options); err != nil {
		return PlanDTO{}, err
	}
	paths, err := s.Store.ListPlanPaths(ctx, p.ID)
	if err != nil {
		return PlanDTO{}, err
	}
	for _, path := range paths {
		if path.Kind == "exclude" {
			dto.Paths.Exclude = append(dto.Paths.Exclude, path.Path)
		} else {
			dto.Paths.Include = append(dto.Paths.Include, path.Path)
		}
	}
	schedules, err := s.Store.ListPlanSchedules(ctx, p.ID)
	if err != nil {
		return PlanDTO{}, err
	}
	for _, sc := range schedules {
		dto.Schedules = append(dto.Schedules, scheduleDTO(sc.Job, sc.Cron, sc.Enabled))
	}
	ret, err := s.Store.GetPlanRetention(ctx, p.ID)
	switch {
	case err == nil:
		dto.Retention = &RetentionDTO{
			KeepLast: ret.KeepLast, KeepHourly: ret.KeepHourly, KeepDaily: ret.KeepDaily, KeepWeekly: ret.KeepWeekly,
			KeepMonthly: ret.KeepMonthly, KeepYearly: ret.KeepYearly, KeepWithin: ret.KeepWithin, Prune: ret.Prune,
		}
	case !errors.Is(err, sql.ErrNoRows):
		return PlanDTO{}, err
	}
	return dto, nil
}

func scheduleDTO(job, cron string, enabled bool) ScheduleDTO {
	out := ScheduleDTO{Job: job, Cron: cron, Enabled: enabled}
	if spec, err := scheduler.Parse(cron); err == nil && enabled {
		next := spec.Next(time.Now()).UTC()
		out.NextRun = &next
	}
	return out
}

// plan reads the :plan of the repository :id, answering 400 or 404 and
// false when there is none.
func (s *Server) plan(c *gin.Context) (db.Plan, bool) {
	repoID, ok := repositoryID(c)
	if !ok {
		return db.Plan{}, false
	}
	id, err := strconv.ParseInt(c.Param("plan"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "invalid plan id")
		return db.Plan{}, false
	}
	p, err := s.Store.GetPlan(c.Request.Context(), id)
	if err == nil && p.RepositoryID != repoID {
		err = sql.ErrNoRows
	}
	if err != nil {
		notFound(c, err, "plan")
		return db.Plan{}, false
	}
	return p, true
}

func (s *Server) createPlan(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	s.savePlan(c, id, 0, http.StatusCreated)
}

func (s *Server) updatePlan(c *gin.Context) {
	p, ok := s.plan(c)
	if !ok {
		return
	}
	s.savePlan(c, p.RepositoryID, p.ID, http.StatusOK)
}

func (s *Server) savePlan(c *gin.Context, repositoryID, planID int64, status int) {
	var in PlanInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	plan, err := in.validate(s.Roots)
	if err != nil {
		invalid(c, err)
		return
	}
	if !s.exists(c, repositoryID) {
		return
	}
	ctx := c.Request.Context()
	var saved db.Plan
	err = s.Store.InTx(ctx, func(q *db.Queries) error {
		var err error
		saved, err = store.SavePlan(ctx, q, repositoryID, planID, plan)
		return err
	})
	if isUniqueViolation(err) {
		fail(c, http.StatusConflict, fmt.Sprintf("the repository already has a plan named %q", plan.Name))
		return
	}
	if err != nil {
		internal(c, err)
		return
	}
	s.Scheduler.Reload()
	dto, err := s.planDTO(ctx, saved)
	if err != nil {
		internal(c, err)
		return
	}
	if status == http.StatusCreated {
		c.Header("Location", fmt.Sprintf("%s/repositories/%d/plans/%d", BasePath, repositoryID, saved.ID))
	}
	c.JSON(status, dto)
}

func (s *Server) deletePlan(c *gin.Context) {
	p, ok := s.plan(c)
	if !ok {
		return
	}
	if s.Jobs.Busy(p.RepositoryID) {
		fail(c, http.StatusConflict, "a job is queued or running for this repository; wait for it or cancel it first")
		return
	}
	if _, err := s.Store.DeletePlan(c.Request.Context(), p.ID); err != nil {
		internal(c, err)
		return
	}
	s.Scheduler.Reload()
	// the plan's snapshots stay in the repository
	c.Status(http.StatusNoContent)
}

// previewRetention runs "restic forget --dry-run" with the given policy,
// which need not be saved yet, on the snapshots a forget job of the plan
// would consider.
func (s *Server) previewRetention(c *gin.Context) {
	p, ok := s.plan(c)
	if !ok {
		return
	}
	var in RetentionDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	policy, err := in.policy()
	if err != nil {
		invalid(c, err)
		return
	}
	ctx := c.Request.Context()
	r, err := s.Store.GetRepository(ctx, p.RepositoryID)
	if err != nil {
		notFound(c, err, "repository")
		return
	}
	target, err := s.resticRepo(r)
	if err != nil {
		internal(c, err)
		return
	}
	opts := jobs.ForgetOptions(r, p, policy)
	opts.DryRun = true
	groups, err := s.Restic.Forget(ctx, target, opts)
	if err != nil {
		resticFailure(c, err)
		return
	}
	out := RetentionPreview{Keep: []KeptSnapshot{}, Remove: []restic.Snapshot{}}
	for _, g := range groups {
		matches := map[string][]string{}
		for _, reason := range g.Reasons {
			matches[reason.Snapshot.ID] = reason.Matches
		}
		for _, snap := range g.Keep {
			out.Keep = append(out.Keep, KeptSnapshot{Snapshot: snap, Matches: matches[snap.ID]})
		}
		out.Remove = append(out.Remove, g.Remove...)
	}
	sort.Slice(out.Keep, func(i, j int) bool { return out.Keep[i].Time.After(out.Keep[j].Time) })
	sort.Slice(out.Remove, func(i, j int) bool { return out.Remove[i].Time.After(out.Remove[j].Time) })
	c.JSON(http.StatusOK, out)
}
