//go:build !unix

package hooks

import "os/exec"

// killGroup leaves cmd as it is: only the command itself is killed.
func killGroup(*exec.Cmd) {}
