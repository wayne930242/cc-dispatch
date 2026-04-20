package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/wayne930242/cc-dispatch/internal/auth"
	"github.com/wayne930242/cc-dispatch/internal/config"
)

type Client struct {
	BaseURL string
	Token   string
	http    *http.Client
}

func FromConfigFile() (*Client, error) {
	data, err := os.ReadFile(config.ConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg auth.DaemonConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return New(cfg.Port, cfg.Token), nil
}

func New(port int, token string) *Client {
	return &Client{
		BaseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		Token:   token,
		http:    &http.Client{},
	}
}

type HealthResponse struct {
	OK      bool   `json:"ok"`
	Version string `json:"version"`
}

func (c *Client) Health() (*HealthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.HealthTimeout)*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/health", nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health %d", resp.StatusCode)
	}
	var out HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RPC POSTs body to /rpc/<name>, decodes response into out (or skips decode if out is nil).
func (c *Client) RPC(name string, body, out any) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPost, c.BaseURL+"/rpc/"+name, buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rpc %s: %d %s", name, resp.StatusCode, string(b))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
