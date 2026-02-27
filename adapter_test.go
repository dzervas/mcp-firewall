package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dzervas/pretooluse-jsonnet/adapters"
)

func TestClaudeAdapter(t *testing.T) {
	rs := Ruleset{Rules: []Rule{{Name: "d", Deny: []string{"rm .*"}}}}
	in := strings.NewReader(`{"tool_input":{"command":"rm -rf /tmp/x"}}`)
	var out bytes.Buffer
	if err := RunAdapter(&adapters.ClaudeAdapter{}, rs, in, &out); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if (got["hookSpecificOutput"].(map[string]any))["permissionDecision"] != "deny" {
		t.Fatalf("unexpected response: %v", got)
	}
}

// func TestOpenCodeAdapter(t *testing.T) {
// 	rs := Ruleset{Rules: []Rule{{Name: "a", Ask: []string{"kubectl get secrets"}}}}
// 	in := strings.NewReader(`{"command":"kubectl get secrets"}`)
// 	var out bytes.Buffer
// 	if err := RunAdapter(&adapters.OpenCodeAdapter{}, rs, in, &out); err != nil {
// 		t.Fatal(err)
// 	}
// 	if !strings.Contains(out.String(), `"decision":"ask"`) {
// 		t.Fatalf("unexpected response: %s", out.String())
// 	}
// }

func TestCopilotAdapter(t *testing.T) {
	rs := Ruleset{Rules: []Rule{{Name: "d", Deny: []string{"rm .*"}}}}
	in := strings.NewReader(`{"toolArgs":"{\"command\":\"rm -rf /tmp/x\"}"}`)
	var out bytes.Buffer
	if err := RunAdapter(&adapters.CopilotAdapter{}, rs, in, &out); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["permissionDecision"] != "deny" {
		t.Fatalf("unexpected response: %v", got)
	}
}
