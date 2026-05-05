package hook

// Input is the JSON payload Crush sends on stdin.
type Input struct {
	Event     string `json:"event"`
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// Output is the JSON envelope returned on stdout.
type Output struct {
	Version  int    `json:"version"`
	Decision string `json:"decision,omitempty"` // "allow", "deny", or omit
	Reason   string `json:"reason,omitempty"`
}
