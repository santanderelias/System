package privilege

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"store/internal/pkgmgr"
)

// ProgressFunc is an alias for pkgmgr.ProgressFunc.
type ProgressFunc = pkgmgr.ProgressFunc

// IsRoot reports whether the current process is running with root privileges (UID 0).
func IsRoot() bool {
	return os.Geteuid() == 0
}

// Execute runs a command. If the process is already root, it runs directly.
// If running as a normal user, it runs via sudo -S using the provided password over stdin.
func Execute(ctx context.Context, password string, progress ProgressFunc, name string, args ...string) error {
	bin, err := exec.LookPath(name)
	if err != nil {
		bin = name
	}

	if IsRoot() {
		// Running directly as root
		if progress != nil {
			progress("$ " + filepath.Base(bin) + " " + strings.Join(args, " "))
		}
		cmd := exec.CommandContext(ctx, bin, args...)
		cmd.Env = sanitizedEnv()

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start %s: %w", filepath.Base(bin), err)
		}

		errCh := make(chan error, 2)
		go func() { errCh <- streamLines(stdout, progress, nil) }()
		go func() { errCh <- streamLines(stderr, progress, nil) }()
		for i := 0; i < 2; i++ {
			if e := <-errCh; e != nil && err == nil {
				err = e
			}
		}
		if waitErr := cmd.Wait(); waitErr != nil {
			return fmt.Errorf("%s: %w", filepath.Base(bin), waitErr)
		}
		return err
	}

	// Running as normal user via sudo -S
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("sudo is required for elevated operations but was not found on PATH")
	}

	sudoArgs := append([]string{"-S", "-k", "-p", "", "--", bin}, args...)

	if progress != nil {
		progress("$ sudo " + filepath.Base(bin) + " " + strings.Join(args, " "))
	}

	cmd := exec.CommandContext(ctx, sudoPath, sudoArgs...)
	cmd.Env = sanitizedEnv()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start sudo: %w", err)
	}

	// Write password to stdin and close pipe immediately
	go func() {
		defer stdin.Close()
		if password != "" {
			_, _ = io.WriteString(stdin, password+"\n")
		}
	}()

	var errBuf bytes.Buffer
	errCh := make(chan error, 2)

	go func() { errCh <- streamLines(stdout, progress, nil) }()
	go func() { errCh <- streamLines(stderr, progress, &errBuf) }()

	for i := 0; i < 2; i++ {
		if e := <-errCh; e != nil && err == nil {
			err = e
		}
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		lower := strings.ToLower(errMsg)
		if strings.Contains(lower, "incorrect password") ||
			strings.Contains(lower, "try again") ||
			strings.Contains(lower, "authentication failure") ||
			strings.Contains(lower, "1 incorrect password attempt") {
			return fmt.Errorf("authentication failed: incorrect password")
		}
		if errMsg != "" {
			return fmt.Errorf("%s: %w (%s)", filepath.Base(bin), waitErr, errMsg)
		}
		return fmt.Errorf("%s: %w", filepath.Base(bin), waitErr)
	}

	return err
}

func streamLines(r io.Reader, progress ProgressFunc, capture *bytes.Buffer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if progress != nil {
			progress(line)
		}
		if capture != nil {
			capture.WriteString(line + "\n")
		}
	}
	return sc.Err()
}

func sanitizedEnv() []string {
	env := os.Environ()
	out := make([]string, 0, len(env)+2)
	for _, e := range env {
		if strings.HasPrefix(e, "LANG=") || strings.HasPrefix(e, "LC_ALL=") || strings.HasPrefix(e, "LC_MESSAGES=") {
			continue
		}
		out = append(out, e)
	}
	out = append(out, "LANG=C", "LC_ALL=C")
	return out
}
