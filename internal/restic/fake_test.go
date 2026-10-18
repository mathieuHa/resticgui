package restic

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var fakeBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "fakerestic-")
	if err != nil {
		panic(err)
	}
	fakeBin = filepath.Join(dir, "restic")
	if out, err := exec.Command("go", "build", "-o", fakeBin, "./internal/fakerestic").CombinedOutput(); err != nil {
		panic(string(out))
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// scenario mirrors fakerestic's Scenario.
type scenario struct {
	Record       string   `json:"record"`
	Stdout       []string `json:"stdout,omitempty"`
	StdoutFile   string   `json:"stdout_file,omitempty"`
	Stderr       []string `json:"stderr,omitempty"`
	LineDelay    int      `json:"line_delay_ms,omitempty"`
	Exit         int      `json:"exit,omitempty"`
	Block        bool     `json:"block,omitempty"`
	IgnoreSigint bool     `json:"ignore_sigint,omitempty"`
	SigintExit   int      `json:"sigint_exit,omitempty"`
}

type invocation struct {
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	Files     map[string]string `json:"files"`
	GotSigint bool              `json:"got_sigint"`
}

type fake struct {
	t      *testing.T
	runner *Runner
	repo   *Repo
	record string
}

// newFake returns a runner using fakerestic with sc, and a repository whose
// environment points the fake at its scenario.
func newFake(t *testing.T, sc scenario) *fake {
	t.Helper()
	dir := t.TempDir()
	sc.Record = filepath.Join(dir, "record.json")
	if sc.StdoutFile != "" && !filepath.IsAbs(sc.StdoutFile) {
		abs, err := filepath.Abs(sc.StdoutFile)
		require.NoError(t, err)
		sc.StdoutFile = abs
	}
	raw, err := json.Marshal(sc)
	require.NoError(t, err)
	scenarioFile := filepath.Join(dir, "scenario.json")
	require.NoError(t, os.WriteFile(scenarioFile, raw, 0o600))

	return &fake{
		t:      t,
		runner: &Runner{Bin: fakeBin, Env: []string{"PATH=" + os.Getenv("PATH")}, WaitDelay: 2 * time.Second},
		repo: &Repo{
			Location: "/srv/restic/docs",
			Password: "repository-password",
			Env:      map[string]string{"FAKE_RESTIC_SCENARIO": scenarioFile},
		},
		record: sc.Record,
	}
}

func (f *fake) invocation() invocation {
	f.t.Helper()
	raw, err := os.ReadFile(f.record)
	require.NoError(f.t, err)
	var inv invocation
	require.NoError(f.t, json.Unmarshal(raw, &inv))
	return inv
}
