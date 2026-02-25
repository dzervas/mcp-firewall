package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
)

const (
	envConfigDir = "PRETOOLUSE_CONFIG_DIR"
	envRuleset   = "PRETOOLUSE_RULESET"
)

// Rule is the internal per-rule representation.
type Rule struct {
	Name  string   `json:"-"`
	Allow []string `json:"allow"`
	Ask   []string `json:"ask"`
	Deny  []string `json:"deny"`
}

// Ruleset is the internal loaded representation.
type Ruleset struct {
	Rules []Rule
}

type configError struct {
	msg string
}

func (e configError) Error() string { return e.msg }

func defaultConfigDir() string {
	if v := strings.TrimSpace(os.Getenv(envConfigDir)); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Could not find the user's home directory:", err)
		return ""
	}
	return filepath.Join(home, ".config", "pretooluse")
}

func resolveProjectRuleset(cwd string) string {
	if v := strings.TrimSpace(os.Getenv(envRuleset)); v != "" {
		st, err := os.Stat(v)
		if err != nil {
			log.Fatalln("Ruleset override", v, "not readable:", err)
			return ""
		}
		if st.IsDir() {
			p := filepath.Join(v, ".pretooluse.jsonnet")
			if _, err := os.Stat(p); err != nil {
				log.Fatalln("Ruleset override dir", v, "missing .pretooluse.jsonnet:", err)
				return ""
			}
			return p
		}
		return v
	}

	local := filepath.Join(cwd, ".pretooluse.jsonnet")
	if _, err := os.Stat(local); err == nil {
		return local
	}

	root, err := gitRoot(cwd)
	if err == nil && root != "" {
		candidate := filepath.Join(root, ".pretooluse.jsonnet")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func gitRoot(cwd string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func loadRuleset(cwd string) (Ruleset, error) {
	cfgDir := defaultConfigDir()

	globalPath := filepath.Join(cfgDir, "config.jsonnet")
	projectPath := resolveProjectRuleset(cwd)

	globalRules, globalOrder, err := loadRuleMap(globalPath, cfgDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Ruleset{}, err
		}
		globalRules = map[string]Rule{}
		globalOrder = nil
	}

	projectRules := map[string]Rule{}
	var projectOrder []string
	if projectPath != "" {
		projectRules, projectOrder, err = loadRuleMap(projectPath, cfgDir)
		if err != nil {
			return Ruleset{}, err
		}
	}

	merged := mergeRules(globalRules, globalOrder, projectRules, projectOrder)
	if len(merged.Rules) == 0 {
		return Ruleset{}, configError{msg: "no rules found (neither global config.jsonnet nor project .pretooluse.jsonnet produced rules)"}
	}
	return merged, nil
}

func loadRuleMap(path, cfgDir string) (map[string]Rule, []string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if st.IsDir() {
		return nil, nil, configError{msg: fmt.Sprintf("ruleset path %q is a directory", path)}
	}

	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{
			filepath.Join(cfgDir, "lib"),
			filepath.Join(cfgDir, "vendor"),
			filepath.Dir(path),
			".",
		},
	})

	out, err := vm.EvaluateFile(path)
	if err != nil {
		return nil, nil, configError{msg: fmt.Sprintf("evaluate %s: %v", path, err)}
	}

	rules, order, err := decodeOrderedRules(out)
	if err != nil {
		return nil, nil, configError{msg: fmt.Sprintf("decode %s: %v", path, err)}
	}
	for _, name := range order {
		if err := validateRule(rules[name]); err != nil {
			return nil, nil, configError{msg: fmt.Sprintf("invalid rule %q in %s: %v", name, path, err)}
		}
	}
	return rules, order, nil
}

