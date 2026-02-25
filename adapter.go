package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dzervas/pretooluse-jsonnet/adapters"
)

type cliError struct {
	code int
	msg  string
}

func (e cliError) Error() string { return e.msg }

func exitCode(err error) int {
	var c cliError
	if errors.As(err, &c) {
		return c.code
	}
	var cfg configError
	if errors.As(err, &cfg) {
		return 2
	}
	return 1
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		log.Fatalln("usage: pretooluse <dump|test|claude|copilot|opencode>")
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("getwd:", err)
		return nil
	}

	rs, err := loadRuleset(cwd)
	if err != nil {
		return err
	}

	switch args[0] {
	case "dump":
		b, err := dumpRulesetJSON(rs)
		if err != nil {
			return err
		}
		_, err = stdout.Write(b)
		return err
	case "test":
		if len(args) < 2 {
			log.Fatalln("usage: pretooluse test <arg...>")
		}

		command := strings.Join(args[1:], " ")
		result, reason := EvaluateCommand(rs, command)
		fmt.Println("reason:", reason)
		if _, err := fmt.Fprintln(stdout, result.Decision); err != nil {
			return err
		}
		if result.Match != nil {
			_, err = fmt.Fprintf(stderr, "rule=%s decision=%s pattern=%s segment=%s\n", result.Match.Rule, result.Match.Decision, result.Match.Pattern, result.Match.Segment)
		} else {
			_, err = fmt.Fprintln(stderr, "rule=<none>")
		}
		return err
	case "claude":
		return runClaudeAdapter(rs, stdin, stdout)
	case "copilot":
		return runCopilotAdapter(rs, stdin, stdout)
	case "opencode":
		return runOpenCodeAdapter(rs, stdin, stdout)
	default:
		log.Fatalln("Unknown subcommand:", args[0])
		return nil
	}
}

func dumpRulesetJSON(rs Ruleset) ([]byte, error) {
	result, err := json.MarshalIndent(rs, "", "  ")
	return result, err
}

func runClaudeAdapter(rs Ruleset, stdin io.Reader, stdout io.Writer) error {
	cmd, err := adapters.DecodeClaudeCommand(stdin)
	if err != nil {
		log.Fatalln("decode claude command:", err)
		return nil
	}

	res, reason := EvaluateCommand(rs, cmd)
	return adapters.EncodeClaudeResponse(stdout, string(res.Decision), reason)
}

func runCopilotAdapter(rs Ruleset, stdin io.Reader, stdout io.Writer) error {
	cmd, err := adapters.DecodeCopilotCommand(stdin)
	if err != nil {
		log.Fatalln("decode copilot command:", err)
		return nil
	}
	res, reason := EvaluateCommand(rs, cmd)
	return adapters.EncodeCopilotResponse(stdout, string(res.Decision), reason)
}

func runOpenCodeAdapter(rs Ruleset, stdin io.Reader, stdout io.Writer) error {
	cmd, err := adapters.DecodeOpenCodeCommand(stdin)
	if err != nil {
		log.Fatalln("decode opencode command:", err)
		return nil
	}
	res, reason := EvaluateCommand(rs, cmd)
	return adapters.EncodeOpenCodeResponse(stdout, string(res.Decision), reason)
}
