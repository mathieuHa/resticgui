package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

var (
	RESTIC_PATH string
)

func resticInit() {
	if runtime.GOOS == "windows" {
		RESTIC_PATH = "restic.exe"
	}

	if runtime.GOOS == "linux" {
		RESTIC_PATH = "restic"
	}
}

func runRestic() {
	cmd := exec.Command(RESTIC_PATH, "version")
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err := cmd.Run()

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(cmdOutput.String())
}
