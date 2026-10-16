package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"gotut/curriculum"
)

const (
	maxOutputBytes = 48 * 1024
	// runTimeout bounds "Run": a learner's main() that loops forever should be
	// killed quickly. testTimeout is more generous because the first "Check" for
	// an exercise with third-party dependencies has to download and compile them.
	runTimeout  = 20 * time.Second
	testTimeout = 75 * time.Second
	maxRunTime  = testTimeout
)

// TestResult is the outcome of a single go test case.
type TestResult struct {
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Output string `json:"output"`
}

// RunResult is what the browser receives after Run / Check.
type RunResult struct {
	OK         bool         `json:"ok"`
	Stdout     string       `json:"stdout"`
	Stderr     string       `json:"stderr"`
	ExitCode   int          `json:"exitCode"`
	DurationMs int64        `json:"durationMs"`
	TimedOut   bool         `json:"timedOut"`
	Tests      []TestResult `json:"tests"`
	Note       string       `json:"note"`
}

// limitBuffer collects output but stops growing past a fixed size.
type limitBuffer struct {
	mu    sync.Mutex
	buf   bytes.Buffer
	limit int
	trunc bool
}

func (l *limitBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	remaining := l.limit - l.buf.Len()
	if remaining <= 0 {
		l.trunc = true
		return len(p), nil
	}
	if len(p) > remaining {
		l.buf.Write(p[:remaining])
		l.trunc = true
		return len(p), nil
	}
	l.buf.Write(p)
	return len(p), nil
}

func (l *limitBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	s := l.buf.String()
	if l.trunc {
		s += "\n… output truncated …"
	}
	return s
}

// executeExercise writes the learner's files (plus the hidden tests, when mode
// is "test") into a throwaway module and runs them.
func executeExercise(ex *curriculum.Exercise, files []curriculum.File, mode, stdin string, timeout time.Duration) RunResult {
	start := time.Now()
	res := RunResult{ExitCode: -1}

	dir, err := os.MkdirTemp("", "gotut-ex-*")
	if err != nil {
		res.Stderr = "could not create sandbox: " + err.Error()
		return res
	}
	defer os.RemoveAll(dir)

	if err := writeModule(dir, ex, files, mode); err != nil {
		res.Stderr = "could not prepare sandbox: " + err.Error()
		return res
	}

	if timeout <= 0 {
		if mode == "test" || mode == "check" {
			timeout = testTimeout
		} else {
			timeout = runTimeout
		}
	}
	if timeout > maxRunTime {
		timeout = maxRunTime
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var args []string
	if mode == "test" || mode == "check" {
		args = []string{"test", "-count=1", "-v", "-timeout", "20s", "."}
	} else {
		args = []string{"run", "."}
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = sandboxEnv()
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	// `go run` and `go test` both spawn a child process. Killing only the go
	// tool leaves that child running and holding our stdout/stderr pipes open,
	// so Run would block forever. Put the whole thing in its own process group
	// and kill the group; WaitDelay is the belt-and-braces guarantee that Run
	// returns even if something still holds a pipe.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	cmd.WaitDelay = 3 * time.Second

	var stdout, stderr limitBuffer
	stdout.limit, stderr.limit = maxOutputBytes, maxOutputBytes
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	res.DurationMs = time.Since(start).Milliseconds()
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()

	if ctx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		res.Note = "Execution stopped after " + timeout.Round(time.Second).String() + " (possible infinite loop?)."
		res.Stderr = strings.TrimSpace(res.Stderr + "\n" + res.Note)
		return res
	}

	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
		res.OK = true
		res.ExitCode = 0
	case errors.As(runErr, &exitErr):
		res.ExitCode = exitErr.ExitCode()
		res.OK = false
	default:
		res.ExitCode = -1
		res.Stderr += "\nfailed to start go toolchain: " + runErr.Error()
	}

	if mode == "test" || mode == "check" {
		res.Tests = parseTests(res.Stdout)
		if len(res.Tests) == 0 && res.ExitCode != 0 {
			// Likely a compile error: keep OK=false so the UI shows the build log.
			res.OK = false
		}
		if len(res.Tests) > 0 {
			allPass := true
			for _, t := range res.Tests {
				if !t.Pass {
					allPass = false
				}
			}
			res.OK = allPass
		}
	}
	return res
}

func writeModule(dir string, ex *curriculum.Exercise, files []curriculum.File, mode string) error {
	gomod := "module exercise\n\ngo 1.22\n"
	// Only pinned dependencies are written explicitly; anything without a
	// version is left out so that GOFLAGS=-mod=mod can resolve it.
	var requires []string
	for _, d := range ex.Deps {
		if strings.Contains(strings.TrimSpace(d), " ") {
			requires = append(requires, strings.TrimSpace(d))
		}
	}
	if len(requires) > 0 {
		gomod += "\nrequire (\n"
		for _, d := range requires {
			gomod += "\t" + d + "\n"
		}
		gomod += ")\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		return err
	}

	all := append([]curriculum.File{}, files...)
	if mode == "test" || mode == "check" {
		all = append(all, ex.Tests...)
	}
	for _, f := range all {
		name := filepath.Base(curriculum.Render(f.Name))
		if name == "" || name == "." {
			name = "main.go"
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(curriculum.Render(f.Content)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func sandboxEnv() []string {
	env := os.Environ()
	env = append(env,
		"GOFLAGS=-mod=mod",
		"GOWORK=off",
		"GOTOOLCHAIN=local",
		"GONOSUMDB=*",
		"GO111MODULE=on",
	)
	return env
}

var (
	reTestRun  = regexp.MustCompile(`(?m)^\s*===\s+RUN\s+(\S+)`)
	reTestPass = regexp.MustCompile(`(?m)^\s*---\s+PASS:\s+(\S+)`)
	reTestFail = regexp.MustCompile(`(?m)^\s*---\s+FAIL:\s+(\S+)`)
	reTestSkip = regexp.MustCompile(`(?m)^\s*---\s+SKIP:\s+(\S+)`)
)

// parseTests turns `go test -v` output into a list of cases in execution order.
func parseTests(out string) []TestResult {
	order := []string{}
	pass := map[string]bool{}
	seen := map[string]bool{}
	for _, m := range reTestRun.FindAllStringSubmatch(out, -1) {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			order = append(order, name)
		}
	}
	for _, m := range reTestPass.FindAllStringSubmatch(out, -1) {
		pass[m[1]] = true
	}
	fail := map[string]bool{}
	for _, m := range reTestFail.FindAllStringSubmatch(out, -1) {
		fail[m[1]] = true
	}
	for _, m := range reTestSkip.FindAllStringSubmatch(out, -1) {
		fail[m[1]] = true
	}

	results := make([]TestResult, 0, len(order))
	for _, name := range order {
		results = append(results, TestResult{
			Name: name,
			Pass: pass[name] && !fail[name],
		})
	}
	return results
}
