package agentapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/mosoo-cli-go/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

func TestProvisionDryRunReportsShortOperatorWorkflow(t *testing.T) {
	root := newAgentAppTestRoot(t, "")
	specPath := writeTestSpec(t, `
app:
  name: Research Desk
  organizationId: org_1
agents:
  - key: researcher
    name: Researcher
    kind: pet
    runtimeId: openai
    provider: openai
    model: gpt-4.1
    prompt: Research A-shares.
    skillIds: []
`)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"agent-app", "provision", "--file", specPath, "--dry-run", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var got provisionResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", out.String(), err)
	}
	if got.DryRun != true {
		t.Fatal("DryRun = false, want true")
	}
	if got.App.Name != "Research Desk" {
		t.Fatalf("App.Name = %q", got.App.Name)
	}
	if len(got.OperatorSteps) == 0 || len(got.OperatorSteps) > 8 {
		t.Fatalf("operator steps count = %d, want 1..8", len(got.OperatorSteps))
	}
	if strings.Contains(out.String(), "test-token") {
		t.Fatal("output leaked auth token")
	}
}

func TestProvisionCreatesAppAgentsAndPublishes(t *testing.T) {
	var operations []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graphql" {
			t.Fatalf("path = %s, want /api/graphql", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("invalid request JSON %q: %v", string(raw), err)
		}
		query, _ := body["query"].(string)
		switch {
		case strings.Contains(query, "appList"):
			operations = append(operations, "appList")
			writeJSONResponse(t, w, map[string]any{"data": map[string]any{"appList": []any{}}})
		case strings.Contains(query, "createApp"):
			operations = append(operations, "createApp")
			input := graphQLInput(t, body)
			if input["name"] != "Research Desk" || input["organizationId"] != "org_1" {
				t.Fatalf("createApp input = %#v", input)
			}
			writeJSONResponse(t, w, map[string]any{"data": map[string]any{"createApp": map[string]any{"id": "app_1", "name": "Research Desk"}}})
		case strings.Contains(query, "createAgent"):
			operations = append(operations, "createAgent")
			input := graphQLInput(t, body)
			if input["appId"] != "app_1" || input["name"] == "" {
				t.Fatalf("createAgent input = %#v", input)
			}
			writeJSONResponse(t, w, map[string]any{"data": map[string]any{"createAgent": map[string]any{"id": "agent_1", "name": input["name"]}}})
		case strings.Contains(query, "publishAgent"):
			operations = append(operations, "publishAgent")
			input := graphQLInput(t, body)
			if input["appId"] != "app_1" || input["agentId"] != "agent_1" {
				t.Fatalf("publishAgent input = %#v", input)
			}
			writeJSONResponse(t, w, map[string]any{"data": map[string]any{"publishAgent": map[string]any{"id": "version_1", "agentId": "agent_1", "isLive": true}}})
		default:
			t.Fatalf("unexpected query: %s", query)
		}
	}))
	defer srv.Close()

	root := newAgentAppTestRoot(t, srv.URL)
	specPath := writeTestSpec(t, `
app:
  name: Research Desk
  organizationId: org_1
agents:
  - key: researcher
    name: Researcher
    kind: pet
    runtimeId: openai
    provider: openai
    model: gpt-4.1
    prompt: Research A-shares.
    skillIds: []
`)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--target", "custom", "--base-url", srv.URL, "agent-app", "provision", "--file", specPath, "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	wantOps := []string{"appList", "createApp", "createAgent", "publishAgent"}
	if strings.Join(operations, ",") != strings.Join(wantOps, ",") {
		t.Fatalf("operations = %#v, want %#v", operations, wantOps)
	}
	var got provisionResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", out.String(), err)
	}
	if got.App.ID != "app_1" || got.Agents[0].ID != "agent_1" || !got.Agents[0].Published {
		t.Fatalf("result = %#v", got)
	}
	if strings.Contains(out.String(), "test-token") {
		t.Fatal("output leaked auth token")
	}
}

func TestProvisionReusesExplicitAppAndAgent(t *testing.T) {
	var operations []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("invalid request JSON %q: %v", string(raw), err)
		}
		query, _ := body["query"].(string)
		operations = append(operations, operationName(query))
		if strings.Contains(query, "publishAgent") {
			writeJSONResponse(t, w, map[string]any{"data": map[string]any{"publishAgent": map[string]any{"id": "version_1", "agentId": "agent_existing", "isLive": true}}})
			return
		}
		t.Fatalf("unexpected query for reuse path: %s", query)
	}))
	defer srv.Close()

	root := newAgentAppTestRoot(t, srv.URL)
	specPath := writeTestSpec(t, `
app:
  id: app_existing
  name: Existing App
agents:
  - key: researcher
    id: agent_existing
    name: Researcher
    kind: pet
    runtimeId: openai
    provider: openai
    model: gpt-4.1
    prompt: Research A-shares.
`)
	root.SetArgs([]string{"--target", "custom", "--base-url", srv.URL, "agent-app", "provision", "--file", specPath, "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Join(operations, ",") != "publishAgent" {
		t.Fatalf("operations = %#v, want publishAgent only", operations)
	}
}

func TestProvisionMissingCredentialsFailsBeforeMutation(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	root := newAgentAppTestRootWithoutAuth(t)
	specPath := writeTestSpec(t, `
app:
  name: Research Desk
  organizationId: org_1
agents:
  - key: researcher
    name: Researcher
    kind: pet
    runtimeId: openai
    provider: openai
    model: gpt-4.1
    prompt: Research A-shares.
`)
	root.SetArgs([]string{"--target", "custom", "--base-url", srv.URL, "agent-app", "provision", "--file", specPath})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected missing credential error")
	}
	if !strings.Contains(err.Error(), "not authenticated to console host") {
		t.Fatalf("error = %q, want credential diagnostic", err.Error())
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
}

