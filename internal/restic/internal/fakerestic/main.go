// Command fakerestic stands in for restic in the internal/restic tests. It
// reads a scenario from the JSON file named by $FAKE_RESTIC_SCENARIO, records
// how it was called, prints canned output and exits with the scenario's code.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Scenario is shared with the tests through JSON.
type Scenario struct {
	// Record is where the invocation is written as JSON.
	Record string `json:"record"`
	// Stdout is printed line by line, LineDelay apart.
	Stdout     []string `json:"stdout"`
	StdoutFile string   `json:"stdout_file"`
	Stderr     []string `json:"stderr"`
	LineDelay  int      `json:"line_delay_ms"`
	Exit       int      `json:"exit"`
	// Block waits for SIGINT after printing; restic then exits with SigintExit.
	Block        bool `json:"block"`
	IgnoreSigint bool `json:"ignore_sigint"`
	SigintExit   int  `json:"sigint_exit"`
}

// Invocation is what the fake records.
type Invocation struct {
	Args []string          `json:"args"`
	Env  map[string]string `json:"env"`
	// Files holds the content of files passed with --files-from-verbatim and --exclude-file.
	Files     map[string]string `json:"files"`
	GotSigint bool              `json:"got_sigint"`
}

func main() {
	var sc Scenario
	if raw, err := os.ReadFile(os.Getenv("FAKE_RESTIC_SCENARIO")); err == nil {
		if err := json.Unmarshal(raw, &sc); err != nil {
			fmt.Fprintln(os.Stderr, "fakerestic: bad scenario:", err)
			os.Exit(99)
		}
	}

	inv := Invocation{Args: os.Args[1:], Env: map[string]string{}, Files: map[string]string{}}
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		inv.Env[k] = v
	}
	for i, a := range inv.Args {
		if (a == "--files-from-verbatim" || a == "--exclude-file") && i+1 < len(inv.Args) {
			content, _ := os.ReadFile(inv.Args[i+1])
			inv.Files[a] = string(content)
		}
	}
	record := func() {
		if sc.Record != "" {
			raw, _ := json.Marshal(inv)
			_ = os.WriteFile(sc.Record, raw, 0o600)
		}
	}
	record()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)

	lines := sc.Stdout
	if sc.StdoutFile != "" {
		raw, err := os.ReadFile(sc.StdoutFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "fakerestic:", err)
			os.Exit(99)
		}
		lines = append(lines, strings.Split(strings.TrimRight(string(raw), "\n"), "\n")...)
	}
	for _, l := range lines {
		fmt.Println(l)
		if sc.LineDelay > 0 {
			time.Sleep(time.Duration(sc.LineDelay) * time.Millisecond)
		}
	}
	for _, l := range sc.Stderr {
		fmt.Fprintln(os.Stderr, l)
	}

	if sc.Block {
		for {
			<-sigint
			inv.GotSigint = true
			record()
			if !sc.IgnoreSigint {
				fmt.Fprintln(os.Stderr, "\x1b[2Ksignal interrupt received, cleaning up")
				os.Exit(sc.SigintExit)
			}
		}
	}
	os.Exit(sc.Exit)
}
