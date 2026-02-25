package main

import (
	"reflect"
	"testing"
)

func TestSplitCommands(t *testing.T) {
	in := "echo hi; ls -la && kubectl get pods || cat f | grep x & sleep 1"
	got := SplitCommandToSegments(in)
	want := []string{"echo hi", "ls -la", "kubectl get pods", "cat f", "grep x", "sleep 1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %#v got %#v", want, got)
	}
}

func TestPrecedenceAndOrder(t *testing.T) {
	rs := Ruleset{Rules: []Rule{
		{Name: "a", Allow: []string{"kubectl get .*"}},
		{Name: "b", Ask: []string{"kubectl get secrets"}},
		{Name: "c", Deny: []string{"kubectl get secret.*"}},
	}}
	res := EvaluateCommand(rs, "kubectl get secrets")
	if res.Decision != DecisionDeny {
		t.Fatalf("want deny got %s", res.Decision)
	}
	if res.Match == nil || res.Match.Rule != "c" {
		t.Fatalf("unexpected match: %+v", res.Match)
	}
}

func TestWorstAcrossSegments(t *testing.T) {
	rs := Ruleset{Rules: []Rule{
		{Name: "allow", Allow: []string{"echo .*"}},
		{Name: "ask", Ask: []string{"kubectl get secrets"}},
	}}
	res := EvaluateCommand(rs, "echo hi && kubectl get secrets")
	if res.Decision != DecisionAsk {
		t.Fatalf("want ask got %s", res.Decision)
	}
}

func TestPrefixAnchoring(t *testing.T) {
	rs := Ruleset{Rules: []Rule{{Name: "r", Allow: []string{"kubectl get .*"}}}}
	res := EvaluateCommand(rs, "please kubectl get pods")
	if res.Decision != DecisionAsk || res.Match != nil {
		t.Fatalf("expected no match and default ask, got %+v", res)
	}
}
