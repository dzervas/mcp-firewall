package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stderr, "mcp-firewall: ", 0)
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		logger.Fatalln(err)
		os.Exit(exitCode(err))
	}
}
