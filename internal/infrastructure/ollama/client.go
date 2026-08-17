// Package ollama provides a thin wrapper around the Ollama CLI.
// It is designed for local machine usage but can be extended to remote calls later.
package ollama

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Client wraps the ollama executable and provides methods for common commands.
type Client struct {
	binPath string
	debug   bool
	stdout  io.Writer
	stderr  io.Writer
	timeout time.Duration // default timeout for commands (0 means no timeout)
}

// Option configures Client at creation.
type Option func(*Client)

// WithBinPath sets a custom path to the ollama binary.
// If not set, NewClient will try to locate it in PATH.
func WithBinPath(path string) Option {
	return func(c *Client) { c.binPath = path }
}

// WithDebug enables logging of executed commands to stdout/stderr.
func WithDebug() Option {
	return func(c *Client) { c.debug = true }
}

// WithOutput sets the writer to which command output will be streamed.
// The output is also captured and returned by methods that return []byte.
func WithOutput(w io.Writer) Option {
	return func(c *Client) {
		c.stdout = w
		c.stderr = w
	}
}

// WithTimeout sets a default timeout for commands. It is applied only when
// the passed context does not already have a deadline.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// NewClient creates a new Ollama client.
// It does not check for the binary; use EnsureInstalled for that.
func NewClient(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	if c.binPath == "" {
		c.binPath = defaultBinPath()
	}
	return c
}

// EnsureInstalled returns an error if the ollama executable cannot be found.
func (c *Client) EnsureInstalled() error {
	if _, err := exec.LookPath(c.binPath); err != nil {
		return fmt.Errorf("ollama binary not found: %w", err)
	}
	return nil
}

// run executes the command and returns combined stdout+stderr.
// Output is simultaneously written to client's writers (if set) via MultiWriter.
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	if c.debug {
		fmt.Fprintf(c.stderr, "[ollama] running: %s %s\n", c.binPath, strings.Join(args, " "))
	}

	// Apply default timeout if ctx has no deadline.
	if c.timeout > 0 {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.timeout)
			defer cancel()
		}
	}

	cmd := exec.CommandContext(ctx, c.binPath, args...)
	var buf bytes.Buffer

	if c.stdout != nil {
		cmd.Stdout = io.MultiWriter(&buf, c.stdout)
	} else {
		cmd.Stdout = &buf
	}

	if c.stderr != nil {
		cmd.Stderr = io.MultiWriter(&buf, c.stderr)
	} else {
		cmd.Stderr = &buf
	}

	err := cmd.Run()
	return buf.Bytes(), err
}

// runStart runs a long-lived command and does not wait for it to finish.
// It returns after the process has been started successfully.
func (c *Client) runStart(ctx context.Context, args ...string) error {
	if c.debug {
		fmt.Fprintf(c.stderr, "[ollama] starting: %s %s\n", c.binPath, strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, c.binPath, args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	// Release resources (zombie prevention) — command will run in background.
	go cmd.Wait()
	return nil
}

// defaultBinPath returns "ollama" or "ollama.exe" depending on OS.
func defaultBinPath() string {
	if runtime.GOOS == "windows" {
		return "ollama.exe"
	}
	return "ollama"
}