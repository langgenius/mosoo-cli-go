package agentapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/langgenius/mosoo-cli-go/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	appListQuery = `query appList($organizationId: ULID!) { appList(organizationId: $organizationId) { id name } }`

	createAppMutation = `mutation createApp($input: CreateAppInput!) { createApp(input: $input) { id name } }`

	createAgentMutation = `mutation createAgent($input: CreateAgentInput!) { createAgent(input: $input) { id name appId kind status } }`

	publishAgentMutation = `mutation publishAgent($input: PublishAgentInput!) { publishAgent(input: $input) { id liveVersion { id agentId isLive versionNumber } } }`
)

type provisionOptions struct {
	file     string
	writeEnv string
	dryRun   bool
	json     bool
}

type smokeOptions struct {
	agentID        string
	file           string
	idempotencyKey string
	wait           bool
	json           bool
}

type writeEnvOptions struct {
	appID   string
	agentID string
	out     string
}

type appSpec struct {
	App    appConfig     `json:"app" yaml:"app"`
	Agents []agentConfig `json:"agents" yaml:"agents"`
}

type appConfig struct {
	ID             string `json:"id" yaml:"id"`
	Name           string `json:"name" yaml:"name"`
	OrganizationID string `json:"organizationId" yaml:"organizationId"`
}

type agentConfig struct {
	Key         string   `json:"key" yaml:"key"`
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Kind        string   `json:"kind" yaml:"kind"`
	RuntimeID   string   `json:"runtimeId" yaml:"runtimeId"`
	Provider    string   `json:"provider" yaml:"provider"`
	Model       string   `json:"model" yaml:"model"`
	Prompt      string   `json:"prompt" yaml:"prompt"`
	SkillIDs    []string `json:"skillIds" yaml:"skillIds"`
	Publish     *bool    `json:"publish" yaml:"publish"`
}

type provisionResult struct {
	DryRun        bool             `json:"dryRun"`
	Target        string           `json:"target"`
	ConsoleHost   string           `json:"consoleHost"`
	PublicHost    string           `json:"publicThreadHost"`
	App           provisionApp     `json:"app"`
	Agents        []provisionAgent `json:"agents"`
	OperatorSteps []string         `json:"operatorSteps,omitempty"`
	Warnings      []string         `json:"warnings,omitempty"`
	Next          []string         `json:"next,omitempty"`
}

type provisionApp struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Action string `json:"action"`
}

type provisionAgent struct {
	Key       string `json:"key,omitempty"`
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Action    string `json:"action"`
	Published bool   `json:"published"`
}

type smokeResult struct {
	AgentID     string         `json:"agentId"`
	ThreadID    string         `json:"threadId,omitempty"`
	RunID       string         `json:"runId,omitempty"`
	RunStatus   string         `json:"runStatus,omitempty"`
	Waited      bool           `json:"waited"`
	Next        []string       `json:"next,omitempty"`
	RawResponse map[string]any `json:"rawResponse,omitempty"`
}

type graphQLClient struct {
	host    string
	opts    latheruntime.ClientOptions
	secrets []string
}

