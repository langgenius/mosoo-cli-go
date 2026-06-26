package publicthreads

import (
	"fmt"
	"io"
	"strings"
)

// writeTranscript renders thread events as a readable, chronological timeline.
//
// The public event surface exposes references (such as message IDs) rather than
// raw model text — by design it never leaks transcripts or diagnostics — so this
// is a structured timeline of what happened, not a reconstruction of the full
// message bodies. Consecutive deltas of the same kind are collapsed into a
// single line with a count so the output stays scannable.
func writeTranscript(w io.Writer, events []Event, includeThinking bool) {
	if len(events) == 0 {
		fmt.Fprintln(w, "(no events)")
		return
	}

	i := 0
	for i < len(events) {
		e := events[i]
		switch e.Type {
		case "agent.thinking.delta":
			run := i + countRun(events[i:], e.Type)
			if includeThinking {
				fmt.Fprintf(w, "%s  thinking%s\n", e.OccurredAt, deltaCount(run-i))
			}
			i = run
			continue
		case "agent.message.delta":
			run := i + countRun(events[i:], e.Type)
			fmt.Fprintf(w, "%s  assistant%s %s\n", e.OccurredAt, deltaCount(run-i), e.Content)
			i = run
			continue
		}

		switch e.Type {
		case "user.message":
			fmt.Fprintf(w, "%s  user           %s\n", e.OccurredAt, e.Content)
		case "tool.use.started":
			fmt.Fprintf(w, "%s  tool ->        %s\n", e.OccurredAt, e.Content)
		case "tool.use.completed":
			fmt.Fprintf(w, "%s  tool <-        %s%s%s\n", e.OccurredAt, statusMark(e), e.Content, durationSuffix(e))
		case "tool.confirmation.required":
			fmt.Fprintf(w, "%s  tool ?         %s\n", e.OccurredAt, e.Content)
		case "run.started":
			fmt.Fprintf(w, "%s  --- run started (%s) ---\n", e.OccurredAt, runRef(e))
		case "run.completed":
			fmt.Fprintf(w, "%s  --- run completed (%s)%s ---\n", e.OccurredAt, runRef(e), durationSuffix(e))
		case "run.failed":
			fmt.Fprintf(w, "%s  --- run failed (%s) ---\n", e.OccurredAt, runRef(e))
		default:
			fmt.Fprintf(w, "%s  %-14s %s\n", e.OccurredAt, e.Type, e.Content)
		}
		i++
	}
}

// countRun counts how many leading events share the same type as events[0].
func countRun(events []Event, typ string) int {
	n := 0
	for n < len(events) && events[n].Type == typ {
		n++
	}
	return n
}

func deltaCount(n int) string {
	if n <= 1 {
		return ""
	}
	return fmt.Sprintf(" (x%d)", n)
}

func statusMark(e Event) string {
	if e.Status == "error" {
		return "FAILED "
	}
	return ""
}

func runRef(e Event) string {
	if e.RunID != nil && *e.RunID != "" {
		return "run " + *e.RunID
	}
	return strings.TrimSpace(e.Content)
}
