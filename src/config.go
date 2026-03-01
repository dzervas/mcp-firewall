package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
)

const (
	envConfigDir       = "PRETOOLUSE_CONFIG_DIR"
	envRuleset         = "PRETOOLUSE_RULESET"
	projectRulesetFile = ".pretooluse.jsonnet"
	globalRulesetFile  = "config.jsonnet"
)

type configError struct {
	msg string
}

func (e configError) Error() string { return e.msg }

func GlobalConfigDir() string {
	configDirOverride := strings.TrimSpace(os.Getenv(envConfigDir))
	if configDirOverride != "" {
		return configDirOverride
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Could not find the user's home directory:", err)
		return ""
	}
	return filepath.Join(home, ".config", "pretooluse")
}

// Get the efftive path of the project's ruleset in the following lookup order:
// 1. PRETOOLUSE_RULESET env var path
// 2. .pretooluse.jsonnet in the current working directory
// 3. .pretooluse.jsonnet in the git root of the current working directory
// If none of the above exist, returns an empty string.
func ResolveProjectRuleset(cwd string) string {
	projectRulesetOverride := strings.TrimSpace(os.Getenv(envRuleset))
	if projectRulesetOverride != "" {
		return projectRulesetOverride
	}

	local := filepath.Join(cwd, projectRulesetFile)
	if _, err := os.Stat(local); err == nil {
		return local
	}

	root, err := gitRoot(cwd)
	if err == nil && root != "" {
		candidate := filepath.Join(root, projectRulesetFile)
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

func LoadAllRulesets(cwd string) (Ruleset, error) {
	cfgDir := GlobalConfigDir()

	globalPath := filepath.Join(cfgDir, globalRulesetFile)
	projectPath := ResolveProjectRuleset(cwd)

	globalRules, err := LoadRuleMap(globalPath, cfgDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Ruleset{}, err
		}
		globalRules = []Rule{}
	}

	projectRules := []Rule{}
	if projectPath != "" {
		projectRules, err = LoadRuleMap(projectPath, cfgDir)
		if err != nil {
			return Ruleset{}, err
		}
	}

	merged := Ruleset{Rules: append(globalRules, projectRules...)}
	if len(merged.Rules) == 0 {
		return Ruleset{}, fmt.Errorf("no rulesets found in global path %q or project path %q", globalPath, projectPath)
	}
	return merged, nil
}

func LoadRuleMap(path, cfgDir string) ([]Rule, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if st.IsDir() {
		return nil, fmt.Errorf("ruleset path %q is a directory", path)
	}

	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{
			filepath.Join(cfgDir, "lib"),
			filepath.Join(cfgDir, "vendor"),
			filepath.Dir(path),
		},
	})

	out, err := vm.EvaluateFile(path)
	if err != nil {
		return nil, fmt.Errorf("evaluate %s: %v", path, err)
	}

	rules, err := DecodeRuleMap(out)
	return rules, err
}

func DecodeRuleMap(raw string) ([]Rule, error) {
	var result []Rule

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode ruleset JSON: %v", err)
	}

	// Validate that all patterns are valid regexes
	for name, rule := range result {
		for _, p := range rule.Allow {
			if _, err := regexp.Compile(p); err != nil {
				return nil, fmt.Errorf("invalid regex in allow pattern %q of rule %q: %v", p, name, err)
			}
		}
		for _, p := range rule.Ask {
			if _, err := regexp.Compile(p); err != nil {
				return nil, fmt.Errorf("invalid regex in allow pattern %q of rule %q: %v", p, name, err)
			}
		}
		for _, p := range rule.Deny {
			if _, err := regexp.Compile(p); err != nil {
				return nil, fmt.Errorf("invalid regex in allow pattern %q of rule %q: %v", p, name, err)
			}
		}
	}

	return result, nil
}

func anchorRegex(p string) string {
	if strings.HasPrefix(p, "^") {
		return p
	}
	return "^" + p
}

func mergeRules(global map[string]Rule, project map[string]Rule) Ruleset {
	out := make(map[string]Rule, len(global)+len(project))

	maps.Copy(out, global)
	maps.Copy(out, project)

	outList := make([]Rule, 0, len(out))
	for _, rule := range out {
		outList = append(outList, rule)
	}

	return Ruleset{Rules: outList}
}
