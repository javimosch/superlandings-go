package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// Semantic exit codes (80-119 range per agent-friendly design)
const (
	ExitSuccess       = 0
	ExitMissingFlag   = 81
	ExitInvalidInput  = 82
	ExitNotFound      = 90
	ExitAlreadyExists = 91
	ExitConflict      = 92
	ExitExtMissing    = 100
	ExitExtFailed     = 101
	ExitInternal      = 110
)

func exitType(code int) string {
	switch {
	case code >= 80 && code < 90:
		return "invalid_input"
	case code >= 90 && code < 100:
		return "resource_error"
	case code >= 100 && code < 110:
		return "external_error"
	case code >= 110:
		return "internal_error"
	default:
		return "unknown"
	}
}

func exitRecoverable(code int) bool {
	switch code {
	case ExitExtFailed:
		return true
	default:
		return false
	}
}

// fail prints a structured JSON error to stderr and exits with the semantic code.
func fail(code int, msg string) {
	err := map[string]interface{}{
		"code":        code,
		"type":        exitType(code),
		"message":     msg,
		"recoverable": exitRecoverable(code),
	}
	out := map[string]interface{}{
		"version": "1.0",
		"success": false,
		"error":   err,
	}
	data, _ := json.Marshal(out)
	fmt.Fprintln(os.Stderr, string(data))
	os.Exit(code)
}

// failf is like fail but with format string.
func failf(code int, format string, args ...interface{}) {
	fail(code, fmt.Sprintf(format, args...))
}

// success prints a structured JSON success response.
func success(message string, extra map[string]interface{}) {
	out := map[string]interface{}{
		"version": "1.0",
		"success": true,
	}
	if message != "" {
		out["message"] = message
	}
	for k, v := range extra {
		out[k] = v
	}
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
}

// writeJSON marshals and writes a map as JSON to stdout.
func writeJSON(out map[string]interface{}) {
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
}
