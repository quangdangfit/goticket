package service

import (
	"strings"
	"testing"
)

// Sanity: the lua sources are embedded and non-empty.
func TestEmbeddedLuaScripts(t *testing.T) {
	for name, s := range map[string]string{
		"hold":    holdScript,
		"release": releaseScript,
		"confirm": confirmScript,
	} {
		if strings.TrimSpace(s) == "" {
			t.Fatalf("%s script empty", name)
		}
	}
}
