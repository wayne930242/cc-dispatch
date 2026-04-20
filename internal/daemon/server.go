package daemon

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wayne930242/cc-dispatch/internal/auth"
	"github.com/wayne930242/cc-dispatch/internal/config"
	ccdb "github.com/wayne930242/cc-dispatch/internal/db"
	"github.com/wayne930242/cc-dispatch/internal/jobs"
)

type Server struct {
	DB    *sql.DB
	Token string
	Mgr   *jobs.JobManager
	srv   *http.Server
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next(w, r)
			return
		}
		if !auth.VerifyToken(r.Header.Get("Authorization"), s.Token) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		next(w, r)
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /rpc/dispatch_start", s.handleDispatchStart)
	mux.HandleFunc("POST /rpc/dispatch_list", s.handleDispatchList)
	mux.HandleFunc("POST /rpc/dispatch_status", s.handleDispatchStatus)
	mux.HandleFunc("POST /rpc/dispatch_tail", s.handleDispatchTail)
	mux.HandleFunc("POST /rpc/dispatch_cancel", s.handleDispatchCancel)

	// Wrap every handler with auth middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.authMiddleware(mux.ServeHTTP)(w, r)
	})
}

func (s *Server) Serve() error {
	port := config.DefaultPort
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := s.writePID(); err != nil {
		return err
	}
	defer s.removePID()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	tracker := jobs.StartTracker(s.DB)
	defer tracker.Stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("daemon listening", "addr", addr)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-stop:
		slog.Info("shutdown signal received")
	case err := <-errCh:
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) writePID() error {
	return os.WriteFile(config.DaemonPIDPath(), []byte(fmt.Sprintf("%d", os.Getpid())), 0o600)
}

func (s *Server) removePID() {
	_ = os.Remove(config.DaemonPIDPath())
}

func NewFromEnv() (*Server, error) {
	if err := os.MkdirAll(config.LogsDir(), 0o700); err != nil {
		return nil, err
	}
	cfg, err := auth.LoadOrCreateConfig(config.ConfigPath())
	if err != nil {
		return nil, err
	}
	db, err := ccdb.Open(config.DBPath())
	if err != nil {
		return nil, err
	}
	return &Server{
		DB:    db,
		Token: cfg.Token,
		Mgr:   jobs.NewJobManager(db),
	}, nil
}

// Placeholder handlers — real implementations land in MG-11.
func (s *Server) handleDispatchStart(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (s *Server) handleDispatchList(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (s *Server) handleDispatchStatus(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (s *Server) handleDispatchTail(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (s *Server) handleDispatchCancel(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
