package db

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func helperDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestInsertAndGet(t *testing.T) {
	db := helperDB(t)
	require.NoError(t, InsertSession(db, InsertSessionInput{
		ID: "s1", Workspace: "moldplan-center", App: "rest-api-v3",
		Task: "fix bug", Cwd: "/tmp/x", Status: StatusQueued,
		JsonlPath: "/a.jsonl", StdoutPath: "/a.out", StderrPath: "/a.err",
		CreatedAt: 1000,
	}))
	row, err := GetSession(db, "s1")
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, "s1", row.ID)
	require.Equal(t, StatusQueued, row.Status)
	require.Nil(t, row.PID)
}

func TestUpdateSpawned(t *testing.T) {
	db := helperDB(t)
	require.NoError(t, InsertSession(db, InsertSessionInput{
		ID: "s2", Workspace: "w", App: "a", Task: "t", Cwd: "/",
		Status: StatusQueued, CreatedAt: 1,
	}))
	require.NoError(t, UpdateSessionSpawned(db, "s2", 12345, 2000))
	row, err := GetSession(db, "s2")
	require.NoError(t, err)
	require.NotNil(t, row.PID)
	require.Equal(t, int64(12345), *row.PID)
	require.Equal(t, StatusRunning, row.Status)
}

func TestSelectRunning(t *testing.T) {
	db := helperDB(t)
	for _, pair := range []struct {
		id     string
		status SessionStatus
	}{
		{"a", StatusQueued}, {"b", StatusRunning}, {"c", StatusCompleted},
	} {
		require.NoError(t, InsertSession(db, InsertSessionInput{
			ID: pair.id, Workspace: "w", App: "a", Task: "t", Cwd: "/",
			Status: pair.status, CreatedAt: 1,
		}))
	}
	rows, err := SelectRunning(db)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "b", rows[0].ID)
}

func TestListWithFilter(t *testing.T) {
	db := helperDB(t)
	for _, pair := range []struct {
		id, ws string
	}{{"x", "aa"}, {"y", "bb"}} {
		require.NoError(t, InsertSession(db, InsertSessionInput{
			ID: pair.id, Workspace: pair.ws, App: "a", Task: "t", Cwd: "/",
			Status: StatusQueued, CreatedAt: 1,
		}))
	}
	rows, err := ListSessions(db, ListOpts{Workspace: "bb", Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "y", rows[0].ID)
}

func TestUpdateExited(t *testing.T) {
	db := helperDB(t)
	require.NoError(t, InsertSession(db, InsertSessionInput{
		ID: "z", Workspace: "w", App: "a", Task: "t", Cwd: "/",
		Status: StatusRunning, CreatedAt: 1,
	}))
	zero := int64(0)
	require.NoError(t, UpdateSessionExited(db, "z", StatusCompleted, &zero, 3000, nil))
	row, err := GetSession(db, "z")
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, row.Status)
	require.NotNil(t, row.ExitCode)
	require.Equal(t, int64(0), *row.ExitCode)
}
