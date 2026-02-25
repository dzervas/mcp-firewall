package adapters

import (
	"encoding/json"
	"fmt"
	"io"
)

type OpenCodeRequest struct {
	Command string `json:"command"`
}

type OpenCodeResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func DecodeOpenCodeCommand(r io.Reader) (string, error) {
	var req OpenCodeRequest
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return "", fmt.Errorf("opencode input JSON invalid: %w", err)
	}
	if req.Command == "" {
		return "", fmt.Errorf("opencode input missing command")
	}
	return req.Command, nil
}

func EncodeOpenCodeResponse(w io.Writer, decision, reason string) error {
	return json.NewEncoder(w).Encode(OpenCodeResponse{Decision: decision, Reason: reason})
}
