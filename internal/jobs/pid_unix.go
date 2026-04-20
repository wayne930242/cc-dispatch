//go:build !windows

package jobs

import "syscall"

func pidAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
