package main

import (
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
