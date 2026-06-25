package agentmanifest

import (
	"reflect"
	"testing"
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

	changes, finalInput, err := planManifestUpdate(remote, local)
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

	changes, finalInput, err := planManifestUpdate(remote, local)
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
