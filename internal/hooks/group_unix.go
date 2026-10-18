//go:build unix

package hooks

import (
	"os/exec"
	"syscall"
)

// killGroup runs cmd in its own process group and makes cancelling it kill
// the whole group, not only the shell.
func killGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
