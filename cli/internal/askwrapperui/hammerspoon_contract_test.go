package askwrapperui

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestHammerspoonLaunchesAskwrapperOnly(t *testing.T) {
	// @req CLI-ASKWRAPPER-025
	// @req CLI-ASKWRAPPER-026
	// @req CLI-ASKWRAPPER-027
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	luaPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../../hammerspoon/modules/askwrapper.lua"))
	raw, err := os.ReadFile(luaPath)
	if err != nil {
		t.Fatalf("read lua module: %v", err)
	}
	content := string(raw)

	if !strings.Contains(content, `launchAskwrapper({ "ui", "ask" })`) {
		t.Fatal("missing alt+shift+A askwrapper ui ask launcher")
	}
	if !strings.Contains(content, `launchAskwrapper({ "ui", "followup" })`) {
		t.Fatal("missing alt+shift+S askwrapper ui followup launcher")
	}

	if strings.Contains(content, "runAsk") {
		t.Fatal("hammerspoon module appears to embed ask execution logic")
	}

	blocked := []string{
		`hs\.task\.new\([^\n]*"ask"`,
		`hs\.task\.new\([^\n]*'ask'`,
		`launchAskwrapper\(\s*{\s*"ask"`,
		`launchAskwrapper\(\s*{\s*'ask'`,
	}
	for _, pattern := range blocked {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			t.Fatalf("hammerspoon module appears to launch ask directly: pattern %q", pattern)
		}
	}
}
