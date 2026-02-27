package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type CopilotRequest struct {
	Name string `json:"toolName"`
	Args string `json:"toolArgs"`
}

type CopilotToolArgs struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type CopilotResponse struct {
	Decision string `json:"permissionDecision"`
	Reason   string `json:"permissionDecisionReason"`
}

type CopilotAdapter struct{}

func (a *CopilotAdapter) DecodeRequest(r io.Reader) (string, error) {
	var req CopilotRequest
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return "", fmt.Errorf("copilot input JSON invalid: %w", err)
	}

	var toolArgs CopilotToolArgs
	if err := json.NewDecoder(strings.NewReader(req.Args)).Decode(&toolArgs); err != nil {
		return "", fmt.Errorf("copilot toolInput JSON invalid: %w", err)
	}

	if toolArgs.Command == "" {
		return "", fmt.Errorf("copilot input missing toolInput.command")
	}
	return toolArgs.Command, nil
}

func (a *CopilotAdapter) EncodeResponse(w io.Writer, decision, reason string) error {
	return json.NewEncoder(w).Encode(CopilotResponse{Decision: decision, Reason: reason})
}
