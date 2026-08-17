package ollama

import (
	"context"
)

// Serve starts the Ollama server. It does not wait for readiness.
func (c *Client) Serve(ctx context.Context) error {
	return c.runStart(ctx, "serve")
}

// Show retrieves detailed information about a model.
func (c *Client) Show(ctx context.Context, model string) ([]byte, error) {
	if model == "" {
		return nil, errEmptyModel
	}
	return c.run(ctx, "show", model)
}

// Stop stops a running model. If model is empty, stops all running models.
func (c *Client) Stop(ctx context.Context, model string) ([]byte, error) {
	if model == "" {
		return c.run(ctx, "stop")
	}
	return c.run(ctx, "stop", model)
}

// Pull downloads a model from the registry.
func (c *Client) Pull(ctx context.Context, model string) ([]byte, error) {
	if model == "" {
		return nil, errEmptyModel
	}
	return c.run(ctx, "pull", model)
}

// List returns all locally available models.
func (c *Client) List(ctx context.Context) ([]byte, error) {
	return c.run(ctx, "list")
}

// Ps returns currently running models.
func (c *Client) Ps(ctx context.Context) ([]byte, error) {
	return c.run(ctx, "ps")
}

// Rm removes a model.
func (c *Client) Rm(ctx context.Context, model string) ([]byte, error) {
	if model == "" {
		return nil, errEmptyModel
	}
	return c.run(ctx, "rm", model)
}