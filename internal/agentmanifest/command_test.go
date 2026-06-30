package agentmanifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/langgenius/mosoo-connector/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

func TestResolveIDsReadsMetadataAndSpec(t *testing.T) {
	manifest := map[string]any{
		"metadata": map[string]any{
			"appId": "app_123",
		},
		"spec": map[string]any{
			"agentId": "ag_123",
		},
	}

	appID, agentID, err := resolveIDs("", "", manifest)
	if err != nil {
		t.Fatalf("resolveIDs: %v", err)
	}
	if appID != "app_123" || agentID != "ag_123" {
		t.Fatalf("ids = %q, %q", appID, agentID)
	}
}

func TestPlanManifestUpdatePreservesOmittedFields(t *testing.T) {
	remote := map[string]any{
		"spec": map[string]any{
			"agentId": "ag_123",
			"appId":   "app_123",
			"kind":    "cattle",
			"mcpServerIds": []any{
				"mcp_1",
			},
			"model":    "gpt-4.1",
			"name":     "Researcher",
			"prompt":   "old prompt",
			"provider": "openai",
			"providerOptions": map[string]any{
				"temperature": float64(0.7),
				"topP":        float64(0.9),
			},
			"runtimeId": "rt_1",
			"skillIds": []any{
				"skill_1",
			},
			"environment": map[string]any{
				"environmentId": "env_1",
			},
		},
	}
	local := map[string]any{
		"apiVersion": "mosoo.ai/v1",
		"kind":       "AgentManifest",
		"metadata": map[string]any{
			"appId":   "app_123",
			"agentId": "ag_123",
		},
		"spec": map[string]any{
			"prompt": "new prompt",
			"providerOptions": map[string]any{
				"temperature": float64(0.2),
			},
		},
	}

	changes, finalInput, err := planManifestUpdate(remote, local, "", "")
	if err != nil {
		t.Fatalf("planManifestUpdate: %v", err)
	}
	if got := finalInput["prompt"]; got != "new prompt" {
		t.Fatalf("prompt = %v", got)
	}
	providerOptions := finalInput["providerOptions"].(map[string]any)
	if providerOptions["temperature"] != float64(0.2) {
		t.Fatalf("temperature = %v", providerOptions["temperature"])
	}
	if providerOptions["topP"] != float64(0.9) {
		t.Fatalf("topP was not preserved: %v", providerOptions["topP"])
	}
	if got := finalInput["model"]; got != "gpt-4.1" {
		t.Fatalf("model was not preserved: %v", got)
	}
	if err := validateUpdateInput(finalInput); err != nil {
		t.Fatalf("validateUpdateInput: %v", err)
	}
	wantPaths := []string{"/prompt", "/providerOptions/temperature"}
	if got := changePaths(changes); !reflect.DeepEqual(got, wantPaths) {
		t.Fatalf("change paths = %#v, want %#v", got, wantPaths)
	}
}

func TestUpdateInputFromExportedAgentManifest(t *testing.T) {
	manifest := map[string]any{
		"sourceAgentId":   "ag_123",
		"manifestVersion": "1",
		"kind":            "pet",
		"metadata": map[string]any{
			"name":        "Portable Agent",
			"description": "Imported safely",
		},
		"runtime": map[string]any{
			"id":       "openai-runtime",
			"provider": "openai",
			"model":    "gpt-5.4",
			"settings": map[string]any{
				"temperature": float64(0.2),
			},
		},
		"prompts": map[string]any{
			"system": "Help",
		},
		"skills": []any{
			map[string]any{
				"skillId": "skill_1",
			},
		},
		"mcpServers": []any{
			map[string]any{
				"serverId": "mcp_1",
			},
		},
		"environment": map[string]any{
			"environmentId": "env_1",
			"expectedName":  "Production tools",
		},
		"builtInTools": []any{
			map[string]any{"name": "browser", "enabled": true},
		},
	}

	input := updateInputFromManifest(manifest)
	if input["agentId"] != "ag_123" {
		t.Fatalf("agentId = %v", input["agentId"])
	}
	if input["name"] != "Portable Agent" || input["prompt"] != "Help" {
		t.Fatalf("name/prompt = %v / %v", input["name"], input["prompt"])
	}
	if input["runtimeId"] != "openai-runtime" || input["provider"] != "openai" || input["model"] != "gpt-5.4" {
		t.Fatalf("runtime fields = %#v", input)
	}
	if got := input["providerOptions"]; !reflect.DeepEqual(got, map[string]any{"temperature": float64(0.2)}) {
		t.Fatalf("providerOptions = %#v", got)
	}
	if got := input["skillIds"]; !reflect.DeepEqual(got, []any{"skill_1"}) {
		t.Fatalf("skillIds = %#v", got)
	}
	if got := input["mcpServerIds"]; !reflect.DeepEqual(got, []any{"mcp_1"}) {
		t.Fatalf("mcpServerIds = %#v", got)
	}
	if got := input["environment"]; !reflect.DeepEqual(got, map[string]any{"environmentId": "env_1"}) {
		t.Fatalf("environment = %#v", got)
	}
}

