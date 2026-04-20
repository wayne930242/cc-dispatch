package daemon

import ccdb "github.com/wayne930242/cc-dispatch/internal/db"

type DispatchStartRequest struct {
	Task      string `json:"task"`
	App       string `json:"app"`
	Cwd       string `json:"cwd"`
	Workspace string `json:"workspace,omitempty"`
}

type DispatchStartResponse struct {
	SessionID string             `json:"session_id"`
	Status    ccdb.SessionStatus `json:"status"`
	Cwd       string             `json:"cwd"`
	JsonlPath string             `json:"jsonl_path"`
	ResumeCmd string             `json:"resume_cmd"`
}

type DispatchListRequest struct {
	Workspace string `json:"workspace,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type DispatchListResponse struct {
	Sessions []ccdb.SessionSummary `json:"sessions"`
}

type DispatchStatusRequest struct {
	SessionID string `json:"session_id"`
}

type DispatchStatusResponse struct {
	ccdb.SessionRow
	ResumeCmd string `json:"resume_cmd"`
}

type DispatchTailRequest struct {
	SessionID string `json:"session_id"`
	Source    string `json:"source,omitempty"` // jsonl | stdout | stderr
	Lines     int    `json:"lines,omitempty"`
}

type DispatchTailResponse struct {
	Lines     []string `json:"lines"`
	Truncated bool     `json:"truncated"`
}

type DispatchCancelRequest struct {
	SessionID string `json:"session_id"`
}

type DispatchCancelResponse struct {
	Killed  bool   `json:"killed"`
	Message string `json:"message,omitempty"`
}

type errorResponse struct {
	Error  string `json:"error"`
	Detail any    `json:"detail,omitempty"`
}
