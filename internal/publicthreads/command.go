package publicthreads

import (
	"context"
	"fmt"
	"time"

	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
	"github.com/spf13/cobra"
)

// Install mounts the hand-maintained Public Thread API usability helpers onto
// the generated public-thread-api command tree:
//
//   - threads create     replaced with a --wait / --final-output aware version
//   - threads transcript added
//   - events wait        added
//
// It is a no-op-safe extension: it only augments the thread/run/event read
// surface and never touches file upload behavior.
func Install(root *cobra.Command) error {
	surface := findChild(root, "public-thread-api")
	if surface == nil {
		return fmt.Errorf("public-thread-api command tree is not mounted")
	}
	threads := findChild(surface, "threads")
	if threads == nil {
		return fmt.Errorf("public-thread-api threads command tree is not mounted")
	}
	events := findChild(surface, "events")
	if events == nil {
		return fmt.Errorf("public-thread-api events command tree is not mounted")
	}

	if existing := findChild(threads, "create"); existing != nil {
		threads.RemoveCommand(existing)
	}
	threads.AddCommand(newCreateCommand())

	if existing := findChild(threads, "transcript"); existing != nil {
		threads.RemoveCommand(existing)
	}
	threads.AddCommand(newTranscriptCommand())

	if existing := findChild(events, "wait"); existing != nil {
		events.RemoveCommand(existing)
	}
	events.AddCommand(newWaitCommand())
	return nil
}

type waitFlags struct {
	timeout      time.Duration
	pollInterval time.Duration
	finalOutput  bool
}

func addWaitFlags(cmd *cobra.Command, f *waitFlags, withFinalOutput bool) {
	cmd.Flags().DurationVar(&f.timeout, "timeout", DefaultWaitTimeout, "Maximum time to wait for the run to finish (0 = wait indefinitely)")
	cmd.Flags().DurationVar(&f.pollInterval, "poll-interval", DefaultPollInterval, "How often to poll thread state while waiting")
	if withFinalOutput {
		cmd.Flags().BoolVar(&f.finalOutput, "final-output", false, "On success, print only the completed run's final output text (implies --wait)")
	}
}