func TestPlanManifestUpdateReplacesArrays(t *testing.T) {
	remote := map[string]any{
		"agentId": "ag_123",
		"appId":   "app_123",
		"kind":    "cattle",
		"mcpServerIds": []any{
			"mcp_1",
		},
		"model":           "gpt-4.1",
		"name":            "Researcher",
		"prompt":          "prompt",
		"provider":        "openai",
		"providerOptions": map[string]any{},
		"runtimeId":       "rt_1",
		"skillIds": []any{
			"skill_1",
			"skill_2",
		},
	}
	local := map[string]any{
		"skillIds": []any{},
	}

	changes, finalInput, err := planManifestUpdate(remote, local, "", "")
	if err != nil {
		t.Fatalf("planManifestUpdate: %v", err)
	}
	if got := finalInput["skillIds"]; !reflect.DeepEqual(got, []any{}) {
		t.Fatalf("skillIds = %#v", got)
	}
	if got := changePaths(changes); !reflect.DeepEqual(got, []string{"/skillIds"}) {
		t.Fatalf("change paths = %#v", got)
	}
}

func TestPlanManifestUpdateResolvedIDsWinOverLocalManifestIDs(t *testing.T) {
	remote := map[string]any{
		"agentId":         "ag_target",
		"appId":           "app_target",
		"kind":            "cattle",
		"mcpServerIds":    []any{},
		"model":           "gpt-4.1",
		"name":            "Researcher",
		"prompt":          "prompt",
		"provider":        "openai",
		"providerOptions": map[string]any{},
		"runtimeId":       "rt_1",
		"skillIds":        []any{},
	}
	local := map[string]any{
		"sourceAgentId": "ag_source",
		"appId":         "app_source",
		"name":          "Researcher",
	}

	changes, finalInput, err := planManifestUpdate(remote, local, "app_target", "ag_target")
	if err != nil {
		t.Fatalf("planManifestUpdate: %v", err)
	}
	if finalInput["appId"] != "app_target" || finalInput["agentId"] != "ag_target" {
		t.Fatalf("ids = %v / %v", finalInput["appId"], finalInput["agentId"])
	}
	if got := changePaths(changes); len(got) != 0 {
		t.Fatalf("change paths = %#v, want none", got)
	}
}

func TestPlanManifestUpdatePreservesRemoteValuesForRedactedLocalFields(t *testing.T) {
	remote := map[string]any{
		"agentId": "ag_123",
		"appId":   "app_123",
		"kind":    "cattle",
		"mcpServerIds": []any{
			"mcp_1",
		},
		"model":    "gpt-4.1",
		"name":     "Researcher",
		"prompt":   "prompt",
		"provider": "openai",
		"providerOptions": map[string]any{
			"apiKey":      "live-secret",
			"temperature": float64(0.7),
		},
		"runtimeId": "rt_1",
		"skillIds": []any{
			"skill_1",
		},
	}
	local := map[string]any{
		"providerOptions": map[string]any{
			"apiKey":      "<redacted>",
			"temperature": float64(0.2),
		},
	}

	changes, finalInput, err := planManifestUpdate(remote, local, "", "")
	if err != nil {
		t.Fatalf("planManifestUpdate: %v", err)
	}
	providerOptions := finalInput["providerOptions"].(map[string]any)
	if providerOptions["apiKey"] != "live-secret" {
		t.Fatalf("apiKey = %v", providerOptions["apiKey"])
	}
	if providerOptions["temperature"] != float64(0.2) {
		t.Fatalf("temperature = %v", providerOptions["temperature"])
	}
	if got := changePaths(changes); !reflect.DeepEqual(got, []string{"/providerOptions/temperature"}) {
		t.Fatalf("change paths = %#v", got)
	}
}

