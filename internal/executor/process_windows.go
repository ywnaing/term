//go:build windows

package executor

import "os/exec"

func configureCommand(cmd *exec.Cmd) {}

func terminateCommand(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}
