// Package publicthreads holds hand-maintained Public Thread API helpers that
// Lathe cannot currently express through generated specs: waiting on a thread
// run, printing a completed run's final output, and rendering a thread
// transcript.
//
// The client here is intentionally transport/client-level only — auth headers,
// base URL resolution, request execution, and JSON error decoding for the
// thread/run/event read surface. It deliberately does NOT implement file
// upload (upload-session, byte upload, complete, attach); that surface is
// owned separately by the public file upload helper.
package publicthreads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
	"github.com/spf13/cobra"
)

// transportFunc executes a single Public Thread API request. It is a thin seam
// over latheruntime.DoRawFull so tests can inject a fake without a live host.
type transportFunc func(ctx context.Context, method, path string, body any, headers map[string]string) (*latheruntime.RawResult, error)

// Client is a minimal Public Thread API client scoped to thread, run, and
// event reads plus thread creation. It owns no upload behavior.
type Client struct {
	hostname  string
	opts      latheruntime.ClientOptions
	transport transportFunc
}

// NewClient builds a Client from the resolved host options for cmd. The
// surrounding command tree (mounted under public-thread-api) resolves the
// correct base URL via the target pre-run, exactly like generated commands.
func NewClient(cmd *cobra.Command) (*Client, error) {
	hostname, opts, err := latheruntime.LoadHostOptions(cmd)
	if err != nil {
		return nil, err
	}
	opts.UserAgent = cmd.Root().Use
	opts.Accept = "application/json"
	if debug, derr := cmd.Root().PersistentFlags().GetBool("debug"); derr == nil && debug {
		opts.Debug = true
	}
	return &Client{hostname: hostname, opts: opts}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, headers map[string]string) (*latheruntime.RawResult, error) {
	if c.transport != nil {
		return c.transport(ctx, method, path, body, headers)
	}
	opts := c.opts
	if len(headers) > 0 {
		merged := make(map[string]string, len(c.opts.Headers)+len(headers))
		for k, v := range c.opts.Headers {
			merged[k] = v
		}
		for k, v := range headers {
			merged[k] = v
		}
		opts.Headers = merged
	}
	return latheruntime.DoRawFull(ctx, c.hostname, method, path, body, opts)
}

// APIError is a decoded Public Thread API error envelope. It carries the stable
// machine-readable code so callers and tests can branch on it.
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	switch {
	case e.Code != "" && e.Message != "":
		return fmt.Sprintf("public thread api error (%d %s): %s", e.Status, e.Code, e.Message)
	case e.Message != "":
		return fmt.Sprintf("public thread api error (%d): %s", e.Status, e.Message)
	default:
		return fmt.Sprintf("public thread api error: status %d", e.Status)
	}
}

// decodeError turns a transport error into an APIError when the body carries
// the standard error envelope, so failures surface a clean code/message instead
// of a raw HTTP dump.
func decodeError(err error) error {
	var he *latheruntime.HTTPError
	if !errors.As(err, &he) {
		return err
	}
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if jerr := json.Unmarshal(he.Body, &env); jerr == nil && (env.Error.Code != "" || env.Error.Message != "") {
		return &APIError{Status: he.Status, Code: env.Error.Code, Message: env.Error.Message}
	}
	return &APIError{Status: he.Status, Message: string(he.Body)}
}

func (c *Client) getJSON(ctx context.Context, method, path string, body any, headers map[string]string, out any) ([]byte, error) {
	res, err := c.do(ctx, method, path, body, headers)
	if err != nil {
		return nil, decodeError(err)
	}
	if out != nil {
		if err := json.Unmarshal(res.Body, out); err != nil {
			return res.Body, fmt.Errorf("decode %s %s response: %w", method, path, err)
		}
	}
	return res.Body, nil
}

// CreateThread creates a thread for an agent. body is the raw JSON request body
// (may be nil for an empty thread). idempotencyKey, when non-empty, is sent as
// the Idempotency-Key header for retry-safe creation.
func (c *Client) CreateThread(ctx context.Context, agentID string, body []byte, idempotencyKey string) (*ThreadState, error) {
	var headers map[string]string
	if idempotencyKey != "" {
		headers = map[string]string{"Idempotency-Key": idempotencyKey}
	}
	st := &ThreadState{}
	raw, err := c.getJSON(ctx, "POST", "/agents/"+pathEscape(agentID)+"/threads", bodyOrNil(body), headers, st)
	if err != nil {
		return nil, err
	}
	st.Raw = raw
	return st, nil
}

// RetrieveThread fetches the current thread state, including the most recent run
// with its status, error, and final output.
func (c *Client) RetrieveThread(ctx context.Context, threadID string) (*ThreadState, error) {
	st := &ThreadState{}
	raw, err := c.getJSON(ctx, "GET", "/threads/"+pathEscape(threadID), nil, nil, st)
	if err != nil {
		return nil, err
	}
	st.Raw = raw
	return st, nil
}

// ListEvents returns up to limit of the latest thread events, oldest first.
func (c *Client) ListEvents(ctx context.Context, threadID string, limit int) (*EventList, error) {
	path := "/threads/" + pathEscape(threadID) + "/events"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	list := &EventList{}
	if _, err := c.getJSON(ctx, "GET", path, nil, nil, list); err != nil {
		return nil, err
	}
	return list, nil
}

// bodyOrNil avoids sending an empty "" body, which DoRawFull would otherwise
// encode and send with a Content-Type header.
func bodyOrNil(body []byte) any {
	if len(body) == 0 {
		return nil
	}
	return body
}
