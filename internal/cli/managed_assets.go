package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/Santiago1809/teicos/internal/assets"
)

const managedAssetProvenanceRefusal = "managed reviewer assets are outdated; run `teicos sync`"

// managedAssetDigest returns a content digest of the managed assets this
// binary embeds. It deliberately does NOT use the review capabilities build
// identity: that identity carries vcs.revision, module version, and ldflags,
// so two binaries built from the same source disagree even when they embed
// byte-identical assets. Comparing build identities would therefore declare
// perfectly current assets outdated after any rebuild, and a test binary
// (which Go does not stamp with VCS settings) could never agree with a
// released one. The digest compares the thing #2685 is actually about.
func managedAssetDigest() (string, error) {
	digest, err := assets.ManagedDigest()
	if err != nil {
		return "", fmt.Errorf("derive managed asset digest: %w", err)
	}
	return digest, nil
}

// managedAssetsBareExecutableClass is the class of executable path that needs
// no quoting on any supported shell.
const managedAssetsBareExecutableClass = `[A-Za-z0-9._-]+`

var managedAssetsBareExecutable = regexp.MustCompile(`^` + managedAssetsBareExecutableClass + `$`)

// managedAssetsExecutablePath resolves the binary that would diagnose a
// managed-asset skew, so its remedy can be anchored to it. Var indirection
// keeps the resolution stubbable in tests.
var managedAssetsExecutablePath = os.Executable

// managedAssetsGOOS selects the quoting form for a path that needs it: the
// invoking binary knows the platform the remedy will run on. Var indirection
// keeps it stubbable in tests.
var managedAssetsGOOS = runtime.GOOS

// managedAssetsExecutableToken renders one executable path as a command-line
// token using one exact platform-specific encoding (#4434): POSIX platforms
// quote with single quotes, which no POSIX shell expands ($ and backticks
// expand even inside double quotes); Windows quotes with double quotes, the
// command syntax cmd.exe accepts, because cmd.exe does not treat single
// quotes as command syntax. A path over the safe bare class stays bare and
// runs as printed in every shell.
func managedAssetsExecutableToken(path string) string {
	if managedAssetsBareExecutable.MatchString(path) {
		return path
	}
	if managedAssetsGOOS == "windows" {
		return "\"" + strings.ReplaceAll(path, "\"", "\\\"") + "\""
	}
	// The standard POSIX idiom for an embedded single quote: close the quoted
	// span, emit an escaped quote, reopen the span.
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}

// managedAssetsExecutableIdentity accepts os.Executable's resolved identity
// only when it is non-empty and single-line. In particular, a CR or LF must not
// reach either quoting renderer, because a remedy is always one line.
func managedAssetsExecutableIdentity(path string) bool {
	return path != "" && !strings.ContainsAny(path, "\r\n")
}

// managedAssetsArgvZeroIdentity is narrower than the os.Executable path: an
// unqualified argv[0] may be a PATH command, not this process's exact identity.
func managedAssetsArgvZeroIdentity(path string) bool {
	return managedAssetsExecutableIdentity(path) && filepath.IsAbs(path)
}

// buildVersionAndCommit resolves the build's own version and commit the same
// way internal/app.ResolveVersion does, without importing internal/app (this
// package must not depend on the top-level command dispatcher). AppVersion is
// the build's stamped version, handed to this package by internal/app at
// startup. It is the same string `teicos --version` prints.
//
// Reading build info alone was wrong: for any stamped build the module
// pseudo-version disagrees with what --version reports. Doctor exists precisely
// to say which build is running, so two answers for one binary is a defect.
func buildVersionAndCommit() (string, string) {
	version, commit := "dev", "unknown"
	if stamped := strings.TrimSpace(AppVersion); stamped != "" && stamped != "dev" {
		version = stamped
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value
			}
		}
	}
	return version, commit
}

// encodeJSON writes one indented JSON payload to the given writer.
func encodeJSON(stdout io.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, payload, "", "  "); err != nil {
		return err
	}
	_, err = stdout.Write(append(indented.Bytes(), '\n'))
	return err
}
