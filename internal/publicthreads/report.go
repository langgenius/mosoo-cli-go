package publicthreads

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
)

// failureEventLimit bounds how many recent events are fetched to explain a
// failed run.
const failureEventLimit = 50

// lastEventsShown bounds how many trailing events the failure report prints.
const lastEventsShown = 10

// finalOutputText returns the completed run's final output text, if present.
func finalOutputText(run *RunSummary) string {
	if run == nil || run.FinalOutput == nil {
		return ""
	}
	return run.FinalOutput.Text
}

// renderResultStructured emits the thread state (optionally enriched with the
// events gathered for a failure) in the requested structured format
// (json/yaml/raw), reusing the shared Lathe formatter for consistency.
func renderResultStructured(w io.Writer, format string, st *ThreadState, events []Event) error {
	// Start from the raw response to preserve every server field, then attach
	// the events we fetched for failure context under a sibling key.
	var doc map[string]any
	if len(st.Raw) == 0 || json.Unmarshal(st.Raw, &doc) != nil {
		b, _ := json.Marshal(st)
		_ = json.Unmarshal(b, &doc)
	}
	if doc == nil {
		doc = map[string]any{}
	}
	if events != nil {
		b, _ := json.Marshal(events)
		var ev any
		_ = json.Unmarshal(b, &ev)
		doc["events"] = ev
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return latheruntime.FormatOutput(data, format, w, latheruntime.OutputHints{})
}

// renderEventsStructured emits the event list (and truncation flag) in the
// requested structured format.
func renderEventsStructured(w io.Writer, format string, events []Event, truncated bool) error {
	if events == nil {
		events = []Event{}
	}
	data, err := json.Marshal(map[string]any{"events": events, "truncated": truncated})
	if err != nil {
		return err
	}
	return latheruntime.FormatOutput(data, format, w, latheruntime.OutputHints{ListPath: "events"})
}

// eventsForRunStrict keeps only events tagged with runID, returning an empty
// slice when none match (used for an explicit --run-id filter).
func eventsForRunStrict(events []Event, runID string) []Event {
	out := make([]Event, 0, len(events))
	for _, e := range events {
		if e.RunID != nil && *e.RunID == runID {
			out = append(out, e)
		}
	}
	return out
}

// writeFinalOutput prints just the completed run's final output text.
func writeFinalOutput(w io.Writer, run *RunSummary) {
	text := finalOutputText(run)
	if text == "" {
		return
	}
	fmt.Fprintln(w, text)
}

// writeSuccessSummary prints a human-readable completion summary.
func writeSuccessSummary(w io.Writer, st *ThreadState) {
	run := st.Run
	fmt.Fprintf(w, "Run %s completed.\n", run.ID)
	if text := finalOutputText(run); text != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, text)
	}
}

// writePausedSummary prints a human-readable summary for a run paused awaiting
// input.
func writePausedSummary(w io.Writer, st *ThreadState) {
	fmt.Fprintf(w, "Run %s is waiting for input (status: %s).\n", st.Run.ID, st.Run.Status)
	fmt.Fprintln(w, "Send a follow-up with `mosoo public-thread-api events send` to continue.")
}

// writeFailureReport explains why a run did not complete cleanly: its status,
// the structured run error, any tool failures, and the last relevant events.
func writeFailureReport(w io.Writer, st *ThreadState, events []Event) {
	run := st.Run
	fmt.Fprintf(w, "Run %s did not complete cleanly (status: %s).\n", run.ID, run.Status)

	if run.Error != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Error: [%s] %s\n", run.Error.Code, run.Error.Message)
		fmt.Fprintf(w, "Retryable: %t\n", run.Error.Retryable)
	}

	runEvents := eventsForRun(events, run.ID)
	if tools := toolFailures(runEvents); len(tools) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Tool failures:")
		for _, e := range tools {
			fmt.Fprintf(w, "  - %s %s%s\n", e.Type, e.Content, durationSuffix(e))
		}
	}

	last := lastEvents(runEvents, lastEventsShown)
	if len(last) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Last %d events:\n", len(last))
		for _, e := range last {
			fmt.Fprintf(w, "  %s  %-26s %s\n", e.OccurredAt, e.Type, e.Content)
		}
	}
}

// eventsForRun keeps events scoped to runID. Events with no runId are dropped
// so the report shows only this run's activity; if none match (e.g. the API did
// not tag events), the original slice is returned so context is not lost.
func eventsForRun(events []Event, runID string) []Event {
	out := make([]Event, 0, len(events))
	for _, e := range events {
		if e.RunID != nil && *e.RunID == runID {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		return events
	}
	return out
}

// toolFailures returns events that represent a failed or errored tool use.
func toolFailures(events []Event) []Event {
	var out []Event
	for _, e := range events {
		if strings.HasPrefix(e.Type, "tool.") && e.Status == "error" {
			out = append(out, e)
		}
	}
	return out
}

// lastEvents returns the trailing n events in chronological order.
func lastEvents(events []Event, n int) []Event {
	if len(events) <= n {
		return events
	}
	return events[len(events)-n:]
}

func durationSuffix(e Event) string {
	if e.DurationMs == nil {
		return ""
	}
	return fmt.Sprintf(" (%dms)", *e.DurationMs)
}