// NewCommand returns product-level App provisioning and smoke-test commands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent-app",
		Short: "Provision and smoke-test Mosoo Agent apps",
		Long:  "Provision and smoke-test Mosoo Agent apps by orchestrating existing Console and Public Thread API surfaces. Lower-level token export, thread wait, transcript, and upload helpers remain gated until their product primitives are available.",
	}

	provisionOpts := &provisionOptions{}
	provisionCmd := &cobra.Command{
		Use:   "provision",
		Short: "Create or reuse an App, create Agents, and publish them",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProvision(cmd, provisionOpts)
		},
	}
	provisionCmd.Flags().StringVar(&provisionOpts.file, "file", "", "Agent app YAML or JSON spec")
	provisionCmd.Flags().StringVar(&provisionOpts.writeEnv, "write-env", "", "Env file to write after provisioning; gated until GitHub issue #15 lands")
	provisionCmd.Flags().BoolVar(&provisionOpts.dryRun, "dry-run", false, "Show the planned workflow without remote changes")
	provisionCmd.Flags().BoolVar(&provisionOpts.json, "json", false, "Print machine-readable JSON")
	_ = provisionCmd.MarkFlagRequired("file")

	smokeOpts := &smokeOptions{}
	smokeCmd := &cobra.Command{
		Use:   "smoke-test",
		Short: "Create a Public Thread smoke test for a published Agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSmokeTest(cmd, smokeOpts)
		},
	}
	smokeCmd.Flags().StringVar(&smokeOpts.agentID, "agent-id", "", "Published Agent API endpoint ID")
	smokeCmd.Flags().StringVar(&smokeOpts.file, "file", "", "JSON or YAML public Thread create request body")
	smokeCmd.Flags().StringVar(&smokeOpts.idempotencyKey, "idempotency-key", "", "Idempotency-Key header for thread creation")
	smokeCmd.Flags().BoolVar(&smokeOpts.wait, "wait", false, "Wait for final output; gated until GitHub issue #17 lands")
	smokeCmd.Flags().BoolVar(&smokeOpts.json, "json", false, "Print machine-readable JSON")
	_ = smokeCmd.MarkFlagRequired("agent-id")
	_ = smokeCmd.MarkFlagRequired("file")

	writeEnvOpts := &writeEnvOptions{}
	writeEnvCmd := &cobra.Command{
		Use:   "write-env",
		Short: "Write app env variables for Public Thread apps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWriteEnv(cmd, writeEnvOpts)
		},
	}
	writeEnvCmd.Flags().StringVar(&writeEnvOpts.appID, "app-id", "", "Mosoo App ID")
	writeEnvCmd.Flags().StringVar(&writeEnvOpts.agentID, "agent-id", "", "Published Agent API endpoint ID")
	writeEnvCmd.Flags().StringVar(&writeEnvOpts.out, "out", "", "Env file path")
	_ = writeEnvCmd.MarkFlagRequired("app-id")
	_ = writeEnvCmd.MarkFlagRequired("agent-id")
	_ = writeEnvCmd.MarkFlagRequired("out")

	cmd.AddCommand(provisionCmd, smokeCmd, writeEnvCmd)
	return cmd
}

func runProvision(cmd *cobra.Command, opts *provisionOptions) error {
	if strings.TrimSpace(opts.writeEnv) != "" {
		return envWriterUnavailableError(opts.writeEnv)
	}

	spec, err := readSpec(opts.file)
	if err != nil {
		return err
	}
	if err := validateSpec(spec); err != nil {
		return err
	}
	resolved, err := target.ResolveFromCommand(cmd)
	if err != nil {
		return err
	}
	result := provisionResult{
		DryRun:      opts.dryRun,
		Target:      resolved.Target,
		ConsoleHost: resolved.Hosts[target.SurfaceConsole],
		PublicHost:  resolved.Hosts[target.SurfacePublicThreadAPI],
		App:         provisionApp{ID: spec.App.ID, Name: spec.App.Name, Action: "planned"},
	}
	result.OperatorSteps = operatorSteps(opts.file, spec)
	result.Warnings = dependencyWarnings()

	for _, agent := range spec.Agents {
		result.Agents = append(result.Agents, provisionAgent{
			Key:       agent.Key,
			ID:        agent.ID,
			Name:      agent.Name,
			Action:    "planned",
			Published: shouldPublish(agent),
		})
	}

	if opts.dryRun {
		if wantsJSON(cmd, opts.json) {
			return writeJSON(cmd, result)
		}
		printProvisionHuman(cmd, result)
		return nil
	}

	client, err := newGraphQLClient(cmd, resolved.Hosts[target.SurfaceConsole], "console")
	if err != nil {
		return err
	}
	appID, appAction, err := ensureApp(cmd.Context(), client, spec.App)
	if err != nil {
		return redactError(err, client.secrets)
	}
	result.App.ID = appID
	result.App.Action = appAction

	result.Agents = result.Agents[:0]
	for _, agent := range spec.Agents {
		agentID := strings.TrimSpace(agent.ID)
		action := "reused"
		if agentID == "" {
			createdID, err := createAgent(cmd.Context(), client, appID, agent)
			if err != nil {
				return redactError(fmt.Errorf("partial provisioning failure after app %s was %s; no cleanup was attempted: create agent %q: %w", appID, appAction, agent.Name, err), client.secrets)
			}
			agentID = createdID
			action = "created"
		}
		published := false
		if shouldPublish(agent) {
			if err := publishAgent(cmd.Context(), client, appID, agentID); err != nil {
				return redactError(fmt.Errorf("partial provisioning failure after agent %q was %s as %s; no cleanup was attempted: publish agent %q (%s): %w", agent.Name, action, agentID, agent.Name, agentID, err), client.secrets)
			}
			published = true
		}
		result.Agents = append(result.Agents, provisionAgent{
			Key:       agent.Key,
			ID:        agentID,
			Name:      agent.Name,
			Action:    action,
			Published: published,
		})
	}
	result.Next = nextProvisionCommands(result)

	if wantsJSON(cmd, opts.json) {
		return writeJSON(cmd, result)
	}
	printProvisionHuman(cmd, result)
	return nil
}

