package commands

import (
	"reflect"
	"testing"

	"github.com/go-playground/assert/v2"
	"gitlab.com/resticgui/resticgui/internal/Models"
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

func Test_getProviderFormat(t *testing.T) {
	type args struct {
		repository Models.Repository
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProviderFormat(tt.args.repository); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProviderFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRestoreCommonArgs(t *testing.T) {
	type args struct {
		path    string
		version string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"restore common args", args{"/", "latest"}, "restore latest --target /"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRestoreCommonArgs(tt.args.path, tt.args.version); got != tt.want {
				t.Errorf("getRestoreCommonArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
