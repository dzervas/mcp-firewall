package main

import (
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
		b, err := rs.DumpRulesetJSON()
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
		result := rs.EvaluateCommand(command)
		if _, err := fmt.Fprintln(stdout, result.Decision); err != nil {
			return err
		}
		log.Println("reason:", result.Reason)
		if len(result.Matches) == 0 {
			log.Println("no matching rules")
		}

		for i, m := range result.Matches {
			log.Printf("match %d: rule=%s decision=%s pattern=%s segment=%s\n", i+1, m.RuleName, m.Decision, m.Pattern, m.Segment)
		}
		return err
	case "claude":
		return RunAdapter(&adapters.ClaudeAdapter{}, rs, stdin, stdout)
	case "copilot":
		return RunAdapter(&adapters.CopilotAdapter{}, rs, stdin, stdout)
	case "opencode":
		return RunAdapter(&adapters.OpenCodeAdapter{}, rs, stdin, stdout)
	default:
		log.Fatalln("Unknown subcommand:", args[0])
		return nil
	}
}