func newCreateCommand() *cobra.Command {
	var (
		agentID        string
		file           string
		sets           []string
		stringSets     []string
		idempotencyKey string
		wait           bool
		wf             waitFlags
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a thread for an agent",
		Long: "Create a new thread against an agent API endpoint.\n\n" +
			"With --wait, block until the initial run reaches a terminal state and report the outcome. " +
			"With --final-output, print only the completed run's final output text (implies --wait). " +
			"On failure, the run status, run error, tool failures, and last relevant events are shown.",
		Example: "mosoo public-thread-api threads create --agent-id <agent-id> --file body.json --wait --final-output\n",
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := buildCreateBody(file, sets, stringSets)
			if err != nil {
				return err
			}
			client, err := NewClient(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			st, err := client.CreateThread(ctx, agentID, body, idempotencyKey)
			if err != nil {
				return err
			}

			doWait := wait || wf.finalOutput
			format := outputFormat(cmd)
			if !doWait {
				// Preserve the generated command's behavior: print the full
				// create response in whatever format the user selected.
				return latheruntime.FormatOutput(st.Raw, format, cmd.OutOrStdout(), latheruntime.OutputHints{})
			}

			if st.Run == nil {
				// An empty thread (no input) has no run to wait for.
				if isStructured(format) {
					return renderResultStructured(cmd.OutOrStdout(), format, st, nil)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Created empty thread %s; no run was queued to wait for.\n", st.Thread.ID)
				return nil
			}

			final, waitErr := WaitForRun(ctx, client, st.Thread.ID, wf.pollInterval, wf.timeout)
			return finishWait(cmd, client, final, waitErr, wf.finalOutput)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&agentID, "agent-id", "", "Agent API Endpoint ID from the Agent's API Access panel. v1 IDs are bare ULIDs. (path, required, ulid)")
	flags.StringVarP(&file, "file", "f", "", "path to JSON body file, or '-' for stdin")
	flags.StringArrayVar(&sets, "set", nil, "set body field with type inference, e.g. --set input.type=user.message (repeatable; nested via dots)")
	flags.StringArrayVar(&stringSets, "set-str", nil, "set body field as string (repeatable; nested via dots)")
	flags.StringVar(&idempotencyKey, "idempotency-key", "", "Optional key for retry-safe create-thread calls. (header)")
	flags.BoolVar(&wait, "wait", false, "Block until the initial run reaches a terminal state")
	addWaitFlags(cmd, &wf, true)
	_ = cmd.MarkFlagRequired("agent-id")
	return cmd
}

func newWaitCommand() *cobra.Command {
	var (
		threadID string
		wf       waitFlags
	)
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for a thread's run to reach a terminal state",
		Long: "Poll a thread until its current run completes, fails, is cancelled/expired, or pauses waiting for input.\n\n" +
			"On success the final output is available; on failure the run status, run error, tool failures, and last " +
			"relevant events are shown so you can understand why the run did not complete cleanly.",
		Example: "mosoo public-thread-api events wait --thread-id <thread-id> --final-output\n",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := NewClient(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			// Pre-check so an empty thread fails fast instead of polling to the
			// timeout.
			st, err := client.RetrieveThread(ctx, threadID)
			if err != nil {
				return err
			}
			if st.Run == nil {
				if isStructured(outputFormat(cmd)) {
					return renderResultStructured(cmd.OutOrStdout(), outputFormat(cmd), st, nil)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Thread %s has no run to wait for.\n", st.Thread.ID)
				return nil
			}
			if isStop(st.Run.Status) {
				return finishWait(cmd, client, st, nil, wf.finalOutput)
			}

			final, waitErr := WaitForRun(ctx, client, threadID, wf.pollInterval, wf.timeout)
			return finishWait(cmd, client, final, waitErr, wf.finalOutput)
		},
	}
	cmd.Flags().StringVar(&threadID, "thread-id", "", "Thread ID returned by create thread. v1 IDs are bare ULIDs. (required, ulid)")
	addWaitFlags(cmd, &wf, true)
	_ = cmd.MarkFlagRequired("thread-id")
	return cmd
}

func newTranscriptCommand() *cobra.Command {
	var (
		threadID        string
		runID           string
		limit           int
		includeThinking bool
	)
	cmd := &cobra.Command{
		Use:   "transcript",
		Short: "Print a thread's event transcript",
		Long: "Fetch the latest thread events and render them as a readable, chronological transcript.\n\n" +
			"The public event surface exposes references rather than raw message text, so this is a structured " +
			"timeline of the conversation and tool activity. Use -o json for the raw events.",
		Example: "mosoo public-thread-api threads transcript --thread-id <thread-id>\n",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := NewClient(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			list, err := client.ListEvents(ctx, threadID, limit)
			if err != nil {
				return err
			}
			events := list.Events
			if runID != "" {
				events = eventsForRunStrict(events, runID)
			}

			format := outputFormat(cmd)
			if isStructured(format) {
				return renderEventsStructured(cmd.OutOrStdout(), format, events, list.Truncated)
			}
			if list.Truncated {
				fmt.Fprintln(cmd.OutOrStdout(), "(older events were truncated; raise --limit to see more)")
			}
			writeTranscript(cmd.OutOrStdout(), events, includeThinking)
			return nil
		},
	}
	cmd.Flags().StringVar(&threadID, "thread-id", "", "Thread ID returned by create thread. v1 IDs are bare ULIDs. (required, ulid)")
	cmd.Flags().StringVar(&runID, "run-id", "", "Only include events for this run ID")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of latest thread events to fetch")
	cmd.Flags().BoolVar(&includeThinking, "include-thinking", false, "Include agent thinking events in the transcript")
	_ = cmd.MarkFlagRequired("thread-id")
	return cmd
}

// finishWait renders the outcome of a wait and returns a non-nil error when the
// run did not complete cleanly (so the process exit code reflects the failure).
func finishWait(cmd *cobra.Command, client *Client, st *ThreadState, waitErr error, finalOutputOnly bool) error {
	format := outputFormat(cmd)

	// A timeout still returns the last observed state; report it without
	// pretending the run finished.
	if waitErr != nil {
		if st == nil || st.Run == nil {
			return waitErr
		}
		if isStructured(format) {
			_ = renderResultStructured(cmd.OutOrStdout(), format, st, nil)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Stopped waiting on run %s (last status: %s): %v\n", st.Run.ID, st.Run.Status, waitErr)
		}
		return waitErr
	}
	if st == nil || st.Run == nil {
		return fmt.Errorf("no run state available after wait")
	}

	run := st.Run
	if isFailure(run.Status) {
		events := fetchFailureEvents(cmd.Context(), client, st.Thread.ID)
		if isStructured(format) {
			if err := renderResultStructured(cmd.OutOrStdout(), format, st, events); err != nil {
				return err
			}
		} else {
			writeFailureReport(cmd.OutOrStdout(), st, events)
		}
		return fmt.Errorf("run %s %s", run.ID, run.Status)
	}

	if isStructured(format) {
		return renderResultStructured(cmd.OutOrStdout(), format, st, nil)
	}
	switch {
	case run.Status == StatusWaitingInput:
		writePausedSummary(cmd.OutOrStdout(), st)
	case finalOutputOnly:
		writeFinalOutput(cmd.OutOrStdout(), run)
	default:
		writeSuccessSummary(cmd.OutOrStdout(), st)
	}
	return nil
}

func fetchFailureEvents(ctx context.Context, client *Client, threadID string) []Event {
	if ctx == nil {
		ctx = context.Background()
	}
	list, err := client.ListEvents(ctx, threadID, failureEventLimit)
	if err != nil || list == nil {
		return nil
	}
	return list.Events
}

func outputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Root().PersistentFlags().GetString("output")
	if format == "" {
		return "table"
	}
	return format
}

// isStructured reports whether the format should emit the raw response document
// rather than a human-readable summary.
func isStructured(format string) bool {
	switch format {
	case "json", "yaml", "raw":
		return true
	default:
		return false
	}
}

func findChild(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
