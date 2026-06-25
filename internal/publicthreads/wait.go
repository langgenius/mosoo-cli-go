package publicthreads

import (
	"context"
	"errors"
	"time"
)

// DefaultPollInterval is how often WaitForRun re-reads thread state.
const DefaultPollInterval = 2 * time.Second

// DefaultWaitTimeout bounds how long a wait blocks before giving up.
const DefaultWaitTimeout = 5 * time.Minute

// ErrWaitTimeout is returned (wrapped) when the wait deadline elapses before
// the watched run reaches a stop state. The last observed ThreadState is still
// returned alongside it so callers can report the last known status.
var ErrWaitTimeout = errors.New("timed out waiting for run to finish")

// WaitForRun polls the thread until its current run reaches a stop state
// (terminal, or waiting_input), the context is cancelled, or the deadline
// elapses. It returns the last observed ThreadState even on error so callers
// can surface the last known status. pollInterval and timeout fall back to the
// package defaults when non-positive; a timeout <= 0 means wait indefinitely
// (until ctx is done).
func WaitForRun(ctx context.Context, c *Client, threadID string, pollInterval, timeout time.Duration) (*ThreadState, error) {
	if pollInterval <= 0 {
		pollInterval = DefaultPollInterval
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var last *ThreadState
	for {
		st, err := c.RetrieveThread(ctx, threadID)
		if err != nil {
			// A context error during the request means we hit the deadline or
			// were cancelled; prefer the friendlier timeout error with the last
			// known state.
			if ctxErr := ctx.Err(); ctxErr != nil && last != nil {
				return last, timeoutOrCancel(ctxErr)
			}
			return last, err
		}
		last = st
		if st.Run != nil && isStop(st.Run.Status) {
			return st, nil
		}

		select {
		case <-ctx.Done():
			return last, timeoutOrCancel(ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

func timeoutOrCancel(ctxErr error) error {
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return ErrWaitTimeout
	}
	return ctxErr
}
