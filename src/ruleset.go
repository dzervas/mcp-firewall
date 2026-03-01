package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
)

// Rule is the internal per-rule representation.
type Rule struct {
	Name  string   `json:"name"`
	Allow []string `json:"allow"`
	Ask   []string `json:"ask"`
	Deny  []string `json:"deny"`
}

// Ruleset is the internal loaded representation.
type Ruleset struct {
	Rules []Rule
}

// Based on the provided ruleset, evaluate the command and return the decision and reason.
// It splits the commands in segments (e.g. "echo hi && kubectl get secrets" -> ["echo hi", "kubectl get secrets"]),
// finds the strictest matching rule for each segment and returns the strictest overall decision
// e.g. if any segment matches a deny rule, the overall decision is deny, else if any segment matches an ask rule,
// the overall decision is ask, else allow.
func (rs Ruleset) EvaluateCommand(command string) EvalResult {
	segments := SplitCommandToSegments(command)
	if len(segments) == 0 {
		return EvalResult{
			Decision: DefaultDecision,
			Reason:   "no command segments found",
		}
	}

	var matches []Match
	var strictestMatch *Match
	strictestRank := -1

	// For each segment, find the most restrictive matching rule.
	// For example in "echo hi && kubectl get secrets", "echo hi" matches an allow rule,
	// but "kubectl get secrets" matches an ask or deny, so the overall decision is ask or deny.
	for _, seg := range segments {
		segMatch := rs.EvaluateSegment(seg)
		if segMatch == nil {
			continue
		}

		matches = append(matches, *segMatch)

		segRank := strictestDecisionRank(segMatch.Decision)
		if segRank > strictestRank {
			strictestMatch = segMatch
			strictestRank = segRank
		}
	}

	// If no rules matched any segment, give the default decision.
	if len(matches) == 0 {
		return EvalResult{
			Decision: DefaultDecision,
			Reason:   "no matching rules found",
		}
	}

	return EvalResult{
		Decision: strictestMatch.Decision,
		Matches:  matches,
		Reason:   fmt.Sprintf("%s due to rule %s", strictestMatch.Decision, strictestMatch.RuleName),
	}
}

// Find the strictest (deny > ask > allow) matching rule for the given segment.
func (rs Ruleset) EvaluateSegment(segment string) *Match {
	for _, dec := range []Decision{DecisionDeny, DecisionAsk, DecisionAllow} {
		if m := rs.FindSegmentMatch(segment, dec); m != nil {
			return m
		}
	}

	return nil
}

// Given a segment and a specific decision (allow, ask, deny), find the first matching rule and return the match details.
func (rs Ruleset) FindSegmentMatch(segment string, d Decision) *Match {
	for _, rule := range rs.Rules {
		var patterns []string

		switch d {
		case DecisionAllow:
			patterns = rule.Allow
		case DecisionAsk:
			patterns = rule.Ask
		case DecisionDeny:
			patterns = rule.Deny
		default:
			log.Fatalln("Rule", rule.Name, "does not have a valid decision:", d)
		}

		for _, p := range patterns {
			re := regexp.MustCompile(anchorRegex(p))
			if re.MatchString(segment) {
				return &Match{RuleName: rule.Name, Pattern: p, Decision: d, Segment: segment}
			}
		}
	}
	return nil
}

func (rs Ruleset) DumpRulesetJSON() ([]byte, error) {
	result, err := json.MarshalIndent(rs, "", "  ")
	return result, err
}
