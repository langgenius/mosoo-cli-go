package buildinfo

import "testing"

func TestCurrentUsesDeterministicDefaults(t *testing.T) {
	Version = ""
	Commit = ""
	Date = ""

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
	Version = "v1.2.3"
	Commit = "abcdef123456"
	Date = "2026-06-25T10:32:19Z"

	info := Current()

	if !info.Complete {
		t.Fatal("Complete = false, want true")
	}
	if info.Version != Version || info.Commit != Commit || info.Date != Date {
		t.Fatalf("Current() = %#v", info)
	}
}
