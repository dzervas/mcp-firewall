package main

import (
	"io"
	"log"
)

// The "wrapper" to interface the command evaluation results to a specific agent
type Adapter interface {
	DecodeRequest(r io.Reader) (string, error)
	EncodeResponse(w io.Writer, decision, reason string) error
}

// Evaluate the provided input against the adapter, evaluate any matching commands and return the result to stdout
func RunAdapter(adapter Adapter, rs Ruleset, stdin io.Reader, stdout io.Writer) error {
	cmd, err := adapter.DecodeRequest(stdin)
	if err != nil {
		log.Fatalln("decode command:", err)
		return nil
	}

	res := rs.EvaluateCommand(cmd)
	return adapter.EncodeResponse(stdout, string(res.Decision), res.Reason)
}
