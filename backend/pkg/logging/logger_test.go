package logging

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"verbose", slog.LevelInfo},
		{"  warn ", slog.LevelWarn},
	}

	for _, tc := range cases {
		if got := ParseLevel(tc.in); got != tc.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestSetupThenLogger(t *testing.T) {
	l := Setup("debug", "")
	if l == nil {
		t.Fatal("Setup returned nil logger")
	}

	if Logger() != l {
		t.Error("Logger() should return the logger installed by Setup")
	}
}

func TestSetupWritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logs", "app.log")

	l := Setup("info", path)
	if l == nil {
		t.Fatal("Setup returned nil logger")
	}
	defer Destroy()

	Info("hello from test", slog.String("key", "value"))

	// Wait a moment for the async-ish flush, then read the file back.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(path)
		if err == nil && len(b) > 0 {
			s := string(b)
			if len(s) < 5 {
				continue
			}
			// JSON handler writes a JSON object per line.
			if s[0] != '{' {
				t.Fatalf("expected JSON log line, got: %s", s)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("log file was not written")
}

func TestLoggerDefaults(_ *testing.T) {
	// Fresh default should not crash when used.
	Debug("debug test")
	Info("info test")
	Warn("warn test")
	Error("error test")
}

func TestConsoleOutputIsAlwaysJSON(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(newHandler(&buf, slog.LevelInfo))

	l.Info("hello", slog.String("key", "value"))

	line := buf.String()
	if !strings.HasPrefix(line, "{") {
		t.Fatalf("expected JSON log record, got: %q", line)
	}
	// Sanity check the JSON contains the message and payload.
	if !strings.Contains(line, `"msg":"hello"`) || !strings.Contains(line, `"key":"value"`) {
		t.Fatalf("unexpected JSON record: %q", line)
	}
}
