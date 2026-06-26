package target

import (
	"os"
	"path/filepath"
	"testing"

	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

func bindTestManifest(t *testing.T, configDir string) {
	t.Helper()
	latheconfig.Bind(&latheconfig.Manifest{CLI: latheconfig.CLIInfo{
		Name:         "mosoo",
		ConfigDir:    "mosoo",
		ConfigDirEnv: "MOSOO_CONFIG_DIR",
		HostEnv:      "MOSOO_HOST",
	}})
	t.Setenv("MOSOO_CONFIG_DIR", configDir)
	t.Setenv("MOSOO_HOST", "")
	t.Setenv(TargetEnv, "")
	t.Setenv(BaseURLEnv, "")
}

func TestResolveDefaultsToLocal(t *testing.T) {
	dir := t.TempDir()
	bindTestManifest(t, filepath.Join(t.TempDir(), "config"))

	resolved, err := Resolve(dir)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Target != LocalTarget {
		t.Fatalf("Target = %q, want %q", resolved.Target, LocalTarget)
	}
	if resolved.Source != SourceDefaultLocal {
		t.Fatalf("Source = %q, want %q", resolved.Source, SourceDefaultLocal)
	}
	if resolved.Hosts[SurfaceConsole] != DefaultLocalBaseURL+"/api" {
		t.Fatalf("console host = %q", resolved.Hosts[SurfaceConsole])
	}
}

func TestResolveUsesProjectConfig(t *testing.T) {
	dir := t.TempDir()
	bindTestManifest(t, filepath.Join(t.TempDir(), "config"))
	if err := os.MkdirAll(filepath.Join(dir, ".mosoo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".mosoo", "config.json"), []byte(`{"target":"custom","baseUrl":"https://example.com/mosoo"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	resolved, err := Resolve(nested)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Target != CustomTarget {
		t.Fatalf("Target = %q, want %q", resolved.Target, CustomTarget)
	}
	if resolved.Source != SourceProjectConfig {
		t.Fatalf("Source = %q, want %q", resolved.Source, SourceProjectConfig)
	}
	if got, want := resolved.Hosts[SurfacePublicThreadAPI], "https://example.com/mosoo/api/v1"; got != want {
		t.Fatalf("public-thread-api host = %q, want %q", got, want)
	}
}

func TestResolveDetectsMosooSourceRoot(t *testing.T) {
	dir := t.TempDir()
	bindTestManifest(t, filepath.Join(t.TempDir(), "config"))
	for _, name := range []string{
		"package.json",
		"justfile",
		filepath.Join("apps", "api", "wrangler.toml"),
	} {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nested := filepath.Join(dir, "apps", "web")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	resolved, err := Resolve(nested)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Source != SourceCwdMosooRepo {
		t.Fatalf("Source = %q, want %q", resolved.Source, SourceCwdMosooRepo)
	}
	if resolved.ProjectRoot != dir {
		t.Fatalf("ProjectRoot = %q, want %q", resolved.ProjectRoot, dir)
	}
}

func TestStateFromResolutionStructuresMachineFields(t *testing.T) {
	resolved := Resolution{
		Target:           CustomTarget,
		Source:           SourceTargetFlag,
		BaseURL:          "https://example.com",
		Hosts:            HostsForBaseURL("https://example.com"),
		ConfigPath:       "/tmp/mosoo/config.json",
		ProjectRoot:      "/tmp/mosoo",
		ExplicitHostname: "example.com/api",
	}

	state := StateFromResolution(resolved)

	if state.Name != CustomTarget {
		t.Fatalf("Name = %q, want %q", state.Name, CustomTarget)
	}
	if state.Source != SourceTargetFlag {
		t.Fatalf("Source = %q, want %q", state.Source, SourceTargetFlag)
	}
	if state.BaseURL != "https://example.com" {
		t.Fatalf("BaseURL = %q", state.BaseURL)
	}
	if state.Hosts[SurfaceConsoleREST] != "https://example.com/api" {
		t.Fatalf("console-rest host = %q", state.Hosts[SurfaceConsoleREST])
	}
	if state.ConfigPath != "/tmp/mosoo/config.json" {
		t.Fatalf("ConfigPath = %q", state.ConfigPath)
	}
	if state.ProjectRoot != "/tmp/mosoo" {
		t.Fatalf("ProjectRoot = %q", state.ProjectRoot)
	}
	if state.ExplicitHostname != "example.com/api" {
		t.Fatalf("ExplicitHostname = %q", state.ExplicitHostname)
	}
	if state.Local {
		t.Fatal("Local = true, want false")
	}
}

func TestInstallSetsHostnameForGeneratedSurface(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")
	bindTestManifest(t, configDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"target":"cloud"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := &cobra.Command{Use: "mosoo"}
	root.PersistentFlags().String("hostname", "", "")
	console := &cobra.Command{Use: SurfaceConsole}
	viewer := &cobra.Command{
		Use: "viewer",
		RunE: func(cmd *cobra.Command, _ []string) error {
			got, _ := cmd.Root().PersistentFlags().GetString("hostname")
			if want := DefaultCloudBaseURL + "/api"; got != want {
				t.Fatalf("hostname = %q, want %q", got, want)
			}
			return nil
		},
	}
	console.AddCommand(viewer)
	root.AddCommand(console)
	Install(root)

	root.SetArgs([]string{SurfaceConsole, "viewer"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}
