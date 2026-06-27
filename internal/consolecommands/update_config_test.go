package consolecommands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
	"github.com/spf13/cobra"
)

func TestUpdateConfigSendsProviderOptionsAsJSONObject(t *testing.T) {
	var rawBody []byte
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		rawBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"updateAgentConfig":{"id":"agent_1"}}}`))
	}))
	defer srv.Close()

	root := newTestRoot(t, srv.URL)
	root.SetArgs(append([]string{"--hostname", srv.URL, "console", "agents", "update-config"}, validUpdateConfigArgs(`{"temperature":0.2,"flags":{"stream":true}}`)...))
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want Bearer token", gotAuth)
	}
	var got map[string]any
	if err := json.Unmarshal(rawBody, &got); err != nil {
		t.Fatalf("invalid request JSON %q: %v", string(rawBody), err)
	}
	variables, _ := got["variables"].(map[string]any)
	input, _ := variables["input"].(map[string]any)
	providerOptions, ok := input["providerOptions"].(map[string]any)
	if !ok {
		t.Fatalf("providerOptions = %#v (%T), want JSON object", input["providerOptions"], input["providerOptions"])
	}
	if providerOptions["temperature"] != 0.2 {
		t.Fatalf("providerOptions.temperature = %#v", providerOptions["temperature"])
	}
	if _, ok := input["providerOptions"].(string); ok {
		t.Fatal("providerOptions was sent as a string")
	}
}

func TestUpdateConfigRejectsInvalidProviderOptionsLocally(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "invalid-json", raw: `{`, want: "invalid --input-provider-options JSON object"},
		{name: "array", raw: `[]`, want: "--input-provider-options must be a JSON object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			root := newTestRoot(t, srv.URL)
			root.SetArgs(append([]string{"--hostname", srv.URL, "console", "agents", "update-config"}, validUpdateConfigArgs(tc.raw)...))
			err := root.Execute()
			if err == nil {
				t.Fatal("expected local validation error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tc.want)
			}
			if hits != 0 {
				t.Fatalf("server hits = %d, want 0", hits)
			}
		})
	}
}

func TestInstallReplacesExistingGeneratedUpdateConfig(t *testing.T) {
	root := &cobra.Command{Use: "mosoo"}
	console := &cobra.Command{Use: "console"}
	agents := &cobra.Command{Use: "agents"}
	generated := &cobra.Command{Use: "update-config", Short: "generated"}
	agents.AddCommand(generated)
	console.AddCommand(agents)
	root.AddCommand(console)

	if err := Install(root); err != nil {
		t.Fatal(err)
	}
	got := findChild(agents, "update-config")
	if got == nil {
		t.Fatal("update-config was not mounted")
	}
	if got == generated {
		t.Fatal("generated update-config command was not replaced")
	}
	if got.Short != "Update an agent config" {
		t.Fatalf("update-config short = %q", got.Short)
	}
}

func TestInstallAttachesHiddenCatalogEntry(t *testing.T) {
	root := &cobra.Command{Use: "mosoo"}
	console := &cobra.Command{Use: "console"}
	agents := &cobra.Command{Use: "agents"}
	agents.AddCommand(&cobra.Command{Use: "update-config", Short: "generated"})
	console.AddCommand(agents)
	root.AddCommand(console)

	if err := Install(root); err != nil {
		t.Fatal(err)
	}
	entry, ok := latheruntime.FindCatalogCommand(root, []string{"console", "agents", "update-config"}, latheruntime.CatalogOptions{IncludeHidden: true})
	if !ok {
		t.Fatal("catalog does not include hidden console agents update-config")
	}
	if !entry.Hidden {
		t.Fatal("update-config catalog entry should be hidden")
	}
	if entry.HTTP.PathTemplate != "/graphql" {
		t.Fatalf("HTTP = %+v", entry.HTTP)
	}
}

func newTestRoot(t *testing.T, host string) *cobra.Command {
	t.Helper()
	latheconfig.Bind(&latheconfig.Manifest{CLI: latheconfig.CLIInfo{
		Name:         "mosoo",
		ConfigDir:    "mosoo",
		ConfigDirEnv: "MOSOO_CONFIG_DIR",
		HostEnv:      "MOSOO_HOST",
	}})
	t.Setenv("MOSOO_CONFIG_DIR", filepath.Join(t.TempDir(), "config"))
	t.Setenv("MOSOO_HOST", "")
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		t.Fatal(err)
	}
	hosts.Set(host, latheconfig.HostEntry{AuthType: "bearer", OAuthToken: "test-token"})
	if err := hosts.Save(); err != nil {
		t.Fatal(err)
	}

	root := &cobra.Command{Use: "mosoo"}
	root.PersistentFlags().String("hostname", "", "")
	root.PersistentFlags().StringP("output", "o", "raw", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.PersistentFlags().Bool("insecure", false, "")
	console := &cobra.Command{Use: "console"}
	agents := &cobra.Command{Use: "agents"}
	console.AddCommand(agents)
	root.AddCommand(console)
	if err := Install(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func validUpdateConfigArgs(providerOptions string) []string {
	return []string{
		"--input-agent-id", "agent_1",
		"--input-app-id", "app_1",
		"--input-kind", "pet",
		"--input-mcp-server-ids", "mcp_1",
		"--input-model", "gpt-4.1",
		"--input-name", "Agent",
		"--input-prompt", "Hello",
		"--input-provider", "openai",
		"--input-provider-options", providerOptions,
		"--input-runtime-id", "runtime_1",
		"--input-skill-ids", "skill_1",
	}
}
