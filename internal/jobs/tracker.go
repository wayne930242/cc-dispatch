package jobs

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/wayne930242/cc-dispatch/internal/config"
	ccdb "github.com/wayne930242/cc-dispatch/internal/db"
)

// TickOnce scans running rows; dead PIDs transition to completed with exit_code=NULL.
// Exposed for tests; the real loop calls it every TrackerInterval seconds.
func TickOnce(db *sql.DB) {
	rows, err := ccdb.SelectRunning(db)
	if err != nil {
		slog.Error("tracker: SelectRunning failed", "err", err)
		return
	}
	now := time.Now().UnixMilli()
	msg := "daemon restarted or subprocess observed dead; exit code unavailable"
	for _, r := range rows {
		if r.PID == nil {
			continue
		}
		if pidAlive(int(*r.PID)) {
			continue
		}
		if err := ccdb.UpdateSessionExited(db, r.ID, ccdb.StatusCompleted, nil, now, &msg); err != nil {
			slog.Error("tracker: update failed", "id", r.ID, "err", err)
			continue
		}
		_ = ccdb.InsertEvent(db, r.ID, "exit", map[string]string{"reason": "pid_dead_on_tick"})
	}
}

type TrackerHandle struct {
	stop chan struct{}
}

func (h *TrackerHandle) Stop() { close(h.stop) }

func StartTracker(db *sql.DB) *TrackerHandle {
	h := &TrackerHandle{stop: make(chan struct{})}
	go func() {
		t := time.NewTicker(time.Duration(config.TrackerInterval) * time.Second)
		defer t.Stop()
		for {
			select {
			case <-h.stop:
				return
			case <-t.C:
				TickOnce(db)
			}
		}
	}()
	return h
}