func runSmokeTest(cmd *cobra.Command, opts *smokeOptions) error {
	if opts.wait {
		return threadWaitUnavailableError(opts.agentID, opts.file)
	}
	body, err := readRequestBody(opts.file)
	if err != nil {
		return err
	}
	resolved, err := target.ResolveFromCommand(cmd)
	if err != nil {
		return err
	}
	clientOpts, secrets, err := loadHostOptions(cmd, resolved.Hosts[target.SurfacePublicThreadAPI], "public-thread-api")
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.idempotencyKey) != "" {
		if clientOpts.Headers == nil {
			clientOpts.Headers = map[string]string{}
		}
		clientOpts.Headers["Idempotency-Key"] = strings.TrimSpace(opts.idempotencyKey)
	}
	path := "/agents/" + url.PathEscape(strings.TrimSpace(opts.agentID)) + "/threads"
	data, err := latheruntime.DoRaw(cmd.Context(), resolved.Hosts[target.SurfacePublicThreadAPI], http.MethodPost, path, body, clientOpts)
	if err != nil {
		return redactError(fmt.Errorf("create public thread: %w", err), secrets)
	}
	raw, err := decodeObject(data)
	if err != nil {
		return err
	}
	result := smokeResult{
		AgentID:     opts.agentID,
		ThreadID:    firstStringAt(raw, []string{"thread", "id"}, []string{"id"}),
		RunID:       firstStringAt(raw, []string{"run", "id"}, []string{"thread", "last_run_id"}),
		RunStatus:   firstStringAt(raw, []string{"run", "status"}),
		Waited:      false,
		RawResponse: raw,
	}
	if result.ThreadID != "" {
		result.Next = []string{
			fmt.Sprintf("mosoo public-thread-api events list-events --thread-id %s", result.ThreadID),
			fmt.Sprintf("mosoo agent-app smoke-test --agent-id %s --file %s --wait", opts.agentID, opts.file),
		}
	}
	if wantsJSON(cmd, opts.json) {
		return writeJSON(cmd, result)
	}
	printSmokeHuman(cmd, result)
	return nil
}

func runWriteEnv(_ *cobra.Command, opts *writeEnvOptions) error {
	return envWriterUnavailableError(opts.out)
}

func readSpec(path string) (appSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return appSpec{}, fmt.Errorf("read %s: %w", path, err)
	}
	var spec appSpec
	if err := unmarshalStructured(data, &spec); err != nil {
		return appSpec{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return spec, nil
}

func readRequestBody(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var body map[string]any
	if err := unmarshalStructured(data, &body); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return body, nil
}

func unmarshalStructured(data []byte, out any) error {
	if err := json.Unmarshal(data, out); err == nil {
		return nil
	}
	return yaml.Unmarshal(data, out)
}

func validateSpec(spec appSpec) error {
	if strings.TrimSpace(spec.App.ID) == "" {
		if strings.TrimSpace(spec.App.Name) == "" {
			return fmt.Errorf("app.name is required when app.id is not set")
		}
		if strings.TrimSpace(spec.App.OrganizationID) == "" {
			return fmt.Errorf("app.organizationId is required when creating or reusing by name")
		}
	}
	if len(spec.Agents) == 0 {
		return fmt.Errorf("agents must contain at least one Agent")
	}
	seen := map[string]struct{}{}
	for i, agent := range spec.Agents {
		label := agent.Name
		if label == "" {
			label = fmt.Sprintf("#%d", i+1)
		}
		if strings.TrimSpace(agent.Key) != "" {
			if _, ok := seen[agent.Key]; ok {
				return fmt.Errorf("agents[%d].key %q is duplicated", i, agent.Key)
			}
			seen[agent.Key] = struct{}{}
		}
		if strings.TrimSpace(agent.ID) != "" {
			continue
		}
		for field, value := range map[string]string{
			"name":      agent.Name,
			"kind":      agent.Kind,
			"runtimeId": agent.RuntimeID,
			"provider":  agent.Provider,
			"model":     agent.Model,
			"prompt":    agent.Prompt,
		} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("agent %s requires %s when agent.id is not set", label, field)
			}
		}
		if agent.Kind != "pet" && agent.Kind != "cattle" {
			return fmt.Errorf("agent %s kind must be one of pet, cattle", label)
		}
	}
	return nil
}

