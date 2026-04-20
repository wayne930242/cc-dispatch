// Simulates `claude -p <task> --session-id <id> --output-format stream-json ...`
// for test use only. Honors env:
//   FAKE_CLAUDE_SLEEP_MS  (default 100)
//   FAKE_CLAUDE_EXIT_CODE (default 0)
//   FAKE_CLAUDE_STDERR    (default empty)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func findSessionID(args []string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--session-id" {
			return args[i+1]
		}
	}
	return ""
}

func main() {
	sleep := envInt("FAKE_CLAUDE_SLEEP_MS", 100)
	exitCode := envInt("FAKE_CLAUDE_EXIT_CODE", 0)
	errText := os.Getenv("FAKE_CLAUDE_STDERR")

	sid := findSessionID(os.Args[1:])
	start, _ := json.Marshal(map[string]string{"type": "start", "session_id": sid})
	fmt.Println(string(start))

	time.Sleep(time.Duration(sleep) * time.Millisecond)

	end, _ := json.Marshal(map[string]string{"type": "end", "session_id": sid})
	fmt.Println(string(end))

	if errText != "" {
		fmt.Fprintln(os.Stderr, errText)
	}
	os.Exit(exitCode)
}
