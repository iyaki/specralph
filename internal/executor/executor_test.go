package executor_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/iyaki/specralph/internal/executor"
)

type synchronizedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *synchronizedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *synchronizedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func TestExecuteCommandSuccess(t *testing.T) {
	var out bytes.Buffer
	result, err := executor.ExecuteCommand("sh", []string{"-c", "echo out && echo err 1>&2"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "out") || !strings.Contains(result, "err") {
		t.Fatalf("unexpected result: %q", result)
	}
	if !strings.Contains(out.String(), "out") || !strings.Contains(out.String(), "err") {
		t.Fatalf("expected writer output to contain stdout and stderr; got %q", out.String())
	}
}

func TestExecuteCommandFailureReturnsOutputAndError(t *testing.T) {
	result, err := executor.ExecuteCommand("sh", []string{"-c", "echo partial && exit 2"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "command execution failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "partial") {
		t.Fatalf("expected partial output to be returned, got %q", result)
	}
}

func TestExecuteCommandStreamsOutputInRealTime(t *testing.T) {
	out := &synchronizedBuffer{}

	done := make(chan struct{})
	var (
		result string
		err    error
	)

	go func() {
		result, err = executor.ExecuteCommand(
			"sh",
			[]string{"-c", "printf 'first\n'; sleep 0.3; printf 'second\n'"},
			out,
		)
		close(done)
	}()

	deadline := time.Now().Add(200 * time.Millisecond)
	sawFirstLine := false
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), "first") {
			sawFirstLine = true

			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !sawFirstLine {
		t.Fatalf("expected first line to be streamed before command completion, got %q", out.String())
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for command completion")
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "first") || !strings.Contains(result, "second") {
		t.Fatalf("expected complete output in result, got %q", result)
	}
}
