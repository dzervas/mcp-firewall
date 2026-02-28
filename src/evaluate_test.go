package main

import (
	"reflect"
	"testing"
)

// TODO: Define many test cases
// type Case struct {
// 	ruleset  Ruleset
// 	commands []CaseCommand
// }

// type CaseCommand struct {
// 	command  string
// 	decision Decision
// }

// var cases = []Case{
// 	{
// 		ruleset: Ruleset{Rules: []Rule{{
// 			Name:  "single",
// 			Allow: []string{"echo "},
// 			Ask:   []string{"kubectl get secrets"},
// 			Deny:  []string{"rm "},
// 		}}},
// 		commands: []CaseCommand{
// 			CaseCommand{"echo hi", DecisionAllow},
// 			CaseCommand{"kubectl get secrets", DecisionAsk},
// 			CaseCommand{"rm -rf /tmp/x", DecisionDeny},
// 		},
// 	},
// }

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
	res := rs.EvaluateCommand("kubectl get secrets")
	if res.Decision != DecisionDeny {
		t.Fatalf("want deny got %s", res.Decision)
	}
	if len(res.Matches) != 1 || res.Matches[0].RuleName != "c" {
		t.Fatalf("unexpected match: %+v", res.Matches)
	}
}

func TestWorstAcrossSegments(t *testing.T) {
	rs := Ruleset{Rules: []Rule{
		{Name: "allow", Allow: []string{"echo .*"}},
		{Name: "ask", Ask: []string{"kubectl get secrets"}},
	}}
	res := rs.EvaluateCommand("echo hi && kubectl get secrets")
	if res.Decision != DecisionAsk {
		t.Fatalf("want ask got %s", res.Decision)
	}
}

func TestPrefixAnchoring(t *testing.T) {
	rs := Ruleset{Rules: []Rule{{Name: "r", Allow: []string{"kubectl get .*"}}}}
	res := rs.EvaluateCommand("please kubectl get pods")
	if res.Decision != DecisionAsk || len(res.Matches) != 0 {
		t.Fatalf("expected no match and default ask, got %+v", res)
	}
}
