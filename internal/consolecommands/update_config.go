package consolecommands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
	"github.com/spf13/cobra"
)

const updateAgentConfigMutation = `mutation updateAgentConfig($input: UpdateAgentConfigInput!) { updateAgentConfig(input: $input) { createdAt description id kind liveVersion { agentId createdAt createdByAccountId environmentId id isLive kind model provider runtimeId summary versionNumber } model name prompt provider runtimeId skills { ownerName skillId skillName state } status updatedAt visibility appId } }`

// Install mounts hand-maintained replacements for commands that Lathe cannot
// currently express correctly through generated specs.
func Install(root *cobra.Command) error {
	console := findChild(root, "console")
	if console == nil {
		return fmt.Errorf("console command tree is not mounted")
	}
	agents := findChild(console, "agents")
	if agents == nil {
		return fmt.Errorf("console agents command tree is not mounted")
	}
	if existing := findChild(agents, "update-config"); existing != nil {
		agents.RemoveCommand(existing)
	}
	agents.AddCommand(newUpdateConfigCommand())
	return nil
}

type updateConfigOptions struct {
	agentID            string
	appID              string
	description        string
	environmentID      string
	kind               string
	mcpServerIDs       []string
	model              string
	name               string
	prompt             string
	provider           string
	providerOptionsRaw string
	runtimeID          string
	skillIDs           []string
}

func newUpdateConfigCommand() *cobra.Command {
	var opts updateConfigOptions
	cmd := &cobra.Command{
		Use:     "update-config",
		Aliases: []string{"update-agent-config"},
		Short:   "Update an agent config",
		Long:    "Update Agent config through the raw Console GraphQL API. Prefer `mosoo agent manifest apply`, which fetches remote state, preserves omitted fields, shows a field-level diff, and supports `--dry-run`.",
		Example: "mosoo console agents update-config --input-provider-options '{}' -o json",
		Hidden:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			input, err := opts.input(cmd)
			if err != nil {
				return err
			}
			body := map[string]any{
				"query": updateAgentConfigMutation,
				"variables": map[string]any{
					"input": input,
				},
			}
			hostname, clientOpts, err := latheruntime.LoadHostOptions(cmd)
			if err != nil {
				return err
			}
			clientOpts.UserAgent = cmd.Root().Use
			if debug, err := cmd.Root().PersistentFlags().GetBool("debug"); err == nil && debug {
				clientOpts.Debug = true
			}
			data, err := latheruntime.DoRaw(context.Background(), hostname, http.MethodPost, "/graphql", body, clientOpts)
			if err != nil {
				return err
			}
			format, _ := cmd.Root().PersistentFlags().GetString("output")
			return latheruntime.FormatOutput(data, format, os.Stdout, latheruntime.OutputHints{})
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.agentID, "input-agent-id", "", "input.agentId (variable, required)")
	flags.StringVar(&opts.description, "input-description", "", "input.description (variable)")
	flags.StringVar(&opts.environmentID, "input-environment-environment-id", "", "input.environment.environmentId (variable)")
	flags.StringVar(&opts.kind, "input-kind", "", "input.kind (variable, required, one of: pet|cattle)")
	flags.StringSliceVar(&opts.mcpServerIDs, "input-mcp-server-ids", nil, "input.mcpServerIds (variable, required)")
	flags.StringVar(&opts.model, "input-model", "", "input.model (variable, required)")
	flags.StringVar(&opts.name, "input-name", "", "input.name (variable, required)")
	flags.StringVar(&opts.prompt, "input-prompt", "", "input.prompt (variable, required)")
	flags.StringVar(&opts.provider, "input-provider", "", "input.provider (variable, required)")
	flags.StringVar(&opts.providerOptionsRaw, "input-provider-options", "", "input.providerOptions JSON object (variable, required)")
	flags.StringVar(&opts.runtimeID, "input-runtime-id", "", "input.runtimeId (variable, required)")
	flags.StringSliceVar(&opts.skillIDs, "input-skill-ids", nil, "input.skillIds (variable, required)")
	flags.StringVar(&opts.appID, "input-app-id", "", "input.appId (variable, required)")
	for _, name := range []string{
		"input-agent-id",
		"input-kind",
		"input-mcp-server-ids",
		"input-model",
		"input-name",
		"input-prompt",
		"input-provider",
		"input-provider-options",
		"input-runtime-id",
		"input-skill-ids",
		"input-app-id",
	} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}

func (o updateConfigOptions) input(cmd *cobra.Command) (map[string]any, error) {
	if o.kind != "pet" && o.kind != "cattle" {
		return nil, fmt.Errorf("invalid value %q for --input-kind: must be one of pet, cattle", o.kind)
	}
	providerOptions, err := parseJSONObjectFlag("input-provider-options", o.providerOptionsRaw)
	if err != nil {
		return nil, err
	}
	environment := map[string]any{}
	if cmd.Flags().Changed("input-environment-environment-id") {
		environment["environmentId"] = o.environmentID
	}
	input := map[string]any{
		"agentId":         o.agentID,
		"appId":           o.appID,
		"environment":     environment,
		"kind":            o.kind,
		"mcpServerIds":    o.mcpServerIDs,
		"model":           o.model,
		"name":            o.name,
		"prompt":          o.prompt,
		"provider":        o.provider,
		"providerOptions": providerOptions,
		"runtimeId":       o.runtimeID,
		"skillIds":        o.skillIDs,
	}
	if cmd.Flags().Changed("input-description") {
		input["description"] = o.description
	}
	return input, nil
}

func parseJSONObjectFlag(flag string, raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("--%s must be a JSON object", flag)
	}
	var value any
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid --%s JSON object: %w", flag, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("invalid --%s JSON object: %w", flag, err)
		}
		return nil, fmt.Errorf("invalid --%s JSON object: trailing JSON token", flag)
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("--%s must be a JSON object", flag)
	}
	return obj, nil
}

func findChild(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
