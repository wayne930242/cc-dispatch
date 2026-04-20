package jobs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	ccdb "github.com/wayne930242/cc-dispatch/internal/db"
)

func TestTickOnceDeadPID(t *testing.T) {
	db, err := ccdb.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, ccdb.InsertSession(db, ccdb.InsertSessionInput{
		ID: "s1", Workspace: "w", App: "a", Task: "t", Cwd: "/",
		Status: ccdb.StatusQueued, CreatedAt: 1,
	}))
	// A PID that is extremely unlikely to exist on this machine.
	require.NoError(t, ccdb.UpdateSessionSpawned(db, "s1", 999999, 2))

	TickOnce(db)

	row, err := ccdb.GetSession(db, "s1")
	require.NoError(t, err)
	require.Equal(t, ccdb.StatusCompleted, row.Status)
	require.Nil(t, row.ExitCode)
	require.NotNil(t, row.ErrorMessage)
	require.Contains(t, *row.ErrorMessage, "exit code unavailable")
}

func TestTickOnceLivePID(t *testing.T) {
	db, err := ccdb.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, ccdb.InsertSession(db, ccdb.InsertSessionInput{
		ID: "s2", Workspace: "w", App: "a", Task: "t", Cwd: "/",
		Status: ccdb.StatusQueued, CreatedAt: 1,
	}))
	require.NoError(t, ccdb.UpdateSessionSpawned(db, "s2", int64(os.Getpid()), 2))

	TickOnce(db)

	row, err := ccdb.GetSession(db, "s2")
	require.NoError(t, err)
	require.Equal(t, ccdb.StatusRunning, row.Status)
}
