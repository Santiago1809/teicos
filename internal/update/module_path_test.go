package update

import "testing"

// TestModulePathForVersion pins the helper's contract for issue #4687:
// derive the Go module /vN suffix from the version being installed, not
// from the running binary's own major. The seam runningGoMajor is pinned to
// 4 (the current binary major) so every unparseable version falls back to /v4.
func TestModulePathForVersion(t *testing.T) {
	origRunning := runningGoMajor
	t.Cleanup(func() { runningGoMajor = origRunning })
	runningGoMajor = func() int { return 4 }

	const base = "github.com/Santiago1809/teicos/cmd/teicos"
	const repo = "teicos"

	tests := []struct {
		name    string
		base    string
		repo    string
		version string
		want    string
	}{
		{
			name:    "v3.4.0 inserts /v3 before /cmd subpath",
			base:    base,
			repo:    repo,
			version: "v3.4.0",
			want:    "github.com/Santiago1809/teicos/v3/cmd/teicos",
		},
		{
			name:    "v2.0.0 inserts /v2 before /cmd subpath",
			base:    base,
			repo:    repo,
			version: "v2.0.0",
			want:    "github.com/Santiago1809/teicos/v2/cmd/teicos",
		},
		{
			name:    "v4.0.0-rc.1 inserts /v4 before /cmd subpath",
			base:    base,
			repo:    repo,
			version: "v4.0.0-rc.1",
			want:    "github.com/Santiago1809/teicos/v4/cmd/teicos",
		},
		{
			name:    "10.0.0 multi-digit inserts /v10 before /cmd subpath",
			base:    base,
			repo:    repo,
			version: "10.0.0",
			want:    "github.com/Santiago1809/teicos/v10/cmd/teicos",
		},
		{
			name:    "v1.9.0 stays unsuffixed (major 1)",
			base:    base,
			repo:    repo,
			version: "v1.9.0",
			want:    base,
		},
		{
			name:    "v0 stays unsuffixed (major 0)",
			base:    base,
			repo:    repo,
			version: "v0",
			want:    base,
		},
		{
			name:    "unparseable 'latest' falls back to running major /v4",
			base:    base,
			repo:    repo,
			version: "latest",
			want:    "github.com/Santiago1809/teicos/v4/cmd/teicos",
		},
		{
			name:    "unparseable 'main@abc' falls back to running major /v4",
			base:    base,
			repo:    repo,
			version: "main@abc",
			want:    "github.com/Santiago1809/teicos/v4/cmd/teicos",
		},
		{
			name:    "empty version falls back to running major /v4",
			base:    base,
			repo:    repo,
			version: "",
			want:    "github.com/Santiago1809/teicos/v4/cmd/teicos",
		},
		{
			name:    "base without subpath appends /vN",
			base:    "github.com/Santiago1809/teicos",
			repo:    repo,
			version: "v3.0.1",
			want:    "github.com/Santiago1809/teicos/v3",
		},
		{
			name:    "leading v stripped when parsing",
			base:    base,
			repo:    repo,
			version: "V3.0.0",
			want:    "github.com/Santiago1809/teicos/v3/cmd/teicos",
		},
		{
			name:    "repo substring inside longer name is not matched as a segment",
			base:    "github.com/Gentleman-Programming/gentleman-guardian-angel/cmd/gga",
			repo:    "guardian-angel",
			version: "v2.0.0",
			want:    "github.com/Gentleman-Programming/gentleman-guardian-angel/cmd/gga/v2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ModulePathForVersion(tc.base, tc.repo, tc.version)
			if got != tc.want {
				t.Errorf("ModulePathForVersion(%q, %q, %q) = %q, want %q", tc.base, tc.repo, tc.version, got, tc.want)
			}
		})
	}
}

// TestParseMajorVersion covers the version-parsing edge cases independently
// of the suffix-insertion logic. Returns -1 for unparseable input.
func TestParseMajorVersion(t *testing.T) {
	tests := []struct {
		version string
		want    int
	}{
		{"3.4.0", 3},
		{"v3.4.0", 3},
		{"V3.4.0", 3},
		{"v4.0.0-rc.1", 4},
		{"10.0.0", 10},
		{"v1.9.0", 1},
		{"v0", 0},
		{"latest", -1},
		{"main@abc", -1},
		{"", -1},
		{"v", -1},
		{"vabc", -1},
		{"  v3.0.0  ", 3},
	}
	for _, tc := range tests {
		got := parseMajorVersion(tc.version)
		if got != tc.want {
			t.Errorf("parseMajorVersion(%q) = %d, want %d", tc.version, got, tc.want)
		}
	}
}

// TestParseMajorFromModulePath covers the running-binary's module-path
// parser. A suffixed path .../teicos/v4 → 4; an unsuffixed path returns 0.
func TestParseMajorFromModulePath(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{"github.com/Santiago1809/teicos/v4", 4},
		{"github.com/Santiago1809/teicos/v4/cmd/teicos", 4},
		{"github.com/Santiago1809/teicos/v3", 3},
		{"github.com/Santiago1809/teicos/v10/cmd/x", 10},
		{"github.com/Santiago1809/teicos", 0},
		{"github.com/Santiago1809/teicos/cmd/x", 0},
		{"", 0},
		{"/v3", 3},
		{"/v", 0},
		{"/vN", 0},
	}
	for _, tc := range tests {
		got := parseMajorFromModulePath(tc.path)
		if got != tc.want {
			t.Errorf("parseMajorFromModulePath(%q) = %d, want %d", tc.path, got, tc.want)
		}
	}
}
