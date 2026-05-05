package hook

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/arnarg/crushout/internal/rules"
)

// Hook is the protocol-agnostic interface for hook input.
type Hook interface {
	EventName() string
	CWD() string
	ToolName() string
	Command() string
	SessionID() string
	FormatDecision(d rules.Decision, reason string) ([]byte, error)
}

// CrushInput is the JSON payload Crush sends on stdin.
type CrushInput struct {
	EventNameField string `json:"event"`
	SessionIDField string `json:"session_id"`
	CwdField       string `json:"cwd"`
	ToolNameField  string `json:"tool_name"`
	ToolInput      struct {
		CommandField string `json:"command"`
	} `json:"tool_input"`
}

func (i *CrushInput) EventName() string { return i.EventNameField }
func (i *CrushInput) CWD() string       { return i.CwdField }
func (i *CrushInput) ToolName() string  { return i.ToolNameField }
func (i *CrushInput) Command() string   { return i.ToolInput.CommandField }
func (i *CrushInput) SessionID() string { return i.SessionIDField }
func (i *CrushInput) FormatDecision(d rules.Decision, reason string) ([]byte, error) {
	switch d {
	case rules.Allow:
		return json.Marshal(&CrushOutput{Version: 1, Decision: "allow"})
	case rules.Deny:
		return json.Marshal(&CrushOutput{Version: 1, Decision: "deny", Reason: reason})
	default:
		return []byte("{}"), nil
	}
}

// CrushOutput is the JSON envelope returned on stdout for Crush.
type CrushOutput struct {
	Version  int    `json:"version"`
	Decision string `json:"decision,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// ClaudeInput is the JSON payload Claude Code sends on stdin.
type ClaudeInput struct {
	SessionIDField      string `json:"session_id"`
	TranscriptPathField string `json:"transcript_path"`
	CwdField            string `json:"cwd"`
	PermissionModeField string `json:"permission_mode,omitempty"`
	HookEventNameField  string `json:"hook_event_name"`
	ToolNameField       string `json:"tool_name"`
	ToolInput           struct {
		CommandField string `json:"command"`
	} `json:"tool_input"`
}

func (i *ClaudeInput) EventName() string { return i.HookEventNameField }
func (i *ClaudeInput) CWD() string       { return i.CwdField }
func (i *ClaudeInput) ToolName() string  { return i.ToolNameField }
func (i *ClaudeInput) Command() string   { return i.ToolInput.CommandField }
func (i *ClaudeInput) SessionID() string { return i.SessionIDField }
func (i *ClaudeInput) FormatDecision(d rules.Decision, reason string) ([]byte, error) {
	switch d {
	case rules.Allow:
		return json.Marshal(&ClaudeOutput{
			HookSpecificOutput: &ClaudeHookSpecificOutput{
				HookEventName:      i.HookEventNameField,
				PermissionDecision: "allow",
			},
		})
	case rules.Deny:
		return json.Marshal(&ClaudeOutput{
			HookSpecificOutput: &ClaudeHookSpecificOutput{
				HookEventName:            i.HookEventNameField,
				PermissionDecision:       "deny",
				PermissionDecisionReason: reason,
			},
		})
	default:
		return json.Marshal(&ClaudeOutput{
			HookSpecificOutput: &ClaudeHookSpecificOutput{
				HookEventName:      i.HookEventNameField,
				PermissionDecision: "ask",
			},
		})
	}
}

// ClaudeOutput is the JSON envelope returned on stdout for Claude Code.
type ClaudeOutput struct {
	HookSpecificOutput *ClaudeHookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// ClaudeHookSpecificOutput contains the event-specific output fields for Claude Code.
type ClaudeHookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
}

// Decode reads JSON from r and returns the appropriate Hook implementation.
func Decode(r io.Reader) (Hook, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var probe struct {
		Event         string `json:"event"`
		HookEventName string `json:"hook_event_name"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}

	if probe.HookEventName != "" {
		var input ClaudeInput
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, err
		}
		return &input, nil
	}

	var input CrushInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, err
	}
	return &input, nil
}

// IsBashTool returns true if the hook is for a bash command.
func IsBashTool(h Hook) bool {
	return strings.ToLower(h.ToolName()) == "bash" && h.Command() != ""
}
