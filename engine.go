package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Decision is the engine result.
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionAsk   Decision = "ask"
	DecisionDeny  Decision = "deny"
)

// Match captures the matched rule and pattern.
type Match struct {
	Rule     string   `json:"rule"`
	Pattern  string   `json:"pattern"`
	Decision Decision `json:"decision"`
	Segment  string   `json:"segment"`
}

// EvalResult is the final evaluation output.
type EvalResult struct {
	Decision Decision `json:"decision"`
	Match    *Match   `json:"match,omitempty"`
}

func EvaluateCommand(rs Ruleset, command string) (EvalResult, string) {
	segments := splitCommands(command)
	if len(segments) == 0 {
		return EvalResult{Decision: DecisionAsk}, "no command segments found"
	}

	var best *Match
	bestRank := -1
	for _, seg := range segments {
		m := evaluateSegment(rs, seg)
		if m == nil {
			continue
		}
		rank := decisionRank(m.Decision)
		if rank > bestRank {
			best = m
			bestRank = rank
		}
	}

	if best == nil {
		return EvalResult{Decision: DecisionAsk}, "no matching rules found"
	}
	reason := fmt.Sprintf("%s due to rule %s", best.Decision, best.Rule)
	return EvalResult{Decision: best.Decision, Match: best}, reason
}

func evaluateSegment(rs Ruleset, segment string) *Match {
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

func decisionRank(d Decision) int {
	switch d {
	case DecisionAllow:
		return 0
	case DecisionAsk:
		return 1
	case DecisionDeny:
		return 2
	default:
		return -1
	}
}

func splitCommands(input string) []string {
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
