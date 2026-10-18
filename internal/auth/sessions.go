package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrNoSession is returned by Sessions.Get for an unknown token.
var ErrNoSession = errors.New("auth: no such session")

// Session is a login, stored under the SHA-256 of its token.
type Session struct {
	Principal string
	Created   time.Time
	LastSeen  time.Time
}

// Sessions stores the sessions, keyed by token hash.
type Sessions interface {
	Create(ctx context.Context, hash []byte, s Session) error
	Get(ctx context.Context, hash []byte) (Session, error)
	Touch(ctx context.Context, hash []byte, at time.Time) error
	Delete(ctx context.Context, hash []byte) error
	// Prune deletes the sessions created before createdBefore, unused since
	// seenBefore, or of another principal.
	Prune(ctx context.Context, principal string, createdBefore, seenBefore time.Time) error
}

// MemorySessions keeps sessions in memory, for tests.
type MemorySessions struct {
	mu       sync.Mutex
	sessions map[string]Session
}

// NewMemorySessions returns an empty MemorySessions.
func NewMemorySessions() *MemorySessions {
	return &MemorySessions{sessions: map[string]Session{}}
}

func (m *MemorySessions) Create(_ context.Context, hash []byte, s Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[string(hash)] = s
	return nil
}

func (m *MemorySessions) Get(_ context.Context, hash []byte) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[string(hash)]
	if !ok {
		return Session{}, ErrNoSession
	}
	return s, nil
}

func (m *MemorySessions) Touch(_ context.Context, hash []byte, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[string(hash)]; ok {
		s.LastSeen = at
		m.sessions[string(hash)] = s
	}
	return nil
}

func (m *MemorySessions) Delete(_ context.Context, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, string(hash))
	return nil
}

func (m *MemorySessions) Prune(_ context.Context, principal string, createdBefore, seenBefore time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, s := range m.sessions {
		if s.Principal != principal || s.Created.Before(createdBefore) || s.LastSeen.Before(seenBefore) {
			delete(m.sessions, k)
		}
	}
	return nil
}

// Len returns the number of stored sessions.
func (m *MemorySessions) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

// limiter slows password guessing: a client with maxFailures failed logins
// within window must wait until the oldest of them leaves the window.
type limiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
}

const (
	maxFailures = 10
	window      = 15 * time.Minute
	// maxClients bounds the memory an attacker cycling addresses can use
	maxClients = 10000
)

func newLimiter() *limiter { return &limiter{failures: map[string][]time.Time{}} }

// recent drops the failures older than the window; the lock is held.
func (l *limiter) recent(client string, now time.Time) []time.Time {
	kept := l.failures[client][:0]
	for _, t := range l.failures[client] {
		if now.Sub(t) < window {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.failures, client)
		return nil
	}
	l.failures[client] = kept
	return kept
}

// wait returns how long client must wait before trying again, 0 if it may.
func (l *limiter) wait(client string, now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	recent := l.recent(client, now)
	if len(recent) < maxFailures {
		return 0
	}
	return window - now.Sub(recent[0])
}

func (l *limiter) fail(client string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.failures) >= maxClients {
		for c := range l.failures {
			l.recent(c, now)
		}
		if len(l.failures) >= maxClients {
			l.failures = map[string][]time.Time{}
		}
	}
	l.failures[client] = append(l.recent(client, now), now)
}

func (l *limiter) reset(client string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, client)
}
