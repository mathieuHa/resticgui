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

func runRestic(args []string, env []string, path string) (string, string, error) {
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
		log.Error().Err(err).Msg("Error in cmd run")
	}
	log.Debug().Msgf("error: %v", cmdStderr.String())
	if path != "" {
		f, err := os.Create(path)
		if err != nil {
			log.Err(err).Msg("Could not open file dest")
			return "", cmdStderr.String(), err
		}
		defer f.Close()
		f.Write(cmdOutput.Bytes())
		return "File written with success at " + path, cmdStderr.String(), err
	}
	log.Debug().Msgf("output: %v", cmdOutput.String())

	return cmdOutput.String(), cmdStderr.String(), err
}
