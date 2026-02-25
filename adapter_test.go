package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestClaudeAdapter(t *testing.T) {
	rs := Ruleset{Rules: []Rule{{Name: "d", Deny: []string{"rm .*"}}}}
	in := strings.NewReader(`{"tool_input":{"command":"rm -rf /tmp/x"}}`)
	var out bytes.Buffer
	if err := runClaudeAdapter(rs, in, &out); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["decision"] != "deny" {
		t.Fatalf("unexpected response: %v", got)
	}
}

func TestOpenCodeAdapter(t *testing.T) {
	rs := Ruleset{Rules: []Rule{{Name: "a", Ask: []string{"kubectl get secrets"}}}}
	in := strings.NewReader(`{"command":"kubectl get secrets"}`)
	var out bytes.Buffer
	if err := runOpenCodeAdapter(rs, in, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"decision":"ask"`) {
		t.Fatalf("unexpected response: %s", out.String())
	}
}
