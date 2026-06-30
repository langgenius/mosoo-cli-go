package doctor

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/mosoo-connector/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/lathe-cli/lathe/pkg/lathe"
)

func bindTestManifest(t *testing.T) {
	t.Helper()
	latheconfig.Bind(&latheconfig.Manifest{CLI: latheconfig.CLIInfo{
		Name:         "mosoo",
		ConfigDir:    "mosoo",
		ConfigDirEnv: "MOSOO_CONFIG_DIR",
		HostEnv:      "MOSOO_HOST",
	}})
	t.Setenv("MOSOO_CONFIG_DIR", filepath.Join(t.TempDir(), "config"))
	t.Setenv("MOSOO_HOST", "")
}

func TestCheckAuthSkipsLocalTarget(t *testing.T) {
	bindTestManifest(t)

	authenticated, authRequired, check := checkAuth(target.Resolution{
		Target:  target.LocalTarget,
		BaseURL: target.DefaultLocalBaseURL,
		Hosts:   target.HostsForBaseURL(target.DefaultLocalBaseURL),
	})

	if authenticated {
		t.Fatal("authenticated = true, want false")
	}
	if authRequired {
		t.Fatal("authRequired = true, want false")
	}
	if !check.OK {
		t.Fatalf("check.OK = false, message = %q", check.Message)
	}
	if !strings.Contains(check.Message, "not required") {
		t.Fatalf("check.Message = %q, want local auth exemption", check.Message)
	}
}

func TestCheckAuthRequiresCloudCredentials(t *testing.T) {
	bindTestManifest(t)

	authenticated, authRequired, check := checkAuth(target.Resolution{
		Target:  target.CloudTarget,
		BaseURL: target.DefaultCloudBaseURL,
		Hosts:   target.HostsForBaseURL(target.DefaultCloudBaseURL),
	})

	if authenticated {
		t.Fatal("authenticated = true, want false")
	}
	if !authRequired {
		t.Fatal("authRequired = false, want true")
	}
	if check.OK {
		t.Fatal("check.OK = true, want false")
	}
	if !strings.Contains(check.Message, "not authenticated") {
		t.Fatalf("check.Message = %q, want missing auth message", check.Message)
	}
}

func TestReportJSONHasStructuredReadinessSections(t *testing.T) {
	oldVersion := lathe.Version
	oldCommit := lathe.Commit
	oldDate := lathe.Date
	lathe.Version = "v1.2.3"
	lathe.Commit = "abcdef123456"
	lathe.Date = "2026-06-25T10:32:19Z"
	t.Cleanup(func() {
		lathe.Version = oldVersion
		lathe.Commit = oldCommit
		lathe.Date = oldDate
	})

	auth := AuthState{
		Required:        true,
		Authenticated:   false,
		CredentialHosts: []string{},
		MissingHosts: []string{
			"https://try.mosoo.ai/api",
			"https://try.mosoo.ai/api/v1",
		},
	}
	report := NewReport(target.Resolution{
		Target:  target.CloudTarget,
		Source:  target.SourceTargetFlag,
		BaseURL: target.DefaultCloudBaseURL,
		Hosts:   target.HostsForBaseURL(target.DefaultCloudBaseURL),
	}, Check{
		Name:    "api",
		OK:      true,
		Code:    "api_reachable",
		Message: "GET https://try.mosoo.ai/api/access-tokens returned 401 Unauthorized",
	}, auth)

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	if got["schemaVersion"] != float64(1) {
		t.Fatalf("schemaVersion = %v, want 1", got["schemaVersion"])
	}
	if got["ready"] != false {
		t.Fatalf("ready = %v, want false", got["ready"])
	}

	targetState := got["target"].(map[string]any)
	if targetState["name"] != target.CloudTarget {
		t.Fatalf("target.name = %v", targetState["name"])
	}
	if targetState["baseUrl"] != target.DefaultCloudBaseURL {
		t.Fatalf("target.baseUrl = %v", targetState["baseUrl"])
	}

	authState := got["auth"].(map[string]any)
	if authState["required"] != true {
		t.Fatalf("auth.required = %v", authState["required"])
	}
	if authState["authenticated"] != false {
		t.Fatalf("auth.authenticated = %v", authState["authenticated"])
	}
	if len(authState["missingHosts"].([]any)) != 2 {
		t.Fatalf("auth.missingHosts = %v", authState["missingHosts"])
	}

	installState := got["install"].(map[string]any)
	if installState["version"] != "v1.2.3" {
		t.Fatalf("install.version = %v", installState["version"])
	}
	if installState["complete"] != true {
		t.Fatalf("install.complete = %v", installState["complete"])
	}

	failures := got["failures"].([]any)
	if len(failures) != 1 {
		t.Fatalf("failures len = %d, want 1: %v", len(failures), failures)
	}
	failure := failures[0].(map[string]any)
	if failure["code"] != "auth_missing_credentials" {
		t.Fatalf("failure.code = %v", failure["code"])
	}
	if failure["action"] == "" {
		t.Fatal("failure.action is empty")
	}
}
