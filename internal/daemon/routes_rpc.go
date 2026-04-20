package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wayne930242/cc-dispatch/internal/config"
	ccdb "github.com/wayne930242/cc-dispatch/internal/db"
)

func resumeCmd(cwd, sessionID string) string {
	q := strings.ReplaceAll(cwd, "'", `'\''`)
	return fmt.Sprintf("cd '%s' && claude --resume %s", q, sessionID)
}

func (s *Server) handleDispatchStart(w http.ResponseWriter, r *http.Request) {
	var req DispatchStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_request", Detail: err.Error()})
		return
	}
	if req.Task == "" || len(req.Task) > config.MaxTaskBytes {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "task_empty_or_too_large"})
		return
	}
	if req.App == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "app_required"})
		return
	}
	if !filepath.IsAbs(req.Cwd) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cwd_must_be_absolute"})
		return
	}
	if fi, err := os.Stat(req.Cwd); err != nil || !fi.IsDir() {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cwd_not_found"})
		return
	}

	ws := req.Workspace
	if ws == "" {
		ws = filepath.Base(req.Cwd)
	}
	id := uuid.NewString()
	jsonlPath := config.JsonlPathFor(req.Cwd, id)

	if err := ccdb.InsertSession(s.DB, ccdb.InsertSessionInput{
		ID: id, Workspace: ws, App: req.App, Task: req.Task,
		Cwd: req.Cwd, Status: ccdb.StatusQueued,
		JsonlPath:  jsonlPath,
		StdoutPath: config.StdoutLogPath(id),
		StderrPath: config.StderrLogPath(id),
		CreatedAt:  time.Now().UnixMilli(),
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	go s.Mgr.Start(id)

	writeJSON(w, http.StatusOK, DispatchStartResponse{
		SessionID: id,
		Status:    ccdb.StatusQueued,
		Cwd:       req.Cwd,
		JsonlPath: jsonlPath,
		ResumeCmd: resumeCmd(req.Cwd, id),
	})
}

func (s *Server) handleDispatchList(w http.ResponseWriter, r *http.Request) {
	var req DispatchListRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	rows, err := ccdb.ListSessions(s.DB, ccdb.ListOpts{
		Workspace: req.Workspace,
		Status:    ccdb.SessionStatus(req.Status),
		Limit:     req.Limit,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if rows == nil {
		rows = []ccdb.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, DispatchListResponse{Sessions: rows})
}

func (s *Server) handleDispatchStatus(w http.ResponseWriter, r *http.Request) {
	var req DispatchStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	row, err := ccdb.GetSession(s.DB, req.SessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if row == nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, DispatchStatusResponse{
		SessionRow: *row,
		ResumeCmd:  resumeCmd(row.Cwd, row.ID),
	})
}

func readLastLines(path string, maxLines int) (lines []string, truncated bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, false, nil
		}
		return nil, false, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, false, err
	}
	all := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	if len(all) > 0 && all[len(all)-1] == "" {
		all = all[:len(all)-1]
	}
	if len(all) > maxLines {
		return all[len(all)-maxLines:], true, nil
	}
	return all, false, nil
}

func (s *Server) handleDispatchTail(w http.ResponseWriter, r *http.Request) {
	var req DispatchTailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	if req.Source == "" {
		req.Source = "jsonl"
	}
	if req.Lines <= 0 {
		req.Lines = 50
	}
	if req.Lines > 1000 {
		req.Lines = 1000
	}
	row, err := ccdb.GetSession(s.DB, req.SessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if row == nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}
	var path string
	switch req.Source {
	case "jsonl":
		if row.JsonlPath != nil {
			path = *row.JsonlPath
		}
	case "stdout":
		if row.StdoutPath != nil {
			path = *row.StdoutPath
		}
	case "stderr":
		if row.StderrPath != nil {
			path = *row.StderrPath
		}
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_source"})
		return
	}
	if path == "" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "path_not_set"})
		return
	}
	lines, truncated, err := readLastLines(path, req.Lines)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, DispatchTailResponse{Lines: lines, Truncated: truncated})
}

func (s *Server) handleDispatchCancel(w http.ResponseWriter, r *http.Request) {
	var req DispatchCancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	row, err := ccdb.GetSession(s.DB, req.SessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if row == nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}
	killed := s.Mgr.Cancel(req.SessionID)
	msg := "SIGTERM sent; status will transition to cancelled on next tick"
	if !killed {
		msg = "session not cancellable (not running or already ended)"
	}
	writeJSON(w, http.StatusOK, DispatchCancelResponse{Killed: killed, Message: msg})
}
