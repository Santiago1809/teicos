package telemetry

import "testing"

// testEndpoint stands in for a configured collector: with no default endpoint,
// telemetry is disabled unless one is set, so most tests need one.
const testEndpoint = "https://collector.example.invalid/v1/events"

// envMap returns a Getenv over values. EndpointEnvVar defaults to testEndpoint
// unless values sets it explicitly (an empty string means "no endpoint").
func envMap(values map[string]string) Getenv {
	return func(key string) string {
		if value, ok := values[key]; ok {
			return value
		}
		if key == EndpointEnvVar {
			return testEndpoint
		}
		return ""
	}
}

func TestDecidePrecedence(t *testing.T) {
	enabledState := State{Enabled: true}
	disabledState := State{Enabled: false}

	cases := []struct {
		name       string
		env        map[string]string
		state      State
		wantSource Source
	}{
		{"do not track wins over everything", map[string]string{"DO_NOT_TRACK": "1", "GENTLE_AI_TELEMETRY": "1", "CI": "false"}, enabledState, SourceDoNotTrack},
		{"env opt-out wins over CI and state", map[string]string{"GENTLE_AI_TELEMETRY": "0", "CI": "false"}, enabledState, SourceEnvOptOut},
		{"CI true wins over state", map[string]string{"CI": "true"}, enabledState, SourceCI},
		{"CI is case-insensitive", map[string]string{"CI": "True"}, enabledState, SourceCI},
		{"state disable is the last resort", map[string]string{}, disabledState, SourceStateDisable},
		{"default enabled when nothing opts out", map[string]string{"DO_NOT_TRACK": "0", "GENTLE_AI_TELEMETRY": "1", "CI": "false"}, enabledState, SourceDefault},
		{"DO_NOT_TRACK opts out for any non-off value", map[string]string{"DO_NOT_TRACK": "yes"}, enabledState, SourceDoNotTrack},
		{"DO_NOT_TRACK true opts out", map[string]string{"DO_NOT_TRACK": "true"}, enabledState, SourceDoNotTrack},
		{"DO_NOT_TRACK 0 stays default", map[string]string{"DO_NOT_TRACK": "0"}, enabledState, SourceDefault},
		{"DO_NOT_TRACK false stays default", map[string]string{"DO_NOT_TRACK": "false"}, enabledState, SourceDefault},
		{"DO_NOT_TRACK empty stays default", map[string]string{"DO_NOT_TRACK": ""}, enabledState, SourceDefault},
		{"GENTLE_AI_TELEMETRY requires exactly 0", map[string]string{"GENTLE_AI_TELEMETRY": "false"}, enabledState, SourceDefault},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			decision := Decide(envMap(c.env), c.state)
			if decision.Source != c.wantSource {
				t.Fatalf("source = %q, want %q", decision.Source, c.wantSource)
			}
			wantEnabled := c.wantSource == SourceDefault
			if decision.Enabled != wantEnabled {
				t.Fatalf("enabled = %v, want %v", decision.Enabled, wantEnabled)
			}
		})
	}
}

func TestDecideIsDisabledWithoutAnEndpoint(t *testing.T) {
	for _, env := range []map[string]string{{EndpointEnvVar: ""}, {EndpointEnvVar: "  "}} {
		decision := Decide(envMap(env), State{Enabled: true})
		if decision.Enabled || decision.Source != SourceNoEndpoint {
			t.Fatalf("Decide() = %+v, want disabled by %q", decision, SourceNoEndpoint)
		}
	}
}

func TestEndpointDefaultAndOverride(t *testing.T) {
	if DefaultEndpoint != "" {
		t.Fatalf("DefaultEndpoint = %q, want empty: nothing is sent unless an endpoint is configured", DefaultEndpoint)
	}
	if got := Endpoint(envMap(map[string]string{EndpointEnvVar: ""})); got != DefaultEndpoint {
		t.Fatalf("Endpoint() = %q, want default %q", got, DefaultEndpoint)
	}
	custom := "https://example.invalid/v1/events"
	if got := Endpoint(envMap(map[string]string{EndpointEnvVar: custom})); got != custom {
		t.Fatalf("Endpoint() = %q, want override %q", got, custom)
	}
	if got := Endpoint(envMap(map[string]string{EndpointEnvVar: "  "})); got != DefaultEndpoint {
		t.Fatalf("Endpoint() with blank override = %q, want default", got)
	}
}

func TestDecideTreatsCIStyleFlagsAsSet(t *testing.T) {
	cases := map[string]map[string]string{
		"CI=1":                {"CI": "1"},
		"CI=yes":              {"CI": "yes"},
		"GITHUB_ACTIONS=true": {"GITHUB_ACTIONS": "true"},
		"CI=true upper":       {"CI": "TRUE"},
	}
	for name, env := range cases {
		if d := Decide(envMap(env), State{Enabled: true}); d.Enabled || d.Source != SourceCI {
			t.Fatalf("%s: decision = %+v, want disabled by CI", name, d)
		}
	}
	for _, v := range []string{"", "0", "false"} {
		if d := Decide(envMap(map[string]string{"CI": v, "GITHUB_ACTIONS": v}), State{Enabled: true}); !d.Enabled {
			t.Fatalf("CI=%q: decision = %+v, want enabled", v, d)
		}
	}
}
