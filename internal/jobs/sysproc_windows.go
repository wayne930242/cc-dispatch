//go:build windows

package jobs

import (
	"os/exec"
	"syscall"
)

func configureSysProcAttr(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = 0x00000200 // CREATE_NEW_PROCESS_GROUP
	cmd.SysProcAttr.HideWindow = true
}
