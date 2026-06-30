package agentmanifest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/langgenius/mosoo-connector/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	agentManifestQuery = `query agentManifest($appId: ULID!, $agentId: ULID!) { agentManifest(appId: $appId, agentId: $agentId) { agentId json yaml } }`

	updateAgentConfigMutation = `mutation updateAgentConfig($input: UpdateAgentConfigInput!) { updateAgentConfig(input: $input) { createdAt description id kind liveVersion { agentId createdAt createdByAccountId environmentId id isLive kind model provider runtimeId summary versionNumber } model name prompt provider runtimeId skills { ownerName skillId skillName state } status updatedAt visibility appId } }`
)

var requiredUpdateFields = []string{
	"agentId",
	"appId",
	"kind",
	"mcpServerIds",
	"model",
	"name",
	"prompt",
	"provider",
	"providerOptions",
	"runtimeId",
	"skillIds",
}

var allowedManifestFields = map[string]struct{}{
	"advanced":        {},
	"agentId":         {},
	"appId":           {},
	"builtInTools":    {},
	"description":     {},
	"environment":     {},
	"environmentId":   {},
	"id":              {},
	"kind":            {},
	"manifestVersion": {},
	"mcpServerIds":    {},
	"mcpServers":      {},
	"metadata":        {},
	"model":           {},
	"name":            {},
	"prompt":          {},
	"prompts":         {},
	"provider":        {},
	"providerOptions": {},
	"runtime":         {},
	"runtimeId":       {},
	"skillIds":        {},
	"skills":          {},
	"sourceAgentId":   {},
}

type commandOptions struct {
	appID   string
	agentID string
	file    string
	out     string
	dryRun  bool
	json    bool
}

type remoteManifest struct {
	AppID    string         `json:"appId"`
	AgentID  string         `json:"agentId"`
	Manifest map[string]any `json:"manifest"`
	YAML     string         `json:"yaml,omitempty"`
}

