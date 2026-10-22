package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpCommandsHeadingIsAligned(t *testing.T) {
	var buf bytes.Buffer
	printHelp(&buf, "v1.2.3")
	if !strings.Contains(buf.String(), "\nCOMMANDS\n  install") {
		t.Fatalf("help output has inconsistent command indentation:\n%s", buf.String())
	}
}

func TestHelpDoesNotAdvertiseRetiredVerifyValidator(t *testing.T) {
	var output bytes.Buffer
	printHelp(&output, "test")
	if strings.Contains(output.String(), "sdd-verify-validate") {
		t.Fatal("help advertises retired report admission")
	}
}
