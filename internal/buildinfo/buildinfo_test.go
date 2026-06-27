package buildinfo

import (
	"runtime/debug"
	"testing"

	"github.com/lathe-cli/lathe/pkg/lathe"
)

func setLatheBuildInfo(t *testing.T, version, commit, date string) {
	t.Helper()
	oldVersion := lathe.Version
	oldCommit := lathe.Commit
	oldDate := lathe.Date
	lathe.Version = version
	lathe.Commit = commit
	lathe.Date = date
	t.Cleanup(func() {
		lathe.Version = oldVersion
		lathe.Commit = oldCommit
		lathe.Date = oldDate
	})
}

func setGoBuildInfo(t *testing.T, version, commit, date string) {
	t.Helper()
	oldReadBuildInfo := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		settings := []debug.BuildSetting{}
		if commit != "" {
			settings = append(settings, debug.BuildSetting{Key: "vcs.revision", Value: commit})
		}
		if date != "" {
			settings = append(settings, debug.BuildSetting{Key: "vcs.time", Value: date})
		}
		return &debug.BuildInfo{
			Main:     debug.Module{Version: version},
			Settings: settings,
		}, true
	}
	t.Cleanup(func() {
		readBuildInfo = oldReadBuildInfo
	})
}

func TestCurrentUsesDeterministicDefaults(t *testing.T) {
	setLatheBuildInfo(t, "dev", "none", "unknown")
	setGoBuildInfo(t, "", "", "")

	info := Current()

	if info.Version != "dev" {
		t.Fatalf("Version = %q, want dev", info.Version)
	}
	if info.Commit != "none" {
		t.Fatalf("Commit = %q, want none", info.Commit)
	}
	if info.Date != "unknown" {
		t.Fatalf("Date = %q, want unknown", info.Date)
	}
	if info.Complete {
		t.Fatal("Complete = true, want false for default metadata")
	}
}

func TestCurrentReportsCompleteInjectedMetadata(t *testing.T) {
	setLatheBuildInfo(t, "v1.2.3", "abcdef123456", "2026-06-25T10:32:19Z")
	setGoBuildInfo(t, "v9.9.9", "unused", "2026-06-26T10:32:19Z")

	info := Current()

	if !info.Complete {
		t.Fatal("Complete = false, want true")
	}
	if info.Version != lathe.Version || info.Commit != lathe.Commit || info.Date != lathe.Date {
		t.Fatalf("Current() = %#v", info)
	}
}

func TestCurrentUsesGoInstallModuleVersion(t *testing.T) {
	setLatheBuildInfo(t, "dev", "none", "unknown")
	setGoBuildInfo(t, "v1.2.3", "", "")

	info := Current()

	if !info.Complete {
		t.Fatal("Complete = false, want true for module version metadata")
	}
	if info.Version != "v1.2.3" {
		t.Fatalf("Version = %q, want v1.2.3", info.Version)
	}
	if info.Commit != "none" {
		t.Fatalf("Commit = %q, want none", info.Commit)
	}
	if info.Date != "unknown" {
		t.Fatalf("Date = %q, want unknown", info.Date)
	}
}

func TestCurrentUsesGoVCSMetadataForLocalBuilds(t *testing.T) {
	setLatheBuildInfo(t, "dev", "none", "unknown")
	setGoBuildInfo(t, "(devel)", "abcdef1234567890", "2026-06-25T10:32:19Z")

	info := Current()

	if !info.Complete {
		t.Fatal("Complete = false, want true for VCS metadata")
	}
	if info.Version != "dev" {
		t.Fatalf("Version = %q, want dev", info.Version)
	}
	if info.Commit != "abcdef1234567890" {
		t.Fatalf("Commit = %q, want VCS revision", info.Commit)
	}
	if info.Date != "2026-06-25T10:32:19Z" {
		t.Fatalf("Date = %q, want VCS time", info.Date)
	}
}

func TestApplyBackfillsLatheVersionFields(t *testing.T) {
	setLatheBuildInfo(t, "dev", "none", "unknown")
	setGoBuildInfo(t, "v1.2.3", "abcdef1234567890", "2026-06-25T10:32:19Z")

	Apply()

	if lathe.Version != "v1.2.3" {
		t.Fatalf("lathe.Version = %q, want v1.2.3", lathe.Version)
	}
	if lathe.Commit != "abcdef1234567890" {
		t.Fatalf("lathe.Commit = %q, want VCS revision", lathe.Commit)
	}
	if lathe.Date != "2026-06-25T10:32:19Z" {
		t.Fatalf("lathe.Date = %q, want VCS time", lathe.Date)
	}
}
