package jobs

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"syscall"
	"time"

	ccdb "github.com/wayne930242/cc-dispatch/internal/db"
)

type JobManager struct {
	db       *sql.DB
	mu       sync.Mutex
	children map[string]*exec.Cmd
}

func NewJobManager(db *sql.DB) *JobManager {
	return &JobManager{db: db, children: map[string]*exec.Cmd{}}
}

func (m *JobManager) Start(sessionID string) {
	row, err := ccdb.GetSession(m.db, sessionID)
	if err != nil || row == nil {
		slog.Error("manager.Start: session not found", "id", sessionID, "err", err)
		return
	}
	res, err := Spawn(SpawnInput{
		SessionID: row.ID,
		Task:      row.Task,
		Cwd:       row.Cwd,
	})
	if err != nil {
		msg := fmt.Sprintf("spawn failed: %v", err)
		_ = ccdb.UpdateSessionFailed(m.db, row.ID, msg, time.Now().UnixMilli())
		_ = ccdb.InsertEvent(m.db, row.ID, "exit", map[string]string{"error": msg})
		return
	}
	_ = ccdb.UpdateSessionSpawned(m.db, row.ID, int64(res.PID), time.Now().UnixMilli())
	_ = ccdb.InsertEvent(m.db, row.ID, "spawn", map[string]int{"pid": res.PID})

	m.mu.Lock()
	m.children[row.ID] = res.Cmd
	m.mu.Unlock()

	go m.waitForChild(row.ID, res.Cmd)
}

func (m *JobManager) waitForChild(sessionID string, cmd *exec.Cmd) {
	err := cmd.Wait()
	now := time.Now().UnixMilli()

	var status ccdb.SessionStatus
	var exitCode *int64
	var errMsg *string
	var signal string

	if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		signal = ws.Signal().String()
	}

	if signal == "terminated" || signal == "interrupt" {
		status = ccdb.StatusCancelled
	} else if err == nil {
		status = ccdb.StatusCompleted
		zero := int64(0)
		exitCode = &zero
	} else {
		status = ccdb.StatusFailed
		if exitErr, ok := err.(*exec.ExitError); ok {
			ec := int64(exitErr.ExitCode())
			exitCode = &ec
		}
		msg := fmt.Sprintf("subprocess exited: %v (signal=%q)", err, signal)
		errMsg = &msg
	}

	_ = ccdb.UpdateSessionExited(m.db, sessionID, status, exitCode, now, errMsg)
	_ = ccdb.InsertEvent(m.db, sessionID, "exit", map[string]any{
		"exit_code": exitCode,
		"signal":    signal,
	})

	m.mu.Lock()
	delete(m.children, sessionID)
	m.mu.Unlock()
}

// Cancel returns true if a SIGTERM was successfully sent (or the queued row
// was directly marked cancelled). Returns false if the session is not cancellable.
func (m *JobManager) Cancel(sessionID string) bool {
	row, err := ccdb.GetSession(m.db, sessionID)
	if err != nil || row == nil {
		return false
	}
	now := time.Now().UnixMilli()
	if row.Status == ccdb.StatusQueued {
		msg := "cancelled before spawn"
		_ = ccdb.UpdateSessionExited(m.db, sessionID, ccdb.StatusCancelled, nil, now, &msg)
		_ = ccdb.InsertEvent(m.db, sessionID, "cancel_requested", map[string]string{"at": "queued"})
		return true
	}
	if row.Status != ccdb.StatusRunning || row.PID == nil {
		return false
	}
	_ = ccdb.InsertEvent(m.db, sessionID, "cancel_requested", map[string]any{
		"at": "running", "pid": *row.PID,
	})

	m.mu.Lock()
	cmd := m.children[sessionID]
	m.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Signal(syscall.SIGTERM); err == nil {
			return true
		}
	}
	return false
}
