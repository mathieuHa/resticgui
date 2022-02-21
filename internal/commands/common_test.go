package commands

import (
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

func TestGetResticVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResticVersion(); got != tt.want {
				t.Errorf("GetResticVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResticHelp(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResticHelp(); got != tt.want {
				t.Errorf("GetResticHelp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSnapshots(t *testing.T) {
	type args struct {
		repository Models.Repository
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSnapshots(tt.args.repository); got != tt.want {
				t.Errorf("GetLocalSnapshots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitRepository(t *testing.T) {
	type args struct {
		repository Models.Repository
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitRepository(tt.args.repository); got != tt.want {
				t.Errorf("InitRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestoreRepository(t *testing.T) {
	type args struct {
		repository Models.Repository
		path       string
		version    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RestoreRepository(tt.args.repository, tt.args.path, tt.args.version); got != tt.want {
				t.Errorf("RestoreRepository() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRestoreCommonArgs(tt.args.path, tt.args.version); got != tt.want {
				t.Errorf("getRestoreCommonArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestForgetSnaphots(t *testing.T) {
	type args struct {
		repository Models.Repository
		version    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ForgetSnaphots(tt.args.repository, tt.args.version); got != tt.want {
				t.Errorf("ForgetSnaphots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPruneSnaphots(t *testing.T) {
	type args struct {
		repository Models.Repository
		version    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PruneSnaphots(tt.args.repository, tt.args.version); got != tt.want {
				t.Errorf("PruneSnaphots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getProviderFormat(t *testing.T) {
	type args struct {
		repository Models.Repository
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProviderFormat(tt.args.repository); got != tt.want {
				t.Errorf("getProviderFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackupRepository(t *testing.T) {
	type args struct {
		repository Models.Repository
		dryRun     bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BackupRepository(tt.args.repository, tt.args.dryRun); got != tt.want {
				t.Errorf("BackupRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatsRepository(t *testing.T) {
	type args struct {
		repository Models.Repository
		dryRun     bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatsRepository(tt.args.repository, tt.args.dryRun); got != tt.want {
				t.Errorf("StatsRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getBackupCommonArgs(t *testing.T) {
	type args struct {
		repository Models.Repository
		dryRun     bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBackupCommonArgs(tt.args.repository, tt.args.dryRun); got != tt.want {
				t.Errorf("getBackupCommonArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDryRunParameter(t *testing.T) {
	type args struct {
		dryRun bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDryRunParameter(tt.args.dryRun); got != tt.want {
				t.Errorf("getDryRunParameter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMaxSizeParameter(t *testing.T) {
	type args struct {
		repository Models.Repository
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMaxSizeParameter(tt.args.repository); got != tt.want {
				t.Errorf("getMaxSizeParameter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIncludedAndExludedFiles(t *testing.T) {
	type args struct {
		repository Models.Repository
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIncludedAndExludedFiles(tt.args.repository); got != tt.want {
				t.Errorf("getIncludedAndExludedFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLocalRepositoryFormat(t *testing.T) {
	type args struct {
		local Models.LocalProvider
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLocalRepositoryFormat(tt.args.local); got != tt.want {
				t.Errorf("getLocalRepositoryFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getB2RepositoryFormat(t *testing.T) {
	type args struct {
		b2 Models.B2Provider
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getB2RepositoryFormat(tt.args.b2); got != tt.want {
				t.Errorf("getB2RepositoryFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAmazonS3RepositoryFormat(t *testing.T) {
	type args struct {
		aws Models.AWSS3Provider
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAmazonS3RepositoryFormat(tt.args.aws); got != tt.want {
				t.Errorf("getAmazonS3RepositoryFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMinioS3RepositoryFormat(t *testing.T) {
	type args struct {
		minio Models.MinioS3Provider
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMinioS3RepositoryFormat(tt.args.minio); got != tt.want {
				t.Errorf("getMinioS3RepositoryFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckIntegrity(t *testing.T) {
	type args struct {
		repository  string
		full        bool
		pourcentage float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckIntegrity(tt.args.repository, tt.args.full, tt.args.pourcentage); got != tt.want {
				t.Errorf("CheckIntegrity() = %v, want %v", got, tt.want)
			}
		})
	}
}