func ensureApp(ctx context.Context, client graphQLClient, app appConfig) (string, string, error) {
	if strings.TrimSpace(app.ID) != "" {
		return strings.TrimSpace(app.ID), "reused", nil
	}
	existing, err := findAppByName(ctx, client, app.OrganizationID, app.Name)
	if err != nil {
		return "", "", err
	}
	if existing != "" {
		return existing, "reused", nil
	}
	raw, err := client.post(ctx, createAppMutation, map[string]any{
		"input": map[string]any{
			"name":           strings.TrimSpace(app.Name),
			"organizationId": strings.TrimSpace(app.OrganizationID),
		},
	})
	if err != nil {
		return "", "", err
	}
	node, ok := objectAt(raw, "data", "createApp")
	if !ok {
		return "", "", fmt.Errorf("response missing data.createApp")
	}
	id := stringValue(node["id"])
	if id == "" {
		return "", "", fmt.Errorf("response missing data.createApp.id")
	}
	return id, "created", nil
}

func findAppByName(ctx context.Context, client graphQLClient, organizationID, name string) (string, error) {
	raw, err := client.post(ctx, appListQuery, map[string]any{"organizationId": strings.TrimSpace(organizationID)})
	if err != nil {
		return "", err
	}
	apps := objectListAt(raw, "data", "appList")
	for _, app := range apps {
		if stringValue(app["name"]) == strings.TrimSpace(name) {
			return stringValue(app["id"]), nil
		}
	}
	return "", nil
}

func createAgent(ctx context.Context, client graphQLClient, appID string, agent agentConfig) (string, error) {
	input := map[string]any{
		"appId":     strings.TrimSpace(appID),
		"kind":      strings.TrimSpace(agent.Kind),
		"model":     strings.TrimSpace(agent.Model),
		"name":      strings.TrimSpace(agent.Name),
		"prompt":    strings.TrimSpace(agent.Prompt),
		"provider":  strings.TrimSpace(agent.Provider),
		"runtimeId": strings.TrimSpace(agent.RuntimeID),
		"skillIds":  append([]string(nil), agent.SkillIDs...),
	}
	if strings.TrimSpace(agent.Description) != "" {
		input["description"] = strings.TrimSpace(agent.Description)
	}
	raw, err := client.post(ctx, createAgentMutation, map[string]any{"input": input})
	if err != nil {
		return "", err
	}
	node, ok := objectAt(raw, "data", "createAgent")
	if !ok {
		return "", fmt.Errorf("response missing data.createAgent")
	}
	id := firstStringAt(node, []string{"id"}, []string{"agentId"})
	if id == "" {
		return "", fmt.Errorf("response missing data.createAgent.id")
	}
	return id, nil
}

func publishAgent(ctx context.Context, client graphQLClient, appID, agentID string) error {
	raw, err := client.post(ctx, publishAgentMutation, map[string]any{
		"input": map[string]any{
			"appId":   strings.TrimSpace(appID),
			"agentId": strings.TrimSpace(agentID),
		},
	})
	if err != nil {
		return err
	}
	if _, ok := objectAt(raw, "data", "publishAgent"); !ok {
		return fmt.Errorf("response missing data.publishAgent")
	}
	return nil
}

func newGraphQLClient(cmd *cobra.Command, hostname, label string) (graphQLClient, error) {
	opts, secrets, err := loadHostOptions(cmd, hostname, label)
	if err != nil {
		return graphQLClient{}, err
	}
	return graphQLClient{host: hostname, opts: opts, secrets: secrets}, nil
}

