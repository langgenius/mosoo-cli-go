package buildinfo

import (
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

func TestCurrentUsesDeterministicDefaults(t *testing.T) {
	setLatheBuildInfo(t, "", "", "")

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

	info := Current()

	if !info.Complete {
		t.Fatal("Complete = false, want true")
	}
	if info.Version != lathe.Version || info.Commit != lathe.Commit || info.Date != lathe.Date {
		t.Fatalf("Current() = %#v", info)
	}
}
