package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wayne930242/cc-dispatch/internal/config"
)

type DaemonConfig struct {
	Port      int    `json:"port"`
	Token     string `json:"token"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}

func LoadOrCreateConfig(configPath string) (*DaemonConfig, error) {
	if data, err := os.ReadFile(configPath); err == nil {
		var c DaemonConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("parse %s: %w", configPath, err)
		}
		return &c, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	c := DaemonConfig{
		Port:      config.DefaultPort,
		Token:     hex.EncodeToString(tokenBytes),
		Version:   config.DaemonVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return nil, err
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return nil, err
	}
	return &c, nil
}

// VerifyToken returns true iff the Authorization header is "Bearer <token>"
// matching expected. Uses length-check + constant-time byte compare to avoid
// length-oracle timing.
func VerifyToken(authHeader, expected string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return false
	}
	provided := authHeader[len(prefix):]
	pb := []byte(provided)
	eb := []byte(expected)
	if len(pb) != len(eb) {
		return false
	}
	return subtle.ConstantTimeCompare(pb, eb) == 1
}