func TestProvisionPartialPublishFailureReportsCreatedResource(t *testing.T) {
	var operations []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("invalid request JSON %q: %v", string(raw), err)
		}
		query, _ := body["query"].(string)
		operations = append(operations, operationName(query))
		switch {
		case strings.Contains(query, "createAgent"):
			writeJSONResponse(t, w, map[string]any{"data": map[string]any{"createAgent": map[string]any{"id": "agent_1", "name": "Researcher"}}})
		case strings.Contains(query, "publishAgent"):
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":[{"message":"publish failed","extensions":{"code":"PUBLISH_FAILED"}}],"token":"test-token"}`))
		default:
			t.Fatalf("unexpected query: %s", query)
		}
	}))
	defer srv.Close()

	root := newAgentAppTestRoot(t, srv.URL)
	specPath := writeTestSpec(t, `
app:
  id: app_1
agents:
  - key: researcher
    name: Researcher
    kind: pet
    runtimeId: openai
    provider: openai
    model: gpt-4.1
    prompt: Research A-shares.
`)
	root.SetArgs([]string{"--target", "custom", "--base-url", srv.URL, "agent-app", "provision", "--file", specPath})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected publish failure")
	}
	got := err.Error()
	for _, want := range []string{"partial provisioning failure", "agent_1", "PUBLISH_FAILED"} {
		if !strings.Contains(got, want) {
			t.Fatalf("error = %q, want to contain %q", got, want)
		}
	}
	if strings.Contains(got, "test-token") {
		t.Fatal("error leaked auth token")
	}
	if strings.Join(operations, ",") != "createAgent,publishAgent" {
		t.Fatalf("operations = %#v", operations)
	}
}

func TestProvisionWriteEnvIsGatedByDependency(t *testing.T) {
	root := newAgentAppTestRoot(t, "")
	specPath := writeTestSpec(t, `
app:
  id: app_1
agents:
  - key: researcher
    id: agent_1
    name: Researcher
`)
	root.SetArgs([]string{"agent-app", "provision", "--file", specPath, "--write-env", ".dev.vars"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected dependency gate")
	}
	if !strings.Contains(err.Error(), "GitHub issue #15") {
		t.Fatalf("error = %q, want #15 diagnostic", err.Error())
	}
}

func TestSmokeTestCreateFailureReportsPublicAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents/agent_1/threads" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"code":"readiness_blocked","message":"agent is not ready"},"token":"test-token"}`))
	}))
	defer srv.Close()

	root := newAgentAppTestRoot(t, srv.URL)
	bodyPath := filepath.Join(t.TempDir(), "smoke.json")
	if err := os.WriteFile(bodyPath, []byte(`{"input":{"content":"hello"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	root.SetArgs([]string{"--target", "custom", "--base-url", srv.URL, "agent-app", "smoke-test", "--agent-id", "agent_1", "--file", bodyPath})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected public API failure")
	}
	got := err.Error()
	if !strings.Contains(got, "create public thread") || !strings.Contains(got, "readiness_blocked") {
		t.Fatalf("error = %q, want public API details", got)
	}
	if strings.Contains(got, "test-token") {
		t.Fatal("error leaked auth token")
	}
}

func TestSmokeTestWaitIsGatedByDependency(t *testing.T) {
	root := newAgentAppTestRoot(t, "")
	bodyPath := filepath.Join(t.TempDir(), "smoke.json")
	if err := os.WriteFile(bodyPath, []byte(`{"input":{"content":"hello"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	root.SetArgs([]string{"agent-app", "smoke-test", "--agent-id", "agent_1", "--file", bodyPath, "--wait"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected dependency gate")
	}
	if !strings.Contains(err.Error(), "GitHub issue #17") {
		t.Fatalf("error = %q, want #17 diagnostic", err.Error())
	}
}

func newAgentAppTestRoot(t *testing.T, baseURL string) *cobra.Command {
	t.Helper()
	root := newAgentAppTestRootWithoutAuth(t)
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8787"
	}
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		t.Fatal(err)
	}
	hosts.Set(baseURL+"/api", latheconfig.HostEntry{AuthType: "bearer", OAuthToken: "test-token"})
	hosts.Set(baseURL+"/api/v1", latheconfig.HostEntry{AuthType: "bearer", OAuthToken: "test-token"})
	if err := hosts.Save(); err != nil {
		t.Fatal(err)
	}
	return root
}

func newAgentAppTestRootWithoutAuth(t *testing.T) *cobra.Command {
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
	root := &cobra.Command{Use: "mosoo", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("hostname", "", "")
	root.PersistentFlags().String("target", "", "")
	root.PersistentFlags().String("base-url", "", "")
	root.PersistentFlags().StringP("output", "o", "raw", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.PersistentFlags().Bool("insecure", false, "")
	root.AddCommand(NewCommand())
	return root
}

func writeTestSpec(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mosoo-agent-app.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func graphQLInput(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	variables, ok := body["variables"].(map[string]any)
	if !ok {
		t.Fatalf("variables = %#v", body["variables"])
	}
	input, ok := variables["input"].(map[string]any)
	if !ok {
		t.Fatalf("input = %#v", variables["input"])
	}
	return input
}

func operationName(query string) string {
	for _, name := range []string{"appList", "createApp", "createAgent", "publishAgent"} {
		if strings.Contains(query, name) {
			return name
		}
	}
	return "unknown"
}
