package providers

import (
	"encoding/json"
	"strings"
)

var logOnly string

func LogOnly(prefix string) {
	logOnly = prefix
}

func LogDebug(msg string, data any) {
	if logOnly != "" && !strings.HasPrefix(msg, logOnly) {
		return
	}

	if server == nil || server.Log.GetMaxLevel() < 2 {
		return
	}

	str, _ := json.MarshalIndent(data, "", "  ")
	server.Log.Debugf(msg, str)
}

func Debugf(msg string, args ...any) {
	server.Log.Debugf(msg, args...)
}
