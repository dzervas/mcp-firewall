package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dzervas/pretooluse-jsonnet/adapters"
)

func testAdapter(t *testing.T, adapter Adapter, ruleset Ruleset, input string) map[string]any {
	in := strings.NewReader(input)
	var out bytes.Buffer
	if err := RunAdapter(adapter, ruleset, in, &out); err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestClaudeAdapter(t *testing.T) {
	ruleset := Ruleset{Rules: []Rule{{Name: "d", Deny: []string{"rm .*"}}}}
	input := `{"tool_input":{"command":"rm -rf /tmp/x"}}`
	out := testAdapter(t, &adapters.ClaudeAdapter{}, ruleset, input)

	if (out["hookSpecificOutput"].(map[string]any))["permissionDecision"] != "deny" {
		t.Fatalf("unexpected response: %v", out)
	}
}

func TestCopilotAdapter(t *testing.T) {
	ruleset := Ruleset{Rules: []Rule{{Name: "d", Deny: []string{"rm .*"}}}}
	input := `{"toolArgs":"{\"command\":\"rm -rf /tmp/x\"}"}`
	out := testAdapter(t, &adapters.CopilotAdapter{}, ruleset, input)

	if out["permissionDecision"] != "deny" {
		t.Fatalf("unexpected response: %v", out)
	}
}

// func TestOpenCodeAdapter(t *testing.T) {
// 	ruleset := Ruleset{Rules: []Rule{{Name: "a", Ask: []string{"kubectl get secrets"}}}}
// 	input := `{"command":"kubectl get secrets"}`
// 	out := testAdapter(t, &adapters.OpenCodeAdapter{}, ruleset, input)
// 	if out["decision"] != "ask" {
// 		t.Fatalf("unexpected response: %v", out)
// 	}
// }