func TestPlanManifestUpdateRejectsRedactedValuesInsideArrays(t *testing.T) {
	remote := map[string]any{
		"agentId":         "ag_123",
		"appId":           "app_123",
		"kind":            "cattle",
		"mcpServerIds":    []any{},
		"model":           "gpt-4.1",
		"name":            "Researcher",
		"prompt":          "prompt",
		"provider":        "openai",
		"providerOptions": map[string]any{},
		"runtimeId":       "rt_1",
		"skillIds":        []any{},
	}
	local := map[string]any{
		"builtInTools": []any{
			map[string]any{
				"name":  "browser",
				"token": "<redacted>",
			},
		},
	}

	_, _, err := planManifestUpdate(remote, local, "", "")
	if err == nil {
		t.Fatal("expected redacted array value error")
	}
	if !strings.Contains(err.Error(), "/builtInTools") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseManifestValueRejectsNullJSON(t *testing.T) {
	if _, err := parseManifestValue("null", ""); err == nil {
		t.Fatal("expected null JSON manifest error")
	}
}

func TestParseYAMLMapRejectsInvalidYAML(t *testing.T) {
	if _, err := parseYAMLMap([]byte("name: [")); err == nil {
		t.Fatal("expected invalid YAML error")
	}
}

func TestValidateUpdateInputRejectsNilRequiredField(t *testing.T) {
	input := map[string]any{
		"agentId":         "ag_123",
		"appId":           "app_123",
		"kind":            "cattle",
		"mcpServerIds":    []any{},
		"model":           "gpt-4.1",
		"name":            "Researcher",
		"prompt":          "prompt",
		"provider":        "openai",
		"providerOptions": map[string]any{},
		"runtimeId":       nil,
		"skillIds":        []any{},
	}

	if err := validateUpdateInput(input); err == nil {
		t.Fatal("expected nil runtimeId error")
	}
}

func TestRunDiffRejectsInvalidManifestBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	file := filepath.Join(t.TempDir(), "agent.yaml")
	if err := os.WriteFile(file, []byte("spec:\n  prompt: ["), 0o644); err != nil {
		t.Fatal(err)
	}

	root, _ := newAgentManifestTestRoot(t, srv.URL)
	root.SetArgs([]string{
		"--hostname", srv.URL,
		"agent", "manifest", "diff",
		"--app-id", "app_1",
		"--agent-id", "ag_1",
		"--file", file,
	})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse "+file) {
		t.Fatalf("error = %q, want parse filename", err.Error())
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
}

