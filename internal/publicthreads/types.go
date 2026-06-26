package publicthreads

import "net/url"

// ThreadState mirrors the shared shape of CreateThreadResponse and
// RetrieveThreadResponse: the thread summary plus its most recent run.
type ThreadState struct {
	Thread ThreadSummary `json:"thread"`
	Run    *RunSummary   `json:"run"`
	Links  any           `json:"links,omitempty"`
	// Raw holds the original response bytes for faithful -o json/yaml output.
	Raw []byte `json:"-"`
}

// ThreadSummary is the subset of the public ThreadSummary the helpers read.
type ThreadSummary struct {
	ID        string  `json:"id"`
	AgentID   string  `json:"agent_id"`
	Status    string  `json:"status"`
	Title     *string `json:"title"`
	LastRunID *string `json:"last_run_id"`
}

// RunSummary is the subset of the public RunSummary the helpers read.
type RunSummary struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	Trigger     string          `json:"trigger"`
	Error       *RunError       `json:"error"`
	FinalOutput *RunFinalOutput `json:"finalOutput"`
	StartedAt   *string         `json:"startedAt"`
	CompletedAt *string         `json:"completedAt"`
}

// RunError is the public-safe failure summary on a failed run.
type RunError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// RunFinalOutput is the reconstructed final assistant answer for a completed run.
type RunFinalOutput struct {
	Text string `json:"text"`
}

// EventList is a page of thread events.
type EventList struct {
	Events    []Event `json:"events"`
	Truncated bool    `json:"truncated"`
}

// Event is a single public thread event log entry.
type Event struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Content    string  `json:"content"`
	Status     string  `json:"status"`
	RunID      *string `json:"runId"`
	OccurredAt string  `json:"occurredAt"`
	DurationMs *int64  `json:"durationMs"`
	Tokens     *int64  `json:"tokens"`
}

// Run status classification. The public API exposes:
//
//	queued, booting          -> pre-execution
//	running                  -> active
//	waiting_input            -> active but paused for user input
//	completed                -> terminal success
//	failed, cancelled, expired -> terminal failure
const (
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
	StatusCancelled    = "cancelled"
	StatusExpired      = "expired"
	StatusWaitingInput = "waiting_input"
)

// isTerminal reports whether a run will not change state on its own.
func isTerminal(status string) bool {
	switch status {
	case StatusCompleted, StatusFailed, StatusCancelled, StatusExpired:
		return true
	default:
		return false
	}
}

// isStop reports whether waiting should stop: a terminal state, or a run paused
// waiting for input (which will not progress without a send-events call).
func isStop(status string) bool {
	return isTerminal(status) || status == StatusWaitingInput
}

// isFailure reports whether a terminal status represents an unsuccessful run.
func isFailure(status string) bool {
	switch status {
	case StatusFailed, StatusCancelled, StatusExpired:
		return true
	default:
		return false
	}
}

func pathEscape(s string) string {
	return url.PathEscape(s)
}
