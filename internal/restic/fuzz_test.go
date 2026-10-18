package restic

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// restic's output describes repository content, which a repository shared
// with others can make anything: none of it may crash ResticGUI.

func FuzzLineWriter(f *testing.F) {
	f.Add([]byte("{\"message_type\":\"status\"}\n{\"a\":1}\r\n\npartial"), 3)
	f.Add([]byte("\n\n\r\n"), 1)
	f.Fuzz(func(t *testing.T, data []byte, cut int) {
		collect := func(chunks ...[]byte) []string {
			var lines []string
			w := &lineWriter{onLine: func(l []byte) { lines = append(lines, string(l)) }}
			for _, c := range chunks {
				if n, err := w.Write(c); n != len(c) || err != nil {
					t.Fatalf("Write returned %d, %v", n, err)
				}
			}
			w.flush()
			return lines
		}
		if cut < 0 {
			cut = -cut
		}
		cut %= len(data) + 1
		whole, split := collect(data), collect(data[:cut], data[cut:])
		if strings.Join(whole, "\x00") != strings.Join(split, "\x00") {
			t.Fatalf("splitting the stream changed the lines: %q vs %q", whole, split)
		}
		for i, l := range whole[:max(0, len(whole)-1)] {
			if strings.Contains(l, "\n") || l == "" {
				t.Fatalf("line %d is %q", i, l)
			}
		}
	})
}

func FuzzStderrTail(f *testing.F) {
	f.Add([]byte("\x1b[2Kscanning...\r{\"message_type\":\"exit_error\",\"code\":1,\"message\":\"Fatal: wrong password\"}\n"))
	f.Add(bytes.Repeat([]byte("x"), 100))
	f.Fuzz(func(t *testing.T, data []byte) {
		s := &stderrTail{max: 5}
		_, _ = s.Write(data)
		lines := s.lines()
		if len(lines) > 5 {
			t.Fatalf("%d lines kept", len(lines))
		}
		for _, l := range lines {
			if len(l) > maxStderrLine || strings.ContainsAny(l, "\r\n") {
				t.Fatalf("line %q", l)
			}
		}
		_ = s.exitMessage()
	})
}

func FuzzDecode(f *testing.F) {
	for _, seed := range []string{
		`{"message_type":"summary","files_new":1,"snapshot_id":"abc"}`,
		`{"name":"a\u0000b\nc","type":"file","path":"/x/\u001b[31m","size":18446744073709551615}`,
		`[{"id":"1","time":"2026-01-01T00:00:00Z","paths":["/x"],"tags":null}]`,
		`{"message_type":"change","path":"/x","modifier":"+"}`,
		"\xff\xfe",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = messageType(data)
		var (
			snaps  []Snapshot
			node   Node
			change DiffChange
			groups []ForgetGroup
			status BackupStatus
			sum    BackupSummary
			stats  Stats
		)
		for _, v := range []any{&snaps, &node, &change, &groups, &status, &sum, &stats} {
			if json.Unmarshal(data, v) == nil {
				if _, err := json.Marshal(v); err != nil {
					t.Fatalf("decoded but cannot encode again for the API: %v", err)
				}
			}
		}
		if node.Name != "" && !utf8.ValidString(node.Name) {
			// encoding/json replaces invalid UTF-8; the API never sends it raw
			t.Fatalf("invalid UTF-8 kept in %q", node.Name)
		}
	})
}
