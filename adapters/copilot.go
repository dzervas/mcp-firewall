package adapters

import (
	"encoding/json"
	"fmt"
	"io"
)

type CopilotRequest struct {
	ToolInput map[string]any `json:"toolInput"`
}

type CopilotResponse struct {
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

func DecodeCopilotCommand(r io.Reader) (string, error) {
	var req CopilotRequest
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return "", fmt.Errorf("copilot input JSON invalid: %w", err)
	}
	cmd, _ := req.ToolInput["command"].(string)
	if cmd == "" {
		return "", fmt.Errorf("copilot input missing toolInput.command")
	}
	return cmd, nil
}

func EncodeCopilotResponse(w io.Writer, decision, reason string) error {
	return json.NewEncoder(w).Encode(CopilotResponse{PermissionDecision: decision, PermissionDecisionReason: reason})
}
