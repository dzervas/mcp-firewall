package adapters

import (
	"encoding/json"
	"fmt"
	"io"
)

type ClaudeRequest struct {
	ToolInput map[string]any `json:"tool_input"`
}

type ClaudeResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func DecodeClaudeCommand(r io.Reader) (string, error) {
	var req ClaudeRequest
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return "", fmt.Errorf("claude input JSON invalid: %w", err)
	}
	cmd, _ := req.ToolInput["command"].(string)
	if cmd == "" {
		return "", fmt.Errorf("claude input missing tool_input.command")
	}
	return cmd, nil
}

func EncodeClaudeResponse(w io.Writer, decision, reason string) error {
	return json.NewEncoder(w).Encode(ClaudeResponse{Decision: decision, Reason: reason})
}
