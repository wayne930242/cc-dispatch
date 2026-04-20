package auth

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOrCreateConfig(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "config.json")

	cfg1, err := LoadOrCreateConfig(p)
	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`^[a-f0-9]{64}$`), cfg1.Token)
	require.Equal(t, 47821, cfg1.Port)

	cfg2, err := LoadOrCreateConfig(p)
	require.NoError(t, err)
	require.Equal(t, cfg1.Token, cfg2.Token)
}

func TestVerifyToken(t *testing.T) {
	tmp := t.TempDir()
	cfg, err := LoadOrCreateConfig(filepath.Join(tmp, "config.json"))
	require.NoError(t, err)

	require.True(t, VerifyToken("Bearer "+cfg.Token, cfg.Token))
	require.False(t, VerifyToken("Bearer wrong", cfg.Token))
	require.False(t, VerifyToken("", cfg.Token))
	require.False(t, VerifyToken(cfg.Token, cfg.Token)) // no Bearer prefix
}
