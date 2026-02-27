package main

import (
	"io"
	"log"
)

type Adapter interface {
	DecodeRequest(r io.Reader) (string, error)
	EncodeResponse(w io.Writer, decision, reason string) error
}

func RunAdapter(adapter Adapter, rs Ruleset, stdin io.Reader, stdout io.Writer) error {
	cmd, err := adapter.DecodeRequest(stdin)
	if err != nil {
		log.Fatalln("decode command:", err)
		return nil
	}

	res := EvaluateCommand(rs, cmd)
	return adapter.EncodeResponse(stdout, string(res.Decision), res.Reason)
}
