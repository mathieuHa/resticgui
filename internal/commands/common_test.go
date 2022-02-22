package commands

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGetRestoreCommonArgs(t *testing.T) {
	t.Run("test latest restore to /", func(t *testing.T) {
		path := "/"
		version := "latest"
		got := getRestoreCommonArgs(path, version)
		want := "restore latest --target /"
		assert.Equal(t, got, want)
	})
}
