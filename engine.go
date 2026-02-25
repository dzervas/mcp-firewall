package main

import (
	"fmt"
	"log"
	"regexp"
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

// Match captures the matched rule and pattern.
type Match struct {
	Rule     string   `json:"rule"`
	Pattern  string   `json:"pattern"`
	Decision Decision `json:"decision"`
	Segment  string   `json:"segment"`
}

type EvalResult struct {
	Decision Decision `json:"decision"`
	Match    *Match   `json:"match,omitempty"`
	Reason   string   `json:"reason"`
}

// Based on the provided ruleset, evaluate the command and return the decision and reason.
// It splits the commands in segments (e.g. "echo hi && kubectl get secrets" -> ["echo hi", "kubectl get secrets"]),
// finds the strictest matching rule for each segment and returns the strictest overall decision
// e.g. if any segment matches a deny rule, the overall decision is deny, else if any segment matches an ask rule,
// the overall decision is ask, else allow.
func EvaluateCommand(rs Ruleset, command string) EvalResult {
	segments := SplitCommandToSegments(command)
	if len(segments) == 0 {
		return EvalResult{
			Decision: DefaultDecision,
			Reason:   "no command segments found",
		}
	}

	var strictest *Match
	strictestRank := -1
	// For each segment, find the most restrictive matching rule.
	// For example in "echo hi && kubectl get secrets", "echo hi" matches an allow rule,
	// but "kubectl get secrets" matches an ask or deny, so the overall decision is ask or deny.
	for _, seg := range segments {
		segMatch := EvaluateSegment(rs, seg)
		if segMatch == nil {
			continue
		}
		segRank := strictestDecisionRank(segMatch.Decision)
		if segRank > strictestRank {
			strictest = segMatch
			strictestRank = segRank
		}
	}

	// If no rules matched any segment, give the default decision.
	if strictest == nil {
		return EvalResult{
			Decision: DefaultDecision,
			Reason:   "no matching rules found",
		}
	}

	return EvalResult{
		Decision: strictest.Decision,
		Match:    strictest,
		Reason:   fmt.Sprintf("%s due to rule %s", strictest.Decision, strictest.Rule),
	}
}

// Find the strictest (deny > ask > allow) matching rule for the given segment.
func EvaluateSegment(rs Ruleset, segment string) *Match {
	if m := firstMatch(rs, segment, DecisionDeny); m != nil {
		return m
	}
	if m := firstMatch(rs, segment, DecisionAsk); m != nil {
		return m
	}
	if m := firstMatch(rs, segment, DecisionAllow); m != nil {
		return m
	}
	return nil
}

func firstMatch(rs Ruleset, segment string, d Decision) *Match {
	for _, rule := range rs.Rules {
		var patterns []string
		switch d {
		case DecisionAllow:
			patterns = rule.Allow
		case DecisionAsk:
			patterns = rule.Ask
		case DecisionDeny:
			patterns = rule.Deny
		}

		for _, p := range patterns {
			re := regexp.MustCompile(anchorPattern(p))
			if re.MatchString(segment) {
				return &Match{Rule: rule.Name, Pattern: p, Decision: d, Segment: segment}
			}
		}
	}
	return nil
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
	parts := splitByOps(input)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func splitByOps(input string) []string {
	var out []string
	start := 0
	for i := 0; i < len(input); i++ {
		switch input[i] {
		case ';':
			out = append(out, input[start:i])
			start = i + 1
		case '|':
			out = append(out, input[start:i])
			if i+1 < len(input) && input[i+1] == '|' {
				i++
			}
			start = i + 1
		case '&':
			out = append(out, input[start:i])
			if i+1 < len(input) && input[i+1] == '&' {
				i++
			}
			start = i + 1
		}
	}
	out = append(out, input[start:])
	return out
}
