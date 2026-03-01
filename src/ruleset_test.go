package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRulesetMerge(t *testing.T) {
	t.Setenv(envConfigDir, t.TempDir())
	cfgDir := os.Getenv(envConfigDir)
	if err := os.MkdirAll(filepath.Join(cfgDir, "lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cfgDir, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}

	global := `[
  { "name": "g1", "allow": ["echo .*"], "deny": ["rm .*"], "ask": [] },
  { "name": "shared", "allow": ["kubectl get .*"], "ask": [], "deny": [] }
]`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.jsonnet"), []byte(global), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd := t.TempDir()
	project := `[
  { "name": "shared", "allow": [], "ask": ["kubectl get secrets"], "deny": [] },
  { "name": "p1", "allow": ["ls .*"], "ask": [], "deny": [] }
]`
	if err := os.WriteFile(filepath.Join(cwd, ".pretooluse.jsonnet"), []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}

	rs, err := LoadAllRulesets(cwd)
	if err != nil {
		t.Fatalf("loadRuleset error: %v", err)
	}
	if len(rs.Rules) != 4 {
		t.Fatalf("want 4 rules, got %d", len(rs.Rules))
	}
	if rs.Rules[0].Name != "g1" || rs.Rules[1].Name != "shared" || rs.Rules[2].Name != "shared" || rs.Rules[3].Name != "p1" {
		t.Fatalf("unexpected order: %+v", rs.Rules)
	}
}

func TestLoadRulesetRejectUnknownField(t *testing.T) {
	t.Setenv(envConfigDir, t.TempDir())
	cfgDir := os.Getenv(envConfigDir)
	if err := os.MkdirAll(filepath.Join(cfgDir, "lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cfgDir, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.jsonnet"), []byte(`[{"name":"r","allow":[],"ask":[],"deny":[],"oops":true}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAllRulesets(t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRulesetRejectInvalidRegex(t *testing.T) {
	t.Setenv(envConfigDir, t.TempDir())
	cfgDir := os.Getenv(envConfigDir)
	if err := os.MkdirAll(filepath.Join(cfgDir, "lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cfgDir, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.jsonnet"), []byte(`[{"name":"r","allow":["("],"ask":[],"deny":[]}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAllRulesets(t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
}
