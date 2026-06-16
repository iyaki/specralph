// Package main implements a test-only agent for e2e testing.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	delayDuration     = 100 * time.Millisecond
	exitCodeSuccess   = 0
	exitCodeError     = 1
	exitCodeUnknown   = 2
	modeCompleteOnce  = "complete_once"
	modeNeverComplete = "never_complete"
	modeReturnError   = "return_error"
	modeSlowComplete  = "slow_complete"
)

// coverage:ignore
func main() {
	os.Exit(run(os.Args, os.Getenv, os.Stdout, os.Stderr))
}

func run(args []string, getEnv func(string) string, stdout, stderr io.Writer) int {
	mode := getEnv("RALPH_TEST_AGENT_MODE")
	if mode == "" {
		mode = modeCompleteOnce
	}

	// Basic logging to stderr for debugging
	_, _ = fmt.Fprintf(stderr, "[ralph-test-agent] Starting in mode: %s\n", mode)
	_, _ = fmt.Fprintf(stderr, "[ralph-test-agent] Args: %v\n", args)
	emitRequestedEnv(getEnv, stderr)

	switch mode {
	case modeCompleteOnce:
		_, _ = fmt.Fprintln(stdout, "Processing request...")
		_, _ = fmt.Fprintln(stdout, "<promise>COMPLETE</promise>")

		return exitCodeSuccess
	case modeNeverComplete:
		_, _ = fmt.Fprintln(stdout, "Processing request forever...")
		// Just exit without the complete token

		return exitCodeSuccess
	case modeReturnError:
		_, _ = fmt.Fprintln(stderr, "Simulated agent failure")

		return exitCodeError
	case modeSlowComplete:
		_, _ = fmt.Fprintln(stdout, "Thinking...")
		time.Sleep(delayDuration)
		_, _ = fmt.Fprintln(stdout, "<promise>COMPLETE</promise>")

		return exitCodeSuccess
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown mode: %s\n", mode)

		return exitCodeUnknown
	}
}

func emitRequestedEnv(getEnv func(string) string, stderr io.Writer) {
	rawKeys := strings.TrimSpace(getEnv("RALPH_TEST_AGENT_ECHO_ENV_KEYS"))
	if rawKeys == "" {
		return
	}

	for _, rawKey := range strings.Split(rawKeys, ",") {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			continue
		}

		_, _ = fmt.Fprintf(stderr, "[ralph-test-agent] Env %s=%s\n", key, getEnv(key))
	}
}