type change struct {
	Path   string `json:"path"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type diffResult struct {
	AppID   string   `json:"appId"`
	AgentID string   `json:"agentId"`
	Changes []change `json:"changes"`
}

type applyResult struct {
	AppID          string         `json:"appId"`
	AgentID        string         `json:"agentId"`
	DryRun         bool           `json:"dryRun"`
	Changes        []change       `json:"changes"`
	UpdateResponse map[string]any `json:"updateResponse,omitempty"`
}

// NewCommand returns hand-written product workflow commands above the generated
// Console GraphQL command catalog.
func NewCommand() *cobra.Command {
	opts := &commandOptions{}
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Operate Mosoo Agents",
	}
	manifestCmd := &cobra.Command{
		Use:   "manifest",
		Short: "Inspect and apply Agent manifests",
		Long:  "Inspect and apply Agent manifest YAML while preserving remote Agent config fields that the local file does not explicitly change.",
	}
	manifestCmd.PersistentFlags().StringVar(&opts.appID, "app-id", "", "Mosoo App ID")
	manifestCmd.PersistentFlags().StringVar(&opts.agentID, "agent-id", "", "Mosoo Agent ID")
	manifestCmd.PersistentFlags().BoolVar(&opts.json, "json", false, "Print machine-readable JSON")

	probeCmd := &cobra.Command{
		Use:     "probe",
		Aliases: []string{"pull"},
		Short:   "Fetch the current remote Agent manifest",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProbe(cmd, opts)
		},
	}
	probeCmd.Flags().StringVar(&opts.out, "out", "", "Write the remote manifest YAML to this file")

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Show local manifest changes against current remote state",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDiff(cmd, opts)
		},
	}
	diffCmd.Flags().StringVar(&opts.file, "file", "", "Local Agent manifest YAML file")

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a local Agent manifest after fetching and merging remote state",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runApply(cmd, opts)
		},
	}
	applyCmd.Flags().StringVar(&opts.file, "file", "", "Local Agent manifest YAML file")
	applyCmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show the planned field-level changes without writing remote state")

	manifestCmd.AddCommand(probeCmd, diffCmd, applyCmd)
	agentCmd.AddCommand(manifestCmd, newEnvCommand())
	return agentCmd
}

func runProbe(cmd *cobra.Command, opts *commandOptions) error {
	appID, agentID, err := requireIDs(opts.appID, opts.agentID)
	if err != nil {
		return err
	}
	remote, err := fetchRemoteManifest(cmd.Context(), cmd, appID, agentID)
	if err != nil {
		return err
	}
	if opts.out != "" {
		if remote.YAML == "" {
			return fmt.Errorf("remote manifest did not include yaml output")
		}
		if err := os.WriteFile(opts.out, []byte(ensureTrailingNewline(remote.YAML)), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", opts.out, err)
		}
		if wantsJSON(cmd, opts) {
			return writeJSON(cmd, map[string]any{
				"appId":   remote.AppID,
				"agentId": remote.AgentID,
				"out":     opts.out,
			})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", opts.out)
		return nil
	}
	if wantsJSON(cmd, opts) {
		return writeJSON(cmd, remote)
	}
	if remote.YAML != "" {
		_, err := fmt.Fprint(cmd.OutOrStdout(), ensureTrailingNewline(remote.YAML))
		return err
	}
	return writeYAML(cmd, remote.Manifest)
}

func runDiff(cmd *cobra.Command, opts *commandOptions) error {
	if opts.file == "" {
		return fmt.Errorf("--file is required")
	}
	localManifest, err := readManifestFile(opts.file)
	if err != nil {
		return err
	}
	appID, agentID, err := resolveIDs(opts.appID, opts.agentID, localManifest)
	if err != nil {
		return err
	}
	remote, err := fetchRemoteManifest(cmd.Context(), cmd, appID, agentID)
	if err != nil {
		return err
	}
	changes, _, err := planManifestUpdate(remote.Manifest, localManifest, appID, agentID)
	if err != nil {
		return err
	}
	result := diffResult{AppID: appID, AgentID: agentID, Changes: changes}
	if wantsJSON(cmd, opts) {
		return writeJSON(cmd, result)
	}
	printChanges(cmd, "Manifest changes", changes)
	return nil
}

func runApply(cmd *cobra.Command, opts *commandOptions) error {
	if opts.file == "" {
		return fmt.Errorf("--file is required")
	}
	localManifest, err := readManifestFile(opts.file)
	if err != nil {
		return err
	}
	appID, agentID, err := resolveIDs(opts.appID, opts.agentID, localManifest)
	if err != nil {
		return err
	}
	remote, err := fetchRemoteManifest(cmd.Context(), cmd, appID, agentID)
	if err != nil {
		return err
	}
	changes, finalInput, err := planManifestUpdate(remote.Manifest, localManifest, appID, agentID)
	if err != nil {
		return err
	}
	if err := validateUpdateInput(finalInput); err != nil {
		return err
	}
	result := applyResult{AppID: appID, AgentID: agentID, DryRun: opts.dryRun, Changes: changes}
	if opts.dryRun || len(changes) == 0 {
		if wantsJSON(cmd, opts) {
			return writeJSON(cmd, result)
		}
		printChanges(cmd, "Manifest changes", changes)
		if opts.dryRun {
			fmt.Fprintln(cmd.OutOrStdout(), "dry run: no remote changes written")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "no remote changes written")
		}
		return nil
	}
	response, err := updateAgentConfig(cmd.Context(), cmd, finalInput)
	if err != nil {
		return err
	}
	result.UpdateResponse = response
	if wantsJSON(cmd, opts) {
		return writeJSON(cmd, result)
	}
	printChanges(cmd, "Applied manifest changes", changes)
	return nil
}

func fetchRemoteManifest(ctx context.Context, cmd *cobra.Command, appID, agentID string) (remoteManifest, error) {
	host, clientOpts, err := consoleClientOptions(cmd)
	if err != nil {
		return remoteManifest{}, err
	}
	body := map[string]any{
		"query": agentManifestQuery,
		"variables": map[string]any{
			"appId":   appID,
			"agentId": agentID,
		},
	}
	raw, err := doGraphQL(ctx, host, body, clientOpts)
	if err != nil {
		return remoteManifest{}, err
	}
	manifestNode, ok := objectAt(raw, "data", "agentManifest")
	if !ok {
		return remoteManifest{}, fmt.Errorf("response missing data.agentManifest")
	}
	remote := remoteManifest{
		AppID:   appID,
		AgentID: stringValue(manifestNode["agentId"], agentID),
		YAML:    stringValue(manifestNode["yaml"], ""),
	}
	if remote.AgentID == "" {
		remote.AgentID = agentID
	}
	manifest, err := parseManifestValue(manifestNode["json"], remote.YAML)
	if err != nil {
		return remoteManifest{}, err
	}
	remote.Manifest = manifest
	return remote, nil
}

func updateAgentConfig(ctx context.Context, cmd *cobra.Command, input map[string]any) (map[string]any, error) {
	host, clientOpts, err := consoleClientOptions(cmd)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"query": updateAgentConfigMutation,
		"variables": map[string]any{
			"input": input,
		},
	}
	return doGraphQL(ctx, host, body, clientOpts)
}

func doGraphQL(ctx context.Context, hostname string, body map[string]any, opts latheruntime.ClientOptions) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	raw, err := latheruntime.DoRaw(ctx, hostname, http.MethodPost, "/graphql", body, opts)
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("decode GraphQL response: %w", err)
	}
	if errorsValue, ok := decoded["errors"]; ok {
		return nil, fmt.Errorf("graphql error: %s", compactJSON(errorsValue))
	}
	return decoded, nil
}

func consoleClientOptions(cmd *cobra.Command) (string, latheruntime.ClientOptions, error) {
	resolved, err := target.ResolveFromCommand(cmd)
	if err != nil {
		return "", latheruntime.ClientOptions{}, err
	}
	host := resolved.Hosts[target.SurfaceConsole]
	if host == "" {
		return "", latheruntime.ClientOptions{}, fmt.Errorf("no console host resolved")
	}
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		return "", latheruntime.ClientOptions{}, err
	}
	entry, ok := hosts.Get(host)
	if !ok {
		return "", latheruntime.ClientOptions{}, fmt.Errorf("not authenticated to %s (run: mosoo auth login --hostname %s)", host, host)
	}
	auth, err := latheruntime.NewAuthFromHost(entry)
	if err != nil {
		return "", latheruntime.ClientOptions{}, err
	}
	opts := latheruntime.ClientOptions{
		Auth:      auth,
		Insecure:  entry.Insecure,
		UserAgent: cmd.Root().Use,
	}
	if v, err := cmd.Root().PersistentFlags().GetBool("insecure"); err == nil && v {
		opts.Insecure = true
	}
	return host, opts, nil
}

func planManifestUpdate(remoteManifestMap, localManifest map[string]any, appID, agentID string) ([]change, map[string]any, error) {
	remoteInput := updateInputFromManifest(remoteManifestMap)
	ensureUpdateIDs(remoteInput, appID, agentID)
	localPatchSource, err := patchSource(localManifest)
	if err != nil {
		return nil, nil, err
	}
	localPatch := updateInputFromManifest(localPatchSource)
	localPatch, err = omitRedactedPatchValues(localPatch)
	if err != nil {
		return nil, nil, err
	}
	finalInput := mergeMaps(remoteInput, localPatch)
	ensureUpdateIDs(finalInput, appID, agentID)
	changes := diffValues(remoteInput, finalInput)
	return changes, finalInput, nil
}

func ensureUpdateIDs(input map[string]any, appID, agentID string) {
	if appID != "" {
		input["appId"] = appID
	}
	if agentID != "" {
		input["agentId"] = agentID
	}
}

func readManifestFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	manifest, err := parseYAMLMap(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return manifest, nil
}

func parseManifestValue(jsonValue any, yamlValue string) (map[string]any, error) {
	if jsonValue != nil {
		switch value := jsonValue.(type) {
		case map[string]any:
			return value, nil
		case string:
			if strings.TrimSpace(value) != "" {
				return parseJSONMap([]byte(value))
			}
		default:
			raw, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("marshal manifest json: %w", err)
			}
			return parseJSONMap(raw)
		}
	}
	if strings.TrimSpace(yamlValue) == "" {
		return nil, fmt.Errorf("remote manifest included neither json nor yaml")
	}
	return parseYAMLMap([]byte(yamlValue))
}

func parseJSONMap(data []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse manifest json: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("manifest must be a JSON object")
	}
	return out, nil
}

func parseYAMLMap(data []byte) (map[string]any, error) {
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	normalized, ok := normalizeYAMLValue(raw).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("manifest must be a YAML object")
	}
	return normalized, nil
}

func normalizeYAMLValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			out[key] = normalizeYAMLValue(value)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			out[fmt.Sprint(key)] = normalizeYAMLValue(value)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, value := range v {
			out[i] = normalizeYAMLValue(value)
		}
		return out
	default:
		return v
	}
}

func patchSource(manifest map[string]any) (map[string]any, error) {
	if spec, ok := objectAt(manifest, "spec"); ok {
		if err := validateTopLevelManifest(manifest); err != nil {
			return nil, err
		}
		return spec, validatePatchFields(spec)
	}
	patch := make(map[string]any, len(manifest))
	for key, value := range manifest {
		switch key {
		case "apiVersion", "status":
			continue
		default:
			patch[key] = value
		}
	}
	return patch, validatePatchFields(patch)
}

func validateTopLevelManifest(manifest map[string]any) error {
	for key := range manifest {
		switch key {
		case "apiVersion", "kind", "metadata", "spec", "status":
		default:
			return fmt.Errorf("unknown top-level manifest field %q", key)
		}
	}
	return nil
}

func validatePatchFields(patch map[string]any) error {
	for key := range patch {
		if _, ok := allowedManifestFields[key]; !ok {
			return fmt.Errorf("unknown manifest spec field %q", key)
		}
	}
	return nil
}

func updateInputFromManifest(manifest map[string]any) map[string]any {
	source := manifest
	if spec, ok := objectAt(manifest, "spec"); ok {
		source = spec
	}
	out := map[string]any{}
	copyIfPresent(out, source, "agentId")
	if _, ok := out["agentId"]; !ok {
		copyRenameIfPresent(out, source, "id", "agentId")
	}
	if _, ok := out["agentId"]; !ok {
		copyRenameIfPresent(out, source, "sourceAgentId", "agentId")
	}
	copyIfPresent(out, source, "appId")
	if metadata, ok := objectAt(source, "metadata"); ok {
		copyIfPresent(out, metadata, "description")
		copyIfPresent(out, metadata, "name")
	}
	copyIfPresent(out, source, "description")
	copyIfPresent(out, source, "kind")
	copyIfPresent(out, source, "mcpServerIds")
	copyIfPresent(out, source, "model")
	copyIfPresent(out, source, "name")
	copyPromptIfPresent(out, source, "prompt")
	copyIfPresent(out, source, "provider")
	copyIfPresent(out, source, "providerOptions")
	copyIfPresent(out, source, "runtimeId")
	copyIfPresent(out, source, "skillIds")
	copyIfPresent(out, source, "builtInTools")
	if prompts, ok := objectAt(source, "prompts"); ok {
		copyPromptRenameIfPresent(out, prompts, "system", "prompt")
	}
	if runtime, ok := objectAt(source, "runtime"); ok {
		copyRenameIfPresent(out, runtime, "id", "runtimeId")
		copyIfPresent(out, runtime, "model")
		copyIfPresent(out, runtime, "provider")
		copyRenameIfPresent(out, runtime, "settings", "providerOptions")
		copyIfPresent(out, runtime, "providerOptions")
	}
	if environment, ok := objectAt(source, "environment"); ok {
		if environmentID, ok := environment["environmentId"]; ok {
			out["environment"] = map[string]any{"environmentId": environmentID}
		} else {
			out["environment"] = environment
		}
	} else if environmentID, ok := source["environmentId"]; ok {
		out["environment"] = map[string]any{"environmentId": environmentID}
	} else if environmentID, ok := valueAt(source, "liveVersion", "environmentId"); ok {
		out["environment"] = map[string]any{"environmentId": environmentID}
	}
	if _, ok := out["skillIds"]; !ok {
		if ids, ok := idsFromObjectArray(source["skills"], "skillId"); ok {
			out["skillIds"] = ids
		}
	}
	if _, ok := out["mcpServerIds"]; !ok {
		if ids, ok := idsFromObjectArray(source["mcpBindings"], "serverId"); ok {
			out["mcpServerIds"] = ids
		}
	}
	if _, ok := out["mcpServerIds"]; !ok {
		if ids, ok := idsFromObjectArray(source["mcpServers"], "serverId"); ok {
			out["mcpServerIds"] = ids
		}
	}
	return out
}

func validateUpdateInput(input map[string]any) error {
	missing := make([]string, 0)
	for _, field := range requiredUpdateFields {
		value, ok := input[field]
		if !ok || value == nil {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("manifest cannot build updateAgentConfig input; missing %s", strings.Join(missing, ", "))
	}
	return nil
}

func omitRedactedPatchValues(patch map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(patch))
	for key, value := range patch {
		clean, keep, err := omitRedactedPatchValue(value, "/"+escapeJSONPointer(key))
		if err != nil {
			return nil, err
		}
		if keep {
			out[key] = clean
		}
	}
	return out, nil
}

func omitRedactedPatchValue(value any, path string) (any, bool, error) {
	switch typed := value.(type) {
	case string:
		if isRedactedPlaceholder(typed) {
			return nil, false, nil
		}
		return typed, true, nil
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			clean, keep, err := omitRedactedPatchValue(item, path+"/"+escapeJSONPointer(key))
			if err != nil {
				return nil, false, err
			}
			if keep {
				out[key] = clean
			}
		}
		return out, true, nil
	case []any:
		if containsRedactedPlaceholder(typed) {
			return nil, false, fmt.Errorf("redacted manifest value at %s cannot be merged safely inside an array; replace it with a real value or remove the array field", path)
		}
		return deepCopyValue(typed), true, nil
	default:
		return deepCopyValue(typed), true, nil
	}
}

func containsRedactedPlaceholder(value any) bool {
	switch typed := value.(type) {
	case string:
		return isRedactedPlaceholder(typed)
	case map[string]any:
		for _, item := range typed {
			if containsRedactedPlaceholder(item) {
				return true
			}
		}
	case []any:
		for _, item := range typed {
			if containsRedactedPlaceholder(item) {
				return true
			}
		}
	}
	return false
}

func isRedactedPlaceholder(value string) bool {
	trimmed := strings.TrimSpace(value)
	switch strings.ToLower(trimmed) {
	case "redacted", "<redacted>", "[redacted]", "(redacted)", "***redacted***":
		return true
	}
	if len(trimmed) < 4 {
		return false
	}
	for _, r := range trimmed {
		if r != '*' {
			return false
		}
	}
	return true
}

func resolveIDs(flagAppID, flagAgentID string, manifest map[string]any) (string, string, error) {
	appID, agentID := strings.TrimSpace(flagAppID), strings.TrimSpace(flagAgentID)
	if appID == "" {
		appID = firstStringAt(manifest,
			[]string{"metadata", "appId"},
			[]string{"spec", "appId"},
			[]string{"appId"},
		)
	}
	if agentID == "" {
		agentID = firstStringAt(manifest,
			[]string{"metadata", "agentId"},
			[]string{"spec", "agentId"},
			[]string{"agentId"},
			[]string{"sourceAgentId"},
			[]string{"id"},
		)
	}
	return requireIDs(appID, agentID)
}

func requireIDs(appID, agentID string) (string, string, error) {
	appID = strings.TrimSpace(appID)
	agentID = strings.TrimSpace(agentID)
	switch {
	case appID == "" && agentID == "":
		return "", "", fmt.Errorf("--app-id and --agent-id are required")
	case appID == "":
		return "", "", fmt.Errorf("--app-id is required")
	case agentID == "":
		return "", "", fmt.Errorf("--agent-id is required")
	default:
		return appID, agentID, nil
	}
}

func mergeMaps(base, patch map[string]any) map[string]any {
	out := deepCopyMap(base)
	for key, patchValue := range patch {
		if patchMap, ok := asMap(patchValue); ok {
			if baseMap, ok := asMap(out[key]); ok {
				out[key] = mergeMaps(baseMap, patchMap)
				continue
			}
		}
		out[key] = deepCopyValue(patchValue)
	}
	return out
}

func diffValues(before, after map[string]any) []change {
	changes := make([]change, 0)
	collectChanges("", before, after, &changes)
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
	return changes
}

func collectChanges(path string, before, after any, changes *[]change) {
	beforeMap, beforeIsMap := asMap(before)
	afterMap, afterIsMap := asMap(after)
	if beforeIsMap && afterIsMap {
		keys := make([]string, 0, len(beforeMap)+len(afterMap))
		seen := map[string]struct{}{}
		for key := range beforeMap {
			seen[key] = struct{}{}
		}
		for key := range afterMap {
			seen[key] = struct{}{}
		}
		for key := range seen {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectChanges(path+"/"+escapeJSONPointer(key), beforeMap[key], afterMap[key], changes)
		}
		return
	}
	if !reflect.DeepEqual(before, after) {
		if path == "" {
			path = "/"
		}
		*changes = append(*changes, change{Path: path, Before: before, After: after})
	}
}

func printChanges(cmd *cobra.Command, title string, changes []change) {
	out := cmd.OutOrStdout()
	if len(changes) == 0 {
		fmt.Fprintln(out, "No manifest changes.")
		return
	}
	fmt.Fprintf(out, "%s:\n", title)
	for _, change := range changes {
		fmt.Fprintf(out, "  %s\n", change.Path)
	}
}

func wantsJSON(cmd *cobra.Command, opts *commandOptions) bool {
	if opts.json {
		return true
	}
	format, _ := cmd.Root().PersistentFlags().GetString("output")
	return format == "json"
}

func writeJSON(cmd *cobra.Command, value any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func writeYAML(cmd *cobra.Command, value any) error {
	enc := yaml.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(value)
}

func compactJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}

func ensureTrailingNewline(value string) string {
	if strings.HasSuffix(value, "\n") {
		return value
	}
	return value + "\n"
}

func copyIfPresent(out, source map[string]any, key string) {
	if value, ok := source[key]; ok {
		out[key] = deepCopyValue(value)
	}
}

func copyPromptIfPresent(out, source map[string]any, key string) {
	if value, ok := source[key]; ok {
		out[key] = normalizePromptValue(value)
	}
}

func copyRenameIfPresent(out, source map[string]any, from, to string) {
	if value, ok := source[from]; ok {
		out[to] = deepCopyValue(value)
	}
}

func copyPromptRenameIfPresent(out, source map[string]any, from, to string) {
	if value, ok := source[from]; ok {
		out[to] = normalizePromptValue(value)
	}
}

func normalizePromptValue(value any) any {
	text, ok := value.(string)
	if !ok {
		return deepCopyValue(value)
	}
	return strings.TrimSuffix(text, "\n")
}

func objectAt(value map[string]any, path ...string) (map[string]any, bool) {
	raw, ok := valueAt(value, path...)
	if !ok {
		return nil, false
	}
	return asMap(raw)
}

func valueAt(value map[string]any, path ...string) (any, bool) {
	var current any = value
	for _, segment := range path {
		asObject, ok := asMap(current)
		if !ok {
			return nil, false
		}
		current, ok = asObject[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func asMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	default:
		return nil, false
	}
}

func firstStringAt(value map[string]any, paths ...[]string) string {
	for _, path := range paths {
		if raw, ok := valueAt(value, path...); ok {
			if text := strings.TrimSpace(fmt.Sprint(raw)); text != "" {
				return text
			}
		}
	}
	return ""
}

func stringValue(value any, fallback string) string {
	if value == nil {
		return fallback
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return fallback
	}
	return text
}

func idsFromObjectArray(value any, field string) ([]any, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	ids := make([]any, 0, len(items))
	for _, item := range items {
		object, ok := asMap(item)
		if !ok {
			continue
		}
		if id, ok := object[field]; ok {
			if text := strings.TrimSpace(fmt.Sprint(id)); text != "" && text != "<nil>" {
				ids = append(ids, id)
			}
		}
	}
	return ids, true
}

func deepCopyMap(value map[string]any) map[string]any {
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = deepCopyValue(item)
	}
	return out
}

func deepCopyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return deepCopyMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = deepCopyValue(item)
		}
		return out
	default:
		return typed
	}
}

func escapeJSONPointer(value string) string {
	value = strings.ReplaceAll(value, "~", "~0")
	return strings.ReplaceAll(value, "/", "~1")
}
