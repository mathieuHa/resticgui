package commands

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"

	"github.com/rs/zerolog/log"
)

var (
	RESTIC_PATH string
)

func ResticInit() {
	if runtime.GOOS == "windows" {
		RESTIC_PATH = "restic.exe"
	}

	if runtime.GOOS == "linux" {
		RESTIC_PATH = "restic"
	}
}

func runRestic(args []string, env []string) string {
	log.Debug().Msgf("args %v", args)
	log.Debug().Msgf("env %v", env)

	cmd := exec.Command(RESTIC_PATH, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)
	cmdOutput := &bytes.Buffer{}
	cmdStderr := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdStderr
	err := cmd.Run()

	if err != nil {
		log.Error().Err(err)
	}
	log.Info().Msgf("output: %v", cmdOutput.String())
	log.Info().Msgf("error: %v", cmdStderr.String())
	return cmdOutput.String()
}
