package publicthreads

import (
	"context"
	"errors"
	"testing"
	"time"

	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
)

// fakeClient builds a Client whose transport returns the given JSON bodies in
// sequence (the last is repeated once exhausted).
func fakeClient(bodies ...string) *Client {
	i := 0
	return &Client{
		transport: func(_ context.Context, _, _ string, _ any, _ map[string]string) (*latheruntime.RawResult, error) {
			b := bodies[i]
			if i < len(bodies)-1 {
				i++
			}
			return &latheruntime.RawResult{Body: []byte(b), StatusCode: 200}, nil
		},
	}
}

func TestWaitForRunStopsOnCompleted(t *testing.T) {
	c := fakeClient(
		`{"thread":{"id":"t1","status":"RUNNING"},"run":{"id":"r1","status":"running"}}`,
		`{"thread":{"id":"t1","status":"RUNNING"},"run":{"id":"r1","status":"running"}}`,
		`{"thread":{"id":"t1","status":"IDLE"},"run":{"id":"r1","status":"completed","finalOutput":{"text":"done"}}}`,
	)
	st, err := WaitForRun(context.Background(), c, "t1", time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("WaitForRun: %v", err)
	}
	if st.Run.Status != StatusCompleted {
		t.Fatalf("status = %q", st.Run.Status)
	}
	if finalOutputText(st.Run) != "done" {
		t.Fatalf("finalOutput = %q", finalOutputText(st.Run))
	}
}

func TestWaitForRunStopsOnWaitingInput(t *testing.T) {
	c := fakeClient(`{"thread":{"id":"t1"},"run":{"id":"r1","status":"waiting_input"}}`)
	st, err := WaitForRun(context.Background(), c, "t1", time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("WaitForRun: %v", err)
	}
	if st.Run.Status != StatusWaitingInput {
		t.Fatalf("status = %q", st.Run.Status)
	}
}

func TestWaitForRunTimeout(t *testing.T) {
	c := fakeClient(`{"thread":{"id":"t1"},"run":{"id":"r1","status":"running"}}`)
	st, err := WaitForRun(context.Background(), c, "t1", time.Millisecond, 20*time.Millisecond)
	if !errors.Is(err, ErrWaitTimeout) {
		t.Fatalf("err = %v, want ErrWaitTimeout", err)
	}
	if st == nil || st.Run.Status != "running" {
		t.Fatalf("expected last state with running status, got %#v", st)
	}
}

func TestWaitForRunFailureStops(t *testing.T) {
	c := fakeClient(`{"thread":{"id":"t1"},"run":{"id":"r1","status":"failed","error":{"code":"boom","message":"kaboom","retryable":false}}}`)
	st, err := WaitForRun(context.Background(), c, "t1", time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("WaitForRun: %v", err)
	}
	if !isFailure(st.Run.Status) {
		t.Fatalf("status = %q, want failure", st.Run.Status)
	}
	if st.Run.Error == nil || st.Run.Error.Code != "boom" {
		t.Fatalf("run error = %#v", st.Run.Error)
	}
}
