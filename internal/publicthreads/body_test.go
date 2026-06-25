package publicthreads

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCreateBodySetsAndStrings(t *testing.T) {
	raw, err := buildCreateBody("", []string{
		"input.type=user.message",
		"input.content[0].type=text",
	}, []string{
		"input.content[0].text=Say hello",
		"client_external_ref=demo-001",
	})
	if err != nil {
		t.Fatalf("buildCreateBody: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("invalid JSON %q: %v", raw, err)
	}
	input, _ := got["input"].(map[string]any)
	if input["type"] != "user.message" {
		t.Fatalf("input.type = %#v", input["type"])
	}
	content, _ := input["content"].([]any)
	if len(content) != 1 {
		t.Fatalf("content = %#v", input["content"])
	}
	first, _ := content[0].(map[string]any)
	if first["type"] != "text" || first["text"] != "Say hello" {
		t.Fatalf("content[0] = %#v", first)
	}
	if got["client_external_ref"] != "demo-001" {
		t.Fatalf("client_external_ref = %#v", got["client_external_ref"])
	}
}

func TestBuildCreateBodyInfersTypesButSetStrForcesString(t *testing.T) {
	raw, err := buildCreateBody("", []string{"a=3", "b=true"}, []string{"c=3"})
	if err != nil {
		t.Fatalf("buildCreateBody: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["a"] != float64(3) { // JSON numbers decode to float64
		t.Fatalf("a = %#v, want number", got["a"])
	}
	if got["b"] != true {
		t.Fatalf("b = %#v, want bool", got["b"])
	}
	if got["c"] != "3" {
		t.Fatalf("c = %#v, want string", got["c"])
	}
}

func TestBuildCreateBodyFileFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	if err := os.WriteFile(path, []byte(`{"client_external_ref":"from-file"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err := buildCreateBody(path, nil, nil)
	if err != nil {
		t.Fatalf("buildCreateBody: %v", err)
	}
	if string(raw) != `{"client_external_ref":"from-file"}` {
		t.Fatalf("file body = %q", raw)
	}
}

func TestBuildCreateBodyEmpty(t *testing.T) {
	raw, err := buildCreateBody("", nil, nil)
	if err != nil {
		t.Fatalf("buildCreateBody: %v", err)
	}
	if raw != nil {
		t.Fatalf("expected nil body, got %q", raw)
	}
}

func TestBuildCreateBodyInvalidSet(t *testing.T) {
	if _, err := buildCreateBody("", []string{"noequals"}, nil); err == nil {
		t.Fatal("expected error for malformed --set")
	}
}
