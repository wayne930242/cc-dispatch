package daemon

import (
	"encoding/json"
	"net/http"

	"github.com/wayne930242/cc-dispatch/internal/config"
)

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"version": config.DaemonVersion,
	})
}
