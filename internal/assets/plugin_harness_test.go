package assets

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func runOpenCodeTransportPluginHarness(t *testing.T, modules map[string]string, harness, relay string, extraPath ...string) (string, string, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the relay fixture uses a POSIX shell")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is unavailable")
	}
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, source := range modules {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "harness.mts"), []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	if relay != "" {
		if err := os.WriteFile(filepath.Join(bin, "gentle-ai"), []byte(relay), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	logPath := filepath.Join(root, "relay.log")
	probePath := filepath.Join(root, "probe.log")
	command := exec.Command(node, "harness.mts")
	command.Dir = root
	pathValue := bin
	for _, entry := range extraPath {
		pathValue = entry + string(os.PathListSeparator) + pathValue
	}
	if relay != "" {
		pathValue = pathValue + string(os.PathListSeparator) + os.Getenv("PATH")
	}
	command.Env = overridePATH(os.Environ(), pathValue, logPath, probePath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("transport plugin harness failed: %v\n%s", err, output)
	}
	log, err := os.ReadFile(logPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	probe, err := os.ReadFile(probePath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	relayLog := ""
	if log != nil {
		relayLog = string(log)
	}
	probeLog := ""
	if probe != nil {
		probeLog = string(probe)
	}
	return string(output), relayLog, probeLog
}

// overridePATH replaces PATH with pathValue (appending would leave the
// parent's PATH first since getenv(3) returns the first match).
func overridePATH(parent []string, pathValue, logPath, probePath string) []string {
	env := make([]string, 0, len(parent)+3)
	for _, entry := range parent {
		if strings.HasPrefix(entry, "PATH=") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "PATH="+pathValue, "GENTLE_AI_RELAY_LOG="+logPath, "GENTLE_AI_PROBE_LOG="+probePath)
	return env
}