func (c graphQLClient) post(ctx context.Context, query string, variables map[string]any) (map[string]any, error) {
	data, err := latheruntime.DoRaw(ctx, c.host, http.MethodPost, "/graphql", map[string]any{
		"query":     query,
		"variables": variables,
	}, c.opts)
	if err != nil {
		return nil, err
	}
	raw, err := decodeObject(data)
	if err != nil {
		return nil, err
	}
	if errorsValue, ok := raw["errors"]; ok {
		encoded, _ := json.Marshal(errorsValue)
		return nil, fmt.Errorf("graphql errors: %s", string(encoded))
	}
	return raw, nil
}

func loadHostOptions(cmd *cobra.Command, hostname, label string) (latheruntime.ClientOptions, []string, error) {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return latheruntime.ClientOptions{}, nil, fmt.Errorf("no hostname resolved for %s", label)
	}
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		return latheruntime.ClientOptions{}, nil, err
	}
	entry, ok := hosts.Get(hostname)
	if !ok {
		return latheruntime.ClientOptions{}, nil, fmt.Errorf("not authenticated to %s host %s; run `mosoo auth login --hostname %s`", label, hostname, hostname)
	}
	auth, err := latheruntime.NewAuthFromHost(entry)
	if err != nil {
		return latheruntime.ClientOptions{}, nil, err
	}
	insecure := entry.Insecure
	if flag := cmd.Root().PersistentFlags().Lookup("insecure"); flag != nil {
		value, _ := cmd.Root().PersistentFlags().GetBool("insecure")
		insecure = insecure || value
	}
	debug := false
	if flag := cmd.Root().PersistentFlags().Lookup("debug"); flag != nil {
		debug, _ = cmd.Root().PersistentFlags().GetBool("debug")
	}
	opts := latheruntime.ClientOptions{
		Auth:      auth,
		Insecure:  insecure,
		Debug:     debug,
		Timeout:   30 * time.Second,
		UserAgent: cmd.Root().Use,
		// Agent app provisioning issues non-idempotent mutations; do not retry
		// create/publish calls unless a lower-level primitive owns that policy.
		MaxRetries: -1,
	}
	return opts, hostSecrets(entry), nil
}

func hostSecrets(entry latheconfig.HostEntry) []string {
	var secrets []string
	for _, value := range []string{entry.OAuthToken, entry.APIKey, entry.BasicPassword} {
		if strings.TrimSpace(value) != "" {
			secrets = append(secrets, value)
		}
	}
	return secrets
}

func shouldPublish(agent agentConfig) bool {
	return agent.Publish == nil || *agent.Publish
}

func decodeObject(data []byte) (map[string]any, error) {
	var raw map[string]any
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response JSON: %w", err)
	}
	return raw, nil
}

func objectAt(value any, path ...string) (map[string]any, bool) {
	current := value
	for _, key := range path {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current = obj[key]
	}
	obj, ok := current.(map[string]any)
	return obj, ok
}

func objectListAt(value any, path ...string) []map[string]any {
	current := value
	for _, key := range path {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = obj[key]
	}
	return coerceObjectList(current)
}

func coerceObjectList(value any) []map[string]any {
	switch v := value.(type) {
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if obj, ok := item.(map[string]any); ok {
				out = append(out, obj)
			}
		}
		return out
	case map[string]any:
		for _, key := range []string{"items", "nodes", "apps", "data"} {
			if child, ok := v[key]; ok {
				if out := coerceObjectList(child); len(out) > 0 {
					return out
				}
			}
		}
	}
	return nil
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	default:
		return ""
	}
}

func firstStringAt(value any, paths ...[]string) string {
	for _, path := range paths {
		current := value
		for _, key := range path {
			obj, ok := current.(map[string]any)
			if !ok {
				current = nil
				break
			}
			current = obj[key]
		}
		if got := stringValue(current); got != "" {
			return got
		}
	}
	return ""
}

func wantsJSON(cmd *cobra.Command, flag bool) bool {
	if flag {
		return true
	}
	if cmd.Root() != nil && cmd.Root().PersistentFlags().Lookup("output") != nil {
		format, _ := cmd.Root().PersistentFlags().GetString("output")
		return format == "json"
	}
	return false
}

