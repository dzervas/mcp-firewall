package adapters

import (
	"encoding/json"
	"fmt"
	"io"
)

// Add in global and/or project-specific settings (`.claude/settings.json`):
// ```json
// {
//   "hooks": {
//     "PreToolUse": [
//       {
//         "matcher": "",
//         "hooks": [
//           {
//             "type": "command",
//             "command": "pretooluse-jsonnet claude"
//           }
//         ]
//       }
//     ]
//   }
// }
// ```
//
// Test with:
// echo '{"cwd": "/home/dzervas/Lab", "tool_name": "Bash", "tool_input": {"command":"echo hello world"}}' | go run . claude

type ClaudeRequest struct {
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
}

type ClaudeResponse struct {
	HookOutput ClaudeHookSpecificOutput `json:"hookSpecificOutput"`
}

func NewClaudeResponse(decision, reason string) ClaudeResponse {
	return ClaudeResponse{
		HookOutput: ClaudeHookSpecificOutput{
			Name:     "PreToolUse",
			Decision: decision,
			Reason:   reason,
		},
	}
}

type ClaudeHookSpecificOutput struct {
	Name     string `json:"hookEventName"`
	Decision string `json:"permissionDecision"`
	Reason   string `json:"permissionDecisionReason"`
}

type ClaudeAdapter struct{}

func (a *ClaudeAdapter) DecodeRequest(r io.Reader) (string, error) {
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

func (a *ClaudeAdapter) EncodeResponse(w io.Writer, decision, reason string) error {
	return json.NewEncoder(w).Encode(NewClaudeResponse(decision, reason))
}
