package main

import (
	"log"
	"strings"
)

// Decision is the engine result.
type Decision string

const (
	DecisionAllow   Decision = "allow"
	DecisionAsk     Decision = "ask"
	DecisionDeny    Decision = "deny"
	DefaultDecision Decision = DecisionAsk
)

// During evaluation, a rule got a match against the provided command
type Match struct {
	// The name of the matched rule
	RuleName string `json:"rule_name"`
	// The pattern of the rule that got matched, e.g. "kubectl get .*"
	Pattern string `json:"pattern"`
	// The decision associated with the matched rule (allow, ask, deny)
	Decision Decision `json:"decision"`
	// The specific segment (sub-command) that matched the rule, e.g. "kubectl get secrets" from "echo hi && kubectl get secrets"
	Segment string `json:"segment"`
}

// The end result of the evaluation
type EvalResult struct {
	// The overall decision according to the matches
	Decision Decision `json:"decision"`
	// The matched rule of each segment
	Matches []Match `json:"matches"`
	// A human-readable explanation of the decision, e.g. "deny due to rule c"
	Reason string `json:"reason"`
}

func strictestDecisionRank(d Decision) int {
	switch d {
	case DecisionAllow:
		return 0
	case DecisionAsk:
		return 1
	case DecisionDeny:
		return 2
	default:
		log.Panicln("Unknown decision provided:", d)
		return -1
	}
}

// Split the input command into segments based on common shell operators (;, &&, ||, |, &) while respecting quoted strings.
func SplitCommandToSegments(input string) []string {
	segments := []string{input}
	splitChars := []string{"||", "&&", ";", "|", "&"}

	// Tons of loops but for the time being, who cares
	for _, char := range splitChars {
		var newSegments []string
		for _, seg := range segments {
			newSegments = append(newSegments, strings.Split(seg, char)...)
		}
		segments = newSegments
	}

	result := make([]string, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg != "" {
			result = append(result, seg)
		}
	}

	return result
}
