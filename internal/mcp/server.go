package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

func Run() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := EnsureDaemon(exe); err != nil {
		return fmt.Errorf("ensure daemon: %w", err)
	}
	c, err := client.FromConfigFile()
	if err != nil {
		return err
	}

	srv := mcpserver.NewMCPServer("cc-dispatch", "0.1.0",
		mcpserver.WithToolCapabilities(true),
	)

	// dispatch_start
	srv.AddTool(mcp.NewTool("dispatch_start",
		mcp.WithDescription("Start a headless Claude session for a given task in a given cwd. Returns session_id and resume command."),
		mcp.WithString("task", mcp.Required(), mcp.Description("Prompt text for the dispatched claude session")),
		mcp.WithString("app", mcp.Required(), mcp.Description("Free-form app label (e.g. rest-api-v3)")),
		mcp.WithString("cwd", mcp.Required(), mcp.Description("Absolute path where claude should run")),
		mcp.WithString("workspace", mcp.Description("Workspace tag (defaults to basename of cwd)")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		body := daemon.DispatchStartRequest{
			Task:      strArg(args, "task"),
			App:       strArg(args, "app"),
			Cwd:       strArg(args, "cwd"),
			Workspace: strArg(args, "workspace"),
		}
		var out daemon.DispatchStartResponse
		if err := c.RPC("dispatch_start", body, &out); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(formatStart(out)), nil
	})

	// dispatch_list
	srv.AddTool(mcp.NewTool("dispatch_list",
		mcp.WithDescription("List dispatched sessions with optional filters."),
		mcp.WithString("workspace"),
		mcp.WithString("status", mcp.Enum("queued", "running", "completed", "failed", "cancelled")),
		mcp.WithNumber("limit"),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		body := daemon.DispatchListRequest{
			Workspace: strArg(args, "workspace"),
			Status:    strArg(args, "status"),
			Limit:     intArg(args, "limit"),
		}
		var out daemon.DispatchListResponse
		if err := c.RPC("dispatch_list", body, &out); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(formatList(out)), nil
	})

	// dispatch_status
	srv.AddTool(mcp.NewTool("dispatch_status",
		mcp.WithDescription("Get full detail for a session including the shell command to resume it interactively."),
		mcp.WithString("session_id", mcp.Required()),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		body := daemon.DispatchStatusRequest{SessionID: strArg(args, "session_id")}
		var out daemon.DispatchStatusResponse
		if err := c.RPC("dispatch_status", body, &out); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(formatStatus(out)), nil
	})

	// dispatch_tail
	srv.AddTool(mcp.NewTool("dispatch_tail",
		mcp.WithDescription("Read the last N lines of a session's jsonl transcript, stdout, or stderr log."),
		mcp.WithString("session_id", mcp.Required()),
		mcp.WithString("source", mcp.Enum("jsonl", "stdout", "stderr")),
		mcp.WithNumber("lines"),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		body := daemon.DispatchTailRequest{
			SessionID: strArg(args, "session_id"),
			Source:    strArg(args, "source"),
			Lines:     intArg(args, "lines"),
		}
		var out daemon.DispatchTailResponse
		if err := c.RPC("dispatch_tail", body, &out); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(formatTail(out)), nil
	})

	// dispatch_cancel
	srv.AddTool(mcp.NewTool("dispatch_cancel",
		mcp.WithDescription("Cancel a queued or running session (SIGTERM)."),
		mcp.WithString("session_id", mcp.Required()),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		body := daemon.DispatchCancelRequest{SessionID: strArg(args, "session_id")}
		var out daemon.DispatchCancelResponse
		if err := c.RPC("dispatch_cancel", body, &out); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	return mcpserver.ServeStdio(srv)
}

func strArg(a map[string]any, key string) string {
	if v, ok := a[key].(string); ok {
		return v
	}
	return ""
}

func intArg(a map[string]any, key string) int {
	switch v := a[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func formatStart(r daemon.DispatchStartResponse) string {
	return fmt.Sprintf("Session started: %s\nStatus: %s\nWorking dir: %s\n\nResume interactively:\n```\n%s\n```",
		r.SessionID, r.Status, r.Cwd, r.ResumeCmd)
}

func formatList(r daemon.DispatchListResponse) string {
	if len(r.Sessions) == 0 {
		return "(no sessions)"
	}
	lines := []string{"| session | workspace | app | status | created |", "|---|---|---|---|---|"}
	for _, s := range r.Sessions {
		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |",
			short(s.ID), s.Workspace, s.App, s.Status,
			time.UnixMilli(s.CreatedAt).UTC().Format(time.RFC3339)))
	}
	return strings.Join(lines, "\n")
}

func short(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func formatStatus(r daemon.DispatchStatusResponse) string {
	cp := r
	cp.ResumeCmd = ""
	b, _ := json.MarshalIndent(cp, "", "  ")
	return fmt.Sprintf("```json\n%s\n```\n\nResume:\n```\n%s\n```", string(b), r.ResumeCmd)
}

func formatTail(r daemon.DispatchTailResponse) string {
	header := fmt.Sprintf("(full tail, %d lines)", len(r.Lines))
	if r.Truncated {
		header = fmt.Sprintf("(showing last %d lines, older lines omitted)", len(r.Lines))
	}
	return header + "\n```\n" + strings.Join(r.Lines, "\n") + "\n```"
}
