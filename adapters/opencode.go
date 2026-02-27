package adapters

import (
	"encoding/json"
	"io"
)

type OpenCodeRequest struct {
	Args OpenCodeRequestArgs `json:"args"`
}

type OpenCodeRequestArgs struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type OpenCodeResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// To test manually:
// echo -n '{"args":{"command":"git status","description":"Shows working tree status"}}' | base64 | go run . opencode
// plugin (.opencode/plugins/pretooluse-jsonnet.js):
// export const PretoolusePlugin = async ({ $ }) => {
// 	return {
// 		"tool.execute.before": async (_input, output) => {
// 			const outputStr = JSON.stringify(output);
// 			const proc = await $`echo -n '${btoa(outputStr)}' | /home/dzervas/Lab/pretooluse-jsonnet/pretooluse-jsonnet opencode`;
// 			if (proc.status !== 0) {
// 				throw new Error((proc.stderr || "pretooluse-jsonnet opencode failed").trim());
// 			}
// 			const result = JSON.parse(proc.stdout || "{}");
// 			if (result.decision === "allow") return;
// 			throw new Error(result.reason || `pretooluse-jsonnet: ${result.decision || "deny"}`);
// 		},
// 	};
// };

// Base64 encode the input (from the plugin side) to avoid escape issues
func DecodeOpenCodeCommand(r io.Reader) (string, error) {
	panic("OpenCode is not yet supported - plugin hooks can be bypassed by subagents (https://github.com/anomalyco/opencode/issues/5894) and there's no way to return an 'ask' response (force the UI to ask you for permission) ")

	// var req OpenCodeRequest
	// decoded := base64.NewDecoder(base64.StdEncoding, r)
	// if err := json.NewDecoder(decoded).Decode(&req); err != nil {
	// 	return "", fmt.Errorf("opencode input JSON invalid: %w", err)
	// }
	// if req.Args.Command == "" {
	// 	return "", fmt.Errorf("opencode input missing command")
	// }
	// return req.Args.Command, nil
}

func EncodeOpenCodeResponse(w io.Writer, decision, reason string) error {
	return json.NewEncoder(w).Encode(OpenCodeResponse{Decision: decision, Reason: reason})
}
