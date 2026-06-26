package publicthreads

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallReplacesGeneratedCreateAndAddsHelpers(t *testing.T) {
	root := &cobra.Command{Use: "mosoo"}
	surface := &cobra.Command{Use: "public-thread-api"}
	threads := &cobra.Command{Use: "threads"}
	events := &cobra.Command{Use: "events"}
	generatedCreate := &cobra.Command{Use: "create", Short: "generated"}
	threads.AddCommand(generatedCreate)
	surface.AddCommand(threads)
	surface.AddCommand(events)
	root.AddCommand(surface)

	if err := Install(root); err != nil {
		t.Fatalf("Install: %v", err)
	}
	create := findChild(threads, "create")
	if create == nil || create == generatedCreate {
		t.Fatal("generated create was not replaced")
	}
	if findChild(threads, "transcript") == nil {
		t.Fatal("transcript was not mounted")
	}
	if findChild(events, "wait") == nil {
		t.Fatal("wait was not mounted")
	}
}

func TestInstallErrorsWithoutSurface(t *testing.T) {
	root := &cobra.Command{Use: "mosoo"}
	if err := Install(root); err == nil {
		t.Fatal("expected error when public-thread-api tree is missing")
	}
}

// threadServer simulates the public thread API. retrieve flips the run status
// from running to whatever finalStatus describes after flipAfter polls.
type threadServer struct {
	finalBody  string
	flipAfter  int32
	retrieves  int32
	eventsBody string
}

func (s *threadServer) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/agents/"):
			_, _ = w.Write([]byte(`{"thread":{"id":"t1","status":"RUNNING"},"run":{"id":"r1","status":"queued"},"links":{"thread":"u"}}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/events"):
			_, _ = w.Write([]byte(s.eventsBody))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/threads/"):
			n := atomic.AddInt32(&s.retrieves, 1)
			if n <= s.flipAfter {
				_, _ = w.Write([]byte(`{"thread":{"id":"t1","status":"RUNNING"},"run":{"id":"r1","status":"running"},"links":{"thread":"u"}}`))
				return
			}
			_, _ = w.Write([]byte(s.finalBody))
		default:
			http.Error(w, `{"error":{"code":"not_found","message":"no route"}}`, http.StatusNotFound)
		}
	})
	return mux
}

func runArgs(host string, args ...string) []string {
	return append([]string{"--hostname", host}, args...)
}

func TestCreateWaitFinalOutput(t *testing.T) {
	s := &threadServer{
		flipAfter: 1,
		finalBody: `{"thread":{"id":"t1","status":"IDLE"},"run":{"id":"r1","status":"completed","finalOutput":{"text":"Hello from the agent."}},"links":{"thread":"u"}}`,
	}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	root, out := newTestRoot(t, srv.URL)
	root.SetArgs(runArgs(srv.URL, "public-thread-api", "threads", "create", "--poll-interval", "1ms",
		"--agent-id", "agent1", "--set", "input.type=user.message", "--final-output"))
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.TrimSpace(out.String()) != "Hello from the agent." {
		t.Fatalf("output = %q, want just the final output text", out.String())
	}
}

func TestCreateWaitFailureReport(t *testing.T) {
	s := &threadServer{
		flipAfter: 0,
		finalBody: `{"thread":{"id":"t1","status":"IDLE"},"run":{"id":"r1","status":"failed","error":{"code":"tool_error","message":"the search tool failed","retryable":true}},"links":{"thread":"u"}}`,
		eventsBody: `{"events":[` +
			`{"id":"e1","type":"run.started","content":"r1","status":"available","runId":"r1","occurredAt":"2026-06-26T00:00:00Z"},` +
			`{"id":"e2","type":"tool.use.completed","content":"search","status":"error","runId":"r1","occurredAt":"2026-06-26T00:00:01Z","durationMs":1200},` +
			`{"id":"e3","type":"run.failed","content":"r1","status":"available","runId":"r1","occurredAt":"2026-06-26T00:00:02Z"}` +
			`],"truncated":false}`,
	}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	root, out := newTestRoot(t, srv.URL)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs(runArgs(srv.URL, "public-thread-api", "threads", "create", "--poll-interval", "1ms",
		"--agent-id", "agent1", "--set", "input.type=user.message", "--wait"))
	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-nil error for failed run")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Fatalf("error = %v, want to mention failed", err)
	}
	got := out.String()
	for _, want := range []string{"did not complete cleanly", "status: failed", "tool_error", "the search tool failed", "Tool failures", "tool.use.completed"} {
		if !strings.Contains(got, want) {
			t.Fatalf("failure report missing %q\n--- output ---\n%s", want, got)
		}
	}
}

func TestEventsWaitJSON(t *testing.T) {
	s := &threadServer{
		flipAfter: 0,
		finalBody: `{"thread":{"id":"t1","status":"IDLE"},"run":{"id":"r1","status":"completed","finalOutput":{"text":"ok"}},"links":{"thread":"u"}}`,
	}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	root, out := newTestRoot(t, srv.URL)
	root.SetArgs(runArgs(srv.URL, "-o", "json", "public-thread-api", "events", "wait", "--poll-interval", "1ms", "--thread-id", "t1"))
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, out.String())
	}
	run, _ := doc["run"].(map[string]any)
	if run["status"] != "completed" {
		t.Fatalf("run.status = %#v", run["status"])
	}
}

func TestEventsWaitNoRun(t *testing.T) {
	s := &threadServer{finalBody: `{"thread":{"id":"t1","status":"IDLE"},"run":null,"links":{"thread":"u"}}`}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	root, out := newTestRoot(t, srv.URL)
	root.SetArgs(runArgs(srv.URL, "public-thread-api", "events", "wait", "--poll-interval", "1ms", "--thread-id", "t1"))
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "no run to wait for") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTranscriptRendersTimeline(t *testing.T) {
	s := &threadServer{
		eventsBody: `{"events":[` +
			`{"id":"e1","type":"user.message","content":"msg-1","status":"available","runId":"r1","occurredAt":"2026-06-26T00:00:00Z"},` +
			`{"id":"e2","type":"agent.message.delta","content":"msg-2","status":"available","runId":"r1","occurredAt":"2026-06-26T00:00:01Z"},` +
			`{"id":"e3","type":"agent.message.delta","content":"msg-2","status":"available","runId":"r1","occurredAt":"2026-06-26T00:00:02Z"},` +
			`{"id":"e4","type":"run.completed","content":"r1","status":"available","runId":"r1","occurredAt":"2026-06-26T00:00:03Z","durationMs":4200}` +
			`],"truncated":false}`,
	}
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	root, out := newTestRoot(t, srv.URL)
	root.SetArgs(runArgs(srv.URL, "public-thread-api", "threads", "transcript", "--thread-id", "t1"))
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := out.String()
	for _, want := range []string{"user", "assistant (x2)", "run completed"} {
		if !strings.Contains(got, want) {
			t.Fatalf("transcript missing %q\n--- output ---\n%s", want, got)
		}
	}
}

func TestRetrieveDecodesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":{"code":"not_found","message":"Thread not found for this caller."}}`, http.StatusNotFound)
	}))
	defer srv.Close()

	root, out := newTestRoot(t, srv.URL)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.SetArgs(runArgs(srv.URL, "public-thread-api", "events", "wait", "--poll-interval", "1ms", "--thread-id", "missing"))
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, output=%q", out.String())
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v (%T), want *APIError", err, err)
	}
	if apiErr.Code != "not_found" {
		t.Fatalf("code = %q", apiErr.Code)
	}
}
