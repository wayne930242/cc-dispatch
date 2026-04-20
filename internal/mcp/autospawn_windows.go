//go:build windows

package mcp

import (
	"os/exec"
	"syscall"
)

func configureDetach(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = 0x00000200
	cmd.SysProcAttr.HideWindow = true
}
