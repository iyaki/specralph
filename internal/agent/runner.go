package agent

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type synchronizedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.w.Write(p)
}

// BuildEffectiveEnv returns a deterministic child-process environment.
func BuildEffectiveEnv(overrides map[string]string) ([]string, error) {
	effective := mapFromEnvironment(os.Environ())

	for key, value := range overrides {
		if !envKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("invalid environment key %q", key)
		}

		effective[key] = value
	}

	return mapToEnvironment(effective), nil
}

func executeAgentCommand(
	command string,
	args []string,
	env []string,
	output io.Writer,
	errPrefix string,
) (string, error) {
	cmd := exec.Command(command, args...) // #nosec G204 G702 -- internally controlled by agent integrations
	if env != nil {
		cmd.Env = cloneStringSlice(env)
	}

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	if output == nil {
		output = io.Discard
	}
	streamOutput := &synchronizedWriter{w: output}

	cmd.Stdout = io.MultiWriter(&outBuf, streamOutput)
	cmd.Stderr = io.MultiWriter(&errBuf, streamOutput)

	err := cmd.Run()
	result := outBuf.String() + errBuf.String()

	if err != nil {
		return result, fmt.Errorf("%s execution failed: %w", errPrefix, err)
	}

	return result, nil
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}

	cloned := make([]string, len(values))
	copy(cloned, values)

	return cloned
}

func mapFromEnvironment(entries []string) map[string]string {
	envMap := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}

		envMap[key] = value
	}

	return envMap
}

func mapToEnvironment(entries map[string]string) []string {
	if len(entries) == 0 {
		return nil
	}

	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, fmt.Sprintf("%s=%s", key, entries[key]))
	}

	return result
}

func isAgentAvailable(command string) bool {
	_, err := exec.LookPath(command)

	return err == nil
}
