// Package logger manages file-based logging for Ralphex runs.
package logger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/iyaki/ralphex/internal/config"
)

const (
	logDirPerm  = 0o750
	logFilePerm = 0o600
)

// Logger handles logging to file.
type Logger struct {
	file    *os.File
	enabled bool
}

// NewLogger creates a new logger based on configuration.
func NewLogger(cfg *config.Config) (*Logger, error) {
	enabled := cfg.LogFile != ""

	logger := &Logger{
		enabled: enabled,
	}

	if !enabled {
		return logger, nil
	}

	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, logDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := openLogFile(cfg.LogFile, cfg.LogTruncate)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", cfg.LogFile, err)
	}

	logger.file = file

	_, _ = fmt.Fprintf(file, "===== Ralphex run started at %s =====\n", time.Now().Format("2006-01-02 15:04:05 -0700"))

	if gitBranch := getGitBranch(); gitBranch != "" {
		_, _ = fmt.Fprintf(file, "Git branch: %s\n", gitBranch)
	}

	if gitCommit := getGitCommit(); gitCommit != "" {
		_, _ = fmt.Fprintf(file, "Git commit: %s\n", gitCommit)
	}

	return logger, nil
}

func openLogFile(logFile string, truncate bool) (*os.File, error) {
	logAppend := !truncate

	if logAppend {
		// #nosec G304 -- log path is trusted configuration input
		return os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePerm)
	}

	// #nosec G304 -- log path is trusted configuration input
	return os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, logFilePerm)
}

// Close closes the logger.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}

	return nil
}

// Enabled returns whether logging is enabled.
func (l *Logger) Enabled() bool {
	return l.enabled
}

// File returns the log file.
func (l *Logger) File() *os.File {
	return l.file
}

// getGitBranch returns the current git branch name.
func getGitBranch() string {
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}

	branch := strings.TrimSpace(string(output))
	branch = strings.TrimPrefix(branch, "refs/heads/")

	return branch
}

// getGitCommit returns the current git commit hash.
func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}

	commit := strings.TrimSpace(string(output))

	return commit
}
