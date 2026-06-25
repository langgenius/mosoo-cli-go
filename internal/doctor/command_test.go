package doctor

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/mosoo-cli-go/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
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