func TestRunProbeRejectsRemoteMissingManifestData(t *testing.T) {
	srv := newAgentManifestGraphQLServer(t, func(query string, _ map[string]any) map[string]any {
		if !strings.Contains(query, "agentManifest") {
			t.Fatalf("unexpected query: %s", query)
		}
		return map[string]any{
			"data": map[string]any{
				"agentManifest": map[string]any{
					"agentId": "ag_1",
				},
			},
		}
	})
	defer srv.Close()

	root, _ := newAgentManifestTestRoot(t, srv.URL)
	root.SetArgs([]string{
		"--hostname", srv.URL,
		"agent", "manifest", "probe",
		"--app-id", "app_1",
		"--agent-id", "ag_1",
	})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected missing manifest error")
	}
	if !strings.Contains(err.Error(), "remote manifest included neither json nor yaml") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRunApplyDryRunFetchesRemoteButDoesNotUpdate(t *testing.T) {
	var manifestHits, updateHits int
	srv := newAgentManifestGraphQLServer(t, func(query string, _ map[string]any) map[string]any {
		switch {
		case strings.Contains(query, "agentManifest"):
			manifestHits++
			return agentManifestResponse(map[string]any{
				"agentId":         "ag_1",
				"appId":           "app_1",
				"kind":            "cattle",
				"mcpServerIds":    []any{},
				"model":           "gpt-4.1",
				"name":            "Researcher",
				"prompt":          "old prompt",
				"provider":        "openai",
				"providerOptions": map[string]any{},
				"runtimeId":       "rt_1",
				"skillIds":        []any{},
			})
		case strings.Contains(query, "updateAgentConfig"):
			updateHits++
			return map[string]any{"data": map[string]any{"updateAgentConfig": map[string]any{"id": "ag_1"}}}
		default:
			t.Fatalf("unexpected query: %s", query)
		}
		return nil
	})
	defer srv.Close()

	file := writeManifest(t, "spec:\n  prompt: new prompt\n")
	root, out := newAgentManifestTestRoot(t, srv.URL)
	root.SetArgs([]string{
		"--hostname", srv.URL,
		"agent", "manifest", "apply",
		"--app-id", "app_1",
		"--agent-id", "ag_1",
		"--file", file,
		"--dry-run",
		"--json",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if manifestHits != 1 {
		t.Fatalf("manifest hits = %d, want 1", manifestHits)
	}
	if updateHits != 0 {
		t.Fatalf("update hits = %d, want 0", updateHits)
	}
	var got applyResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode output %q: %v", out.String(), err)
	}
	if !got.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if got.UpdateResponse != nil {
		t.Fatalf("UpdateResponse = %#v, want nil", got.UpdateResponse)
	}
	if paths := changePaths(got.Changes); !reflect.DeepEqual(paths, []string{"/prompt"}) {
		t.Fatalf("change paths = %#v", paths)
	}
}

func TestRunApplySkipsUpdateWhenManifestIsCurrent(t *testing.T) {
	var updateHits int
	srv := newAgentManifestGraphQLServer(t, func(query string, _ map[string]any) map[string]any {
		switch {
		case strings.Contains(query, "agentManifest"):
			return agentManifestResponse(map[string]any{
				"agentId":         "ag_1",
				"appId":           "app_1",
				"kind":            "cattle",
				"mcpServerIds":    []any{},
				"model":           "gpt-4.1",
				"name":            "Researcher",
				"prompt":          "same prompt",
				"provider":        "openai",
				"providerOptions": map[string]any{},
				"runtimeId":       "rt_1",
				"skillIds":        []any{},
			})
		case strings.Contains(query, "updateAgentConfig"):
			updateHits++
			return map[string]any{"data": map[string]any{"updateAgentConfig": map[string]any{"id": "ag_1"}}}
		default:
			t.Fatalf("unexpected query: %s", query)
		}
		return nil
	})
	defer srv.Close()

	file := writeManifest(t, "spec:\n  prompt: same prompt\n")
	root, out := newAgentManifestTestRoot(t, srv.URL)
	root.SetArgs([]string{
		"--hostname", srv.URL,
		"agent", "manifest", "apply",
		"--app-id", "app_1",
		"--agent-id", "ag_1",
		"--file", file,
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if updateHits != 0 {
		t.Fatalf("update hits = %d, want 0", updateHits)
	}
	if got := out.String(); !strings.Contains(got, "No manifest changes.") || !strings.Contains(got, "no remote changes written") {
		t.Fatalf("output = %q", got)
	}
}

func TestRunApplySendsMergedUpdateInput(t *testing.T) {
	var updateInput map[string]any
	srv := newAgentManifestGraphQLServer(t, func(query string, variables map[string]any) map[string]any {
		switch {
		case strings.Contains(query, "agentManifest"):
			return agentManifestResponse(map[string]any{
				"agentId":      "ag_1",
				"appId":        "app_1",
				"kind":         "cattle",
				"mcpServerIds": []any{"mcp_1"},
				"model":        "gpt-4.1",
				"name":         "Researcher",
				"prompt":       "old prompt",
				"provider":     "openai",
				"providerOptions": map[string]any{
					"temperature": float64(0.7),
					"topP":        float64(0.9),
				},
				"runtimeId": "rt_1",
				"skillIds":  []any{"skill_1"},
			})
		case strings.Contains(query, "updateAgentConfig"):
			input, ok := variables["input"].(map[string]any)
			if !ok {
				t.Fatalf("variables.input = %#v", variables["input"])
			}
			updateInput = input
			return map[string]any{"data": map[string]any{"updateAgentConfig": map[string]any{"id": "ag_1"}}}
		default:
			t.Fatalf("unexpected query: %s", query)
		}
		return nil
	})
	defer srv.Close()

	file := writeManifest(t, "spec:\n  prompt: new prompt\n  providerOptions:\n    temperature: 0.2\n")
	root, out := newAgentManifestTestRoot(t, srv.URL)
	root.SetArgs([]string{
		"--hostname", srv.URL,
		"agent", "manifest", "apply",
		"--app-id", "app_1",
		"--agent-id", "ag_1",
		"--file", file,
		"--json",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if updateInput == nil {
		t.Fatal("updateAgentConfig was not called")
	}
	if updateInput["prompt"] != "new prompt" {
		t.Fatalf("prompt = %v", updateInput["prompt"])
	}
	providerOptions := updateInput["providerOptions"].(map[string]any)
	if providerOptions["temperature"] != float64(0.2) {
		t.Fatalf("temperature = %v", providerOptions["temperature"])
	}
	if providerOptions["topP"] != float64(0.9) {
		t.Fatalf("topP = %v", providerOptions["topP"])
	}
	if got := updateInput["skillIds"]; !reflect.DeepEqual(got, []any{"skill_1"}) {
		t.Fatalf("skillIds = %#v", got)
	}
	var got applyResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode output %q: %v", out.String(), err)
	}
	if got.UpdateResponse == nil {
		t.Fatal("UpdateResponse = nil, want response")
	}
	if paths := changePaths(got.Changes); !reflect.DeepEqual(paths, []string{"/prompt", "/providerOptions/temperature"}) {
		t.Fatalf("change paths = %#v", paths)
	}
}

func TestPlanManifestUpdateRoundTripsExportedManifest(t *testing.T) {
	remote := map[string]any{
		"sourceAgentId":   "ag_123",
		"manifestVersion": "mosoo.agent.manifest.v1",
		"kind":            "pet",
		"metadata": map[string]any{
			"name":        "Portable Agent",
			"description": nil,
		},
		"runtime": map[string]any{
			"id":              "openai-runtime",
			"provider":        "openai",
			"model":           "gpt-5.4-mini",
			"providerOptions": map[string]any{},
		},
		"prompts": map[string]any{
			"system": "Line one\nLine two",
		},
		"skills":     []any{},
		"mcpServers": []any{},
		"environment": map[string]any{
			"environmentId": "env_1",
			"expectedName":  "System Default",
			"setupScript":   "",
			"envVars":       map[string]any{},
		},
	}
	local := deepCopyMap(remote)
	local["prompts"].(map[string]any)["system"] = "Line one\nLine two\n"

	changes, finalInput, err := planManifestUpdate(remote, local, "app_123", "ag_123")
	if err != nil {
		t.Fatalf("planManifestUpdate: %v", err)
	}
	if got := changePaths(changes); len(got) != 0 {
		t.Fatalf("change paths = %#v, want none", got)
	}
	if finalInput["appId"] != "app_123" || finalInput["agentId"] != "ag_123" {
		t.Fatalf("ids = %v / %v", finalInput["appId"], finalInput["agentId"])
	}
	if got := finalInput["skillIds"]; !reflect.DeepEqual(got, []any{}) {
		t.Fatalf("skillIds = %#v", got)
	}
	if got := finalInput["mcpServerIds"]; !reflect.DeepEqual(got, []any{}) {
		t.Fatalf("mcpServerIds = %#v", got)
	}
	if err := validateUpdateInput(finalInput); err != nil {
		t.Fatalf("validateUpdateInput: %v", err)
	}
}

func TestPatchSourceRejectsUnknownSpecField(t *testing.T) {
	_, err := patchSource(map[string]any{
		"spec": map[string]any{
			"promtp": "typo",
		},
	})
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func changePaths(changes []change) []string {
	out := make([]string, 0, len(changes))
	for _, change := range changes {
		out = append(out, change.Path)
	}
	return out
}

func newAgentManifestTestRoot(t *testing.T, baseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	latheconfig.Bind(&latheconfig.Manifest{CLI: latheconfig.CLIInfo{
		Name:         "mosoo",
		ConfigDir:    "mosoo",
		ConfigDirEnv: "MOSOO_CONFIG_DIR",
		HostEnv:      "MOSOO_HOST",
	}})
	t.Setenv("MOSOO_CONFIG_DIR", filepath.Join(t.TempDir(), "config"))
	t.Setenv("MOSOO_HOST", "")
	t.Setenv(target.TargetEnv, "")
	t.Setenv(target.BaseURLEnv, "")
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		t.Fatal(err)
	}
	hosts.Set(baseURL+"/api", latheconfig.HostEntry{AuthType: "bearer", OAuthToken: "test-token"})
	if err := hosts.Save(); err != nil {
		t.Fatal(err)
	}

	root := &cobra.Command{Use: "mosoo"}
	root.PersistentFlags().String("hostname", "", "")
	root.PersistentFlags().StringP("output", "o", "raw", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.PersistentFlags().Bool("insecure", false, "")
	root.AddCommand(NewCommand())
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	return root, &out
}

func newAgentManifestGraphQLServer(t *testing.T, handler func(query string, variables map[string]any) map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graphql" {
			t.Fatalf("path = %q, want /api/graphql", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want Bearer test-token", got)
		}
		var request struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Variables == nil {
			request.Variables = map[string]any{}
		}
		response := handler(request.Query, request.Variables)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
}

func agentManifestResponse(manifest map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"agentManifest": map[string]any{
				"agentId": fmt.Sprint(manifest["agentId"]),
				"json":    manifest,
			},
		},
	}
}

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "agent.yaml")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}
