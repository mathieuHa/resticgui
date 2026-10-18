package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mathieuHa/resticgui/internal/hooks"
	"github.com/mathieuHa/resticgui/internal/notify"
	"github.com/mathieuHa/resticgui/internal/restic"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// SettingsDTO says which optional features the configuration turns on.
type SettingsDTO struct {
	// HookCommands: command hooks may be set (hooks.commands).
	HookCommands bool `json:"hook_commands"`
	// Metrics: /metrics is served (metrics.enabled).
	Metrics bool `json:"metrics"`
}

func (s *Server) settings(c *gin.Context) {
	c.JSON(http.StatusOK, SettingsDTO{HookCommands: s.HookCommands, Metrics: s.Metrics})
}

// NotifierDTO is a notifier as the API returns it: never its URL, which
// carries tokens, only the service it names.
type NotifierDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Service   string    `json:"service"`
	OnFailure bool      `json:"on_failure"`
	OnSuccess bool      `json:"on_success"`
	OnMissed  bool      `json:"on_missed"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// NotifierInput creates or changes a notifier.
type NotifierInput struct {
	Name string `json:"name"`
	// URL is a Shoutrrr URL, required on create; on update, empty keeps it.
	URL       string `json:"url"`
	OnFailure bool   `json:"on_failure"`
	OnSuccess bool   `json:"on_success"`
	OnMissed  bool   `json:"on_missed"`
	Enabled   bool   `json:"enabled"`
}

func notifierDTO(n db.Notifier) NotifierDTO {
	return NotifierDTO{ID: n.ID, Name: n.Name, Service: n.Service, OnFailure: n.OnFailure, OnSuccess: n.OnSuccess, OnMissed: n.OnMissed, Enabled: n.Enabled, CreatedAt: n.CreatedAt}
}

func notifierID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "invalid notifier id")
		return 0, false
	}
	return id, true
}

// bindNotifier reads a notifier and checks its URL, returning the sealed URL
// and its service when one is given.
func (s *Server) bindNotifier(c *gin.Context, creating bool) (NotifierInput, []byte, string, bool) {
	var in NotifierInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return in, nil, "", false
	}
	in.Name = strings.TrimSpace(in.Name)
	var errs []error
	if in.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	var service string
	switch {
	case in.URL != "":
		var err error
		if service, err = notify.ServiceName(in.URL); err != nil {
			errs = append(errs, fmt.Errorf("url: %w", err))
		}
	case creating:
		errs = append(errs, errors.New("url is required"))
	}
	if err := errors.Join(errs...); err != nil {
		invalid(c, err)
		return in, nil, "", false
	}
	if in.URL == "" {
		return in, nil, "", true
	}
	sealed, err := s.Box.Seal([]byte(in.URL))
	if err != nil {
		internal(c, err)
		return in, nil, "", false
	}
	return in, sealed, service, true
}

func (s *Server) listNotifiers(c *gin.Context) {
	list, err := s.Store.ListNotifiers(c.Request.Context())
	if err != nil {
		internal(c, err)
		return
	}
	out := make([]NotifierDTO, 0, len(list))
	for _, n := range list {
		out = append(out, notifierDTO(n))
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) createNotifier(c *gin.Context) {
	in, sealed, service, ok := s.bindNotifier(c, true)
	if !ok {
		return
	}
	n, err := s.Store.CreateNotifier(c.Request.Context(), db.CreateNotifierParams{
		Name: in.Name, Url: sealed, Service: service, OnFailure: in.OnFailure, OnSuccess: in.OnSuccess, OnMissed: in.OnMissed, Enabled: in.Enabled,
	})
	if isUniqueViolation(err) {
		fail(c, http.StatusConflict, fmt.Sprintf("a notifier named %q already exists", in.Name))
		return
	}
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, notifierDTO(n))
}

func (s *Server) updateNotifier(c *gin.Context) {
	id, ok := notifierID(c)
	if !ok {
		return
	}
	in, sealed, service, ok := s.bindNotifier(c, false)
	if !ok {
		return
	}
	var n db.Notifier
	err := s.Store.InTx(c.Request.Context(), func(q *db.Queries) error {
		var err error
		n, err = q.UpdateNotifier(c.Request.Context(), db.UpdateNotifierParams{
			ID: id, Name: in.Name, OnFailure: in.OnFailure, OnSuccess: in.OnSuccess, OnMissed: in.OnMissed, Enabled: in.Enabled,
		})
		if err != nil || sealed == nil {
			return err
		}
		if _, err := q.UpdateNotifierURL(c.Request.Context(), db.UpdateNotifierURLParams{ID: id, Url: sealed, Service: service}); err != nil {
			return err
		}
		n.Service = service
		return nil
	})
	if isUniqueViolation(err) {
		fail(c, http.StatusConflict, fmt.Sprintf("a notifier named %q already exists", in.Name))
		return
	}
	if err != nil {
		notFound(c, err, "notifier")
		return
	}
	c.JSON(http.StatusOK, notifierDTO(n))
}

func (s *Server) deleteNotifier(c *gin.Context) {
	id, ok := notifierID(c)
	if !ok {
		return
	}
	n, err := s.Store.DeleteNotifier(c.Request.Context(), id)
	if err != nil {
		internal(c, err)
		return
	}
	if n == 0 {
		fail(c, http.StatusNotFound, "notifier not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// testNotifier sends a test message, so that a URL can be checked at once.
func (s *Server) testNotifier(c *gin.Context) {
	id, ok := notifierID(c)
	if !ok {
		return
	}
	n, err := s.Store.GetNotifier(c.Request.Context(), id)
	if err != nil {
		notFound(c, err, "notifier")
		return
	}
	if err := s.Notify.SendTo(n, "🔔 ResticGUI test", "This is a test from ResticGUI. Alerts about your backups will arrive here."); err != nil {
		fail(c, http.StatusBadGateway, "could not send: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// HealthcheckDTO is a repository's Healthchecks check: its host only, as
// the ping URL is a secret.
type HealthcheckDTO struct {
	Host string `json:"host"`
}

// HealthcheckInput sets a repository's Healthchecks ping URL.
type HealthcheckInput struct {
	URL string `json:"url"`
}

func httpURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, errors.New("an http or https URL such as https://hc-ping.com/your-uuid")
	}
	return u, nil
}

func (s *Server) putHealthcheck(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in HealthcheckInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	u, err := httpURL(strings.TrimSpace(in.URL))
	if err != nil {
		invalid(c, fmt.Errorf("url is %w", err))
		return
	}
	if _, err := s.Store.GetRepository(c.Request.Context(), id); err != nil {
		notFound(c, err, "repository")
		return
	}
	sealed, err := s.Box.Seal([]byte(u.String()))
	if err != nil {
		internal(c, err)
		return
	}
	if err := s.Store.UpsertHealthcheck(c.Request.Context(), db.UpsertHealthcheckParams{RepositoryID: id, Url: sealed, Host: u.Hostname()}); err != nil {
		internal(c, err)
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

func (s *Server) deleteHealthcheck(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	if _, err := s.Store.DeleteHealthcheck(c.Request.Context(), id); err != nil {
		internal(c, err)
		return
	}
	s.respondRepository(c, http.StatusOK, id)
}

// HookDTO is a hook as the API returns it: a webhook shows its host only,
// as its URL can carry a token.
type HookDTO struct {
	ID             int64  `json:"id"`
	Event          string `json:"event"`
	Kind           string `json:"kind"`
	Command        string `json:"command"`
	Host           string `json:"host"`
	TimeoutSeconds int64  `json:"timeout_seconds"`
}

// HookInput is one hook of PUT /repositories/{id}/hooks.
type HookInput struct {
	// ID keeps the URL of an existing webhook when URL is empty.
	ID             int64  `json:"id"`
	Event          string `json:"event"`
	Kind           string `json:"kind"`
	Command        string `json:"command"`
	URL            string `json:"url"`
	TimeoutSeconds int64  `json:"timeout_seconds"`
}

var hookEvents = map[string]bool{hooks.Before: true, hooks.Success: true, hooks.Failure: true}

// putHooks replaces the hooks of a repository; they run in the order given.
func (s *Server) putHooks(c *gin.Context) {
	id, ok := repositoryID(c)
	if !ok {
		return
	}
	var in []HookInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	ctx := c.Request.Context()
	if _, err := s.Store.GetRepository(ctx, id); err != nil {
		notFound(c, err, "repository")
		return
	}
	current, err := s.Store.ListHooks(ctx, id)
	if err != nil {
		internal(c, err)
		return
	}
	existing := map[int64]db.Hook{}
	for _, h := range current {
		existing[h.ID] = h
	}

	params := make([]db.CreateHookParams, 0, len(in))
	var errs []error
	for i, h := range in {
		p := db.CreateHookParams{RepositoryID: id, Event: h.Event, Kind: h.Kind, TimeoutSeconds: h.TimeoutSeconds, Position: int64(i)}
		if p.TimeoutSeconds == 0 {
			p.TimeoutSeconds = 300
		}
		if !hookEvents[h.Event] {
			errs = append(errs, fmt.Errorf("hook %d: event is one of before, success, failure", i+1))
		}
		if p.TimeoutSeconds < 1 || p.TimeoutSeconds > 24*3600 {
			errs = append(errs, fmt.Errorf("hook %d: timeout_seconds is between 1 and 86400", i+1))
		}
		switch h.Kind {
		case hooks.Command:
			p.Command = strings.TrimSpace(h.Command)
			if p.Command == "" {
				errs = append(errs, fmt.Errorf("hook %d: command is required", i+1))
			}
			if !s.HookCommands {
				errs = append(errs, fmt.Errorf("hook %d: %w", i+1, hooks.ErrCommandsDisabled))
			}
		case hooks.Webhook:
			if h.URL == "" {
				if old, ok := existing[h.ID]; ok && old.Kind == hooks.Webhook {
					p.Url, p.Host = old.Url, old.Host
					break
				}
				errs = append(errs, fmt.Errorf("hook %d: url is required", i+1))
				break
			}
			u, err := httpURL(strings.TrimSpace(h.URL))
			if err != nil {
				errs = append(errs, fmt.Errorf("hook %d: url is %w", i+1, err))
				break
			}
			if p.Url, err = s.Box.Seal([]byte(u.String())); err != nil {
				internal(c, err)
				return
			}
			p.Host = u.Hostname()
		default:
			errs = append(errs, fmt.Errorf("hook %d: kind is command or webhook", i+1))
		}
		params = append(params, p)
	}
	if err := errors.Join(errs...); err != nil {
		invalid(c, err)
		return
	}
	err = s.Store.InTx(ctx, func(q *db.Queries) error {
		if err := q.DeleteHooks(ctx, id); err != nil {
			return err
		}
		for _, p := range params {
			if _, err := q.CreateHook(ctx, p); err != nil {
				return err
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

// monitoringDTO adds a repository's Healthchecks check and hooks to dto.
func (s *Server) monitoringDTO(ctx context.Context, dto *RepositoryDTO) error {
	hc, err := s.Store.GetHealthcheck(ctx, dto.ID)
	switch {
	case err == nil:
		dto.Healthcheck = &HealthcheckDTO{Host: hc.Host}
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}
	list, err := s.Store.ListHooks(ctx, dto.ID)
	if err != nil {
		return err
	}
	dto.Hooks = make([]HookDTO, 0, len(list))
	// stored by event then position; shown in the order they run
	order := map[string]int{hooks.Before: 0, hooks.Success: 1, hooks.Failure: 2}
	sort.SliceStable(list, func(i, j int) bool { return order[list[i].Event] < order[list[j].Event] })
	for _, h := range list {
		dto.Hooks = append(dto.Hooks, HookDTO{ID: h.ID, Event: h.Event, Kind: h.Kind, Command: h.Command, Host: h.Host, TimeoutSeconds: h.TimeoutSeconds})
	}
	return nil
}

// ProbeDTO says what is at a repository location, before it is added.
type ProbeDTO struct {
	// Exists: a restic repository is there and the password opens it.
	Exists    bool `json:"exists"`
	Snapshots int  `json:"snapshots"`
	// Latest is the newest snapshot, whose host and paths can be reused.
	Latest *restic.Snapshot `json:"latest"`
}

// probe looks at a repository location with the settings of a repository
// not added yet: whether a repository exists there, and its snapshots.
func (s *Server) probe(c *gin.Context) {
	var in RepositoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		invalid(c, err)
		return
	}
	secretEnv := map[string]string{}
	if in.Secrets != nil {
		secretEnv = *in.Secrets
	}
	if err := in.validate(secretEnv, true, s.Roots); err != nil {
		invalid(c, err)
		return
	}
	env := map[string]string{}
	for k, v := range in.Env {
		env[k] = v
	}
	for k, v := range secretEnv {
		env[k] = v
	}
	target := &restic.Repo{Location: in.Location, Password: in.Password, Env: env, CACert: in.CACert, Options: in.Extended}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()
	exists, err := s.Restic.Exists(ctx, target)
	if errors.Is(err, restic.ErrWrongPassword) {
		fail(c, http.StatusConflict, "a repository exists there, but the password does not open it")
		return
	}
	if err != nil {
		resticFailure(c, err)
		return
	}
	out := ProbeDTO{Exists: exists}
	if exists {
		snaps, err := s.Restic.Snapshots(ctx, target, restic.SnapshotFilter{})
		if err != nil {
			resticFailure(c, err)
			return
		}
		out.Snapshots = len(snaps)
		if len(snaps) > 0 {
			out.Latest = &snaps[len(snaps)-1]
		}
	}
	c.JSON(http.StatusOK, out)
}

// diff lists what changed from one snapshot to another (?to=).
func (s *Server) diff(c *gin.Context) {
	_, target, from, ok := s.snapshotTarget(c)
	if !ok {
		return
	}
	to := c.Query("to")
	if !snapshotPattern.MatchString(to) {
		fail(c, http.StatusBadRequest, "to is a snapshot id")
		return
	}
	d, err := s.Restic.Diff(c.Request.Context(), target, from, to)
	if err != nil {
		resticFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}