func writeJSON(cmd *cobra.Command, value any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func printProvisionHuman(cmd *cobra.Command, result provisionResult) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "target: %s\n", result.Target)
	fmt.Fprintf(out, "app: %s", result.App.Action)
	if result.App.ID != "" {
		fmt.Fprintf(out, " %s", result.App.ID)
	}
	if result.App.Name != "" {
		fmt.Fprintf(out, " (%s)", result.App.Name)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "agents:")
	for _, agent := range result.Agents {
		fmt.Fprintf(out, "  - %s", agent.Action)
		if agent.ID != "" {
			fmt.Fprintf(out, " %s", agent.ID)
		}
		if agent.Name != "" {
			fmt.Fprintf(out, " (%s)", agent.Name)
		}
		if agent.Published {
			fmt.Fprint(out, " published")
		}
		fmt.Fprintln(out)
	}
	if len(result.Warnings) > 0 {
		fmt.Fprintln(out, "warnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintf(out, "  - %s\n", warning)
		}
	}
	if len(result.Next) > 0 {
		fmt.Fprintln(out, "next:")
		for _, step := range result.Next {
			fmt.Fprintf(out, "  - %s\n", step)
		}
	}
}

func printSmokeHuman(cmd *cobra.Command, result smokeResult) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "agentId: %s\n", result.AgentID)
	if result.ThreadID != "" {
		fmt.Fprintf(out, "threadId: %s\n", result.ThreadID)
	}
	if result.RunID != "" {
		fmt.Fprintf(out, "runId: %s\n", result.RunID)
	}
	if result.RunStatus != "" {
		fmt.Fprintf(out, "runStatus: %s\n", result.RunStatus)
	}
	if len(result.Next) > 0 {
		fmt.Fprintln(out, "next:")
		for _, step := range result.Next {
			fmt.Fprintf(out, "  - %s\n", step)
		}
	}
}

func operatorSteps(file string, spec appSpec) []string {
	agentRef := "<published-agent-id>"
	for _, agent := range spec.Agents {
		if strings.TrimSpace(agent.ID) != "" {
			agentRef = strings.TrimSpace(agent.ID)
			break
		}
		if strings.TrimSpace(agent.Key) != "" {
			agentRef = "<" + strings.TrimSpace(agent.Key) + "-agent-id>"
			break
		}
	}
	return []string{
		"mosoo doctor --json",
		fmt.Sprintf("mosoo agent-app provision --file %s", file),
		fmt.Sprintf("mosoo agent-app write-env --app-id <app-id> --agent-id %s --out .dev.vars", agentRef),
		fmt.Sprintf("mosoo agent-app smoke-test --agent-id %s --file smoke.json --wait", agentRef),
	}
}

func dependencyWarnings() []string {
	return []string{
		"env writing is gated until GitHub issue #15 lands; this command will not print raw tokens or duplicate token export internals",
		"public Thread wait/final-output/transcript is gated until GitHub issue #17 lands; use generated thread create/list-events commands for low-level debugging",
		"file upload setup is gated until GitHub issue #25 lands when smoke tests require files",
	}
}

func nextProvisionCommands(result provisionResult) []string {
	appID := result.App.ID
	agentID := "<published-agent-id>"
	for _, agent := range result.Agents {
		if agent.ID != "" && agent.Published {
			agentID = agent.ID
			break
		}
	}
	return []string{
		fmt.Sprintf("mosoo agent-app write-env --app-id %s --agent-id %s --out .dev.vars", appID, agentID),
		fmt.Sprintf("mosoo agent-app smoke-test --agent-id %s --file smoke.json --wait", agentID),
	}
}

func envWriterUnavailableError(out string) error {
	if strings.TrimSpace(out) == "" {
		out = ".dev.vars"
	}
	return fmt.Errorf("agent-app env writing requires the token/env export primitive from GitHub issue #15 before this workflow can write %s; run without --write-env for create/publish only or use the existing generated access-token commands manually", out)
}

func threadWaitUnavailableError(agentID, file string) error {
	return fmt.Errorf("agent-app smoke-test --wait requires the public Thread wait/final-output/transcript primitive from GitHub issue #17; current low-level fallback is `mosoo public-thread-api threads create --agent-id %s --file %s` followed by `mosoo public-thread-api events list-events --thread-id <thread-id>`", agentID, file)
}

func redactError(err error, secrets []string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s", redactSecrets(err.Error(), secrets))
}

func redactSecrets(value string, secrets []string) string {
	out := value
	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "[REDACTED]")
	}
	return out
}
