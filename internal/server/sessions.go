package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mathieuHa/resticgui/internal/auth"
	"github.com/mathieuHa/resticgui/internal/store"
	"github.com/mathieuHa/resticgui/internal/store/db"
)

// storeSessions keeps the login sessions in the database, so that they
// survive a restart and a logout ends them for good.
type storeSessions struct{ st *store.Store }

func (s storeSessions) Create(ctx context.Context, hash []byte, se auth.Session) error {
	return s.st.CreateSession(ctx, db.CreateSessionParams{
		TokenHash: hash, Principal: se.Principal, CreatedAt: se.Created.Unix(), LastSeenAt: se.LastSeen.Unix(),
	})
}

func (s storeSessions) Get(ctx context.Context, hash []byte) (auth.Session, error) {
	row, err := s.st.GetSession(ctx, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.Session{}, auth.ErrNoSession
	}
	if err != nil {
		return auth.Session{}, err
	}
	return auth.Session{Principal: row.Principal, Created: time.Unix(row.CreatedAt, 0), LastSeen: time.Unix(row.LastSeenAt, 0)}, nil
}

func (s storeSessions) Touch(ctx context.Context, hash []byte, at time.Time) error {
	return s.st.TouchSession(ctx, db.TouchSessionParams{LastSeenAt: at.Unix(), TokenHash: hash})
}

func (s storeSessions) Delete(ctx context.Context, hash []byte) error {
	return s.st.DeleteSession(ctx, hash)
}

func (s storeSessions) Prune(ctx context.Context, principal string, createdBefore, seenBefore time.Time) error {
	_, err := s.st.DeleteStaleSessions(ctx, db.DeleteStaleSessionsParams{
		CreatedBefore: createdBefore.Unix(), SeenBefore: seenBefore.Unix(), Principal: principal,
	})
	return err
}