func decodeOrderedRules(raw string) (map[string]Rule, []string, error) {
	dec := json.NewDecoder(strings.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, nil, errors.New("top-level must be object")
	}

	rules := map[string]Rule{}
	order := []string{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil, err
		}
		name, ok := tok.(string)
		if !ok {
			return nil, nil, errors.New("rule key must be string")
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, nil, errors.New("rule name cannot be empty")
		}
		if _, exists := rules[name]; exists {
			return nil, nil, fmt.Errorf("duplicate rule name %q", name)
		}

		var obj map[string]json.RawMessage
		if err := dec.Decode(&obj); err != nil {
			return nil, nil, fmt.Errorf("rule %q must be object: %w", name, err)
		}
		r, err := decodeRule(name, obj)
		if err != nil {
			return nil, nil, err
		}
		rules[name] = r
		order = append(order, name)
	}

	tok, err = dec.Token()
	if err != nil {
		return nil, nil, err
	}
	delim, ok = tok.(json.Delim)
	if !ok || delim != '}' {
		return nil, nil, errors.New("top-level object not closed")
	}
	return rules, order, nil
}

func decodeRule(name string, obj map[string]json.RawMessage) (Rule, error) {
	r := Rule{Name: name, Allow: []string{}, Ask: []string{}, Deny: []string{}}
	for k, v := range obj {
		switch k {
		case "allow":
			if err := json.Unmarshal(v, &r.Allow); err != nil {
				return Rule{}, fmt.Errorf("rule %q allow must be array of strings", name)
			}
		case "ask":
			if err := json.Unmarshal(v, &r.Ask); err != nil {
				return Rule{}, fmt.Errorf("rule %q ask must be array of strings", name)
			}
		case "deny":
			if err := json.Unmarshal(v, &r.Deny); err != nil {
				return Rule{}, fmt.Errorf("rule %q deny must be array of strings", name)
			}
		default:
			return Rule{}, fmt.Errorf("rule %q has unknown field %q", name, k)
		}
	}
	r.Allow = ensureStringSlice(r.Allow)
	r.Ask = ensureStringSlice(r.Ask)
	r.Deny = ensureStringSlice(r.Deny)
	return r, nil
}

func ensureStringSlice(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

func validateRule(r Rule) error {
	for _, p := range r.Allow {
		if err := validatePattern(p); err != nil {
			return fmt.Errorf("allow pattern %q: %w", p, err)
		}
	}
	for _, p := range r.Ask {
		if err := validatePattern(p); err != nil {
			return fmt.Errorf("ask pattern %q: %w", p, err)
		}
	}
	for _, p := range r.Deny {
		if err := validatePattern(p); err != nil {
			return fmt.Errorf("deny pattern %q: %w", p, err)
		}
	}
	return nil
}

func validatePattern(p string) error {
	if strings.TrimSpace(p) == "" {
		return errors.New("cannot be empty")
	}
	_, err := regexp.Compile(anchorPattern(p))
	if err != nil {
		return err
	}
	return nil
}

func anchorPattern(p string) string {
	if strings.HasPrefix(p, "^") {
		return p
	}
	return "^" + p
}

func mergeRules(global map[string]Rule, globalOrder []string, project map[string]Rule, projectOrder []string) Ruleset {
	seen := make(map[string]bool, len(global)+len(project))
	out := make([]Rule, 0, len(global)+len(project))

	for _, n := range globalOrder {
		r, ok := global[n]
		if !ok {
			continue
		}
		if pr, ok := project[n]; ok {
			r = pr
		}
		out = append(out, r)
		seen[n] = true
	}

	for _, n := range projectOrder {
		if seen[n] {
			continue
		}
		r, ok := project[n]
		if !ok {
			continue
		}
		out = append(out, r)
		seen[n] = true
	}

	return Ruleset{Rules: out}
}

func (r Ruleset) AsMap() map[string]Rule {
	out := make(map[string]Rule, len(r.Rules))
	for _, rule := range r.Rules {
		copyRule := Rule{
			Allow: append([]string{}, rule.Allow...),
			Ask:   append([]string{}, rule.Ask...),
			Deny:  append([]string{}, rule.Deny...),
		}
		out[rule.Name] = copyRule
	}
	return out
}
