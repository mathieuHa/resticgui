package commands

import (
	"bytes"
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

func runRestic(args string) string {
	cmd := exec.Command(RESTIC_PATH, args)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err := cmd.Run()

	if err != nil {
		log.Error().Err(err)
	}
	log.Info().Msgf("output %v", cmdOutput.String())
	return cmdOutput.String()
}
