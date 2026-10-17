package legacy

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mathieuHa/resticgui/internal/secrets"
	"github.com/mathieuHa/resticgui/internal/store"
)

// Main runs "resticgui migrate-legacy" and returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("migrate-legacy", flag.ContinueOnError)
	flags.SetOutput(stderr)
	from := flags.String("from", "", "path of the v0.1 database (read-only), e.g. resticgui.db")
	to := flags.String("to", "", "path of the new database to create")
	keyFile := flags.String("key-file", "", "secret key file, created if missing (default: resticgui.key next to -to); $"+secrets.KeyEnv+" takes precedence")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: resticgui migrate-legacy -from OLD.db -to NEW.db [-key-file FILE]")
		fmt.Fprintln(stderr, "\nCopies repositories, paths, schedules and retention policies from a ResticGUI v0.1 database.")
		fmt.Fprintln(stderr, "Passwords and credentials are encrypted in the new database. Command logs are not copied.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *from == "" || *to == "" {
		flags.Usage()
		return 2
	}
	if *keyFile == "" {
		*keyFile = filepath.Join(filepath.Dir(*to), "resticgui.key")
	}

	if err := run(*from, *to, *keyFile, stdout); err != nil {
		fmt.Fprintln(stderr, "migrate-legacy:", err)
		return 1
	}
	return 0
}

func run(from, to, keyFile string, out io.Writer) error {
	if _, err := os.Stat(from); err != nil {
		return err
	}
	if _, err := os.Stat(to); err == nil {
		return fmt.Errorf("%s already exists, refusing to import into an existing database", to)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	key, err := secrets.LoadKey(keyFile)
	if err != nil {
		return err
	}
	box, err := secrets.New(key)
	if err != nil {
		return err
	}

	ctx := context.Background()
	dst, err := store.Open(ctx, to)
	if err != nil {
		return err
	}
	report, err := Import(ctx, from, dst, box)
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(to)
		os.Remove(to + "-wal")
		os.Remove(to + "-shm")
		return err
	}

	fmt.Fprintf(out, "Imported %d repositories into %s\n", len(report.Imported), to)
	for _, name := range report.Imported {
		fmt.Fprintln(out, "  +", name)
	}
	for _, s := range report.Skipped {
		fmt.Fprintln(out, "  skipped", s)
	}
	for _, w := range report.Warnings {
		fmt.Fprintln(out, "  warning:", w)
	}
	if report.LogsIgnored > 0 {
		fmt.Fprintf(out, "%d command log entries were not copied.\n", report.LogsIgnored)
	}
	fmt.Fprintf(out, "Keep %s safe (or set $%s): without it the stored passwords cannot be read.\n", keyFile, secrets.KeyEnv)
	return nil
}
