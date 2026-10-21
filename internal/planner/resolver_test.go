package planner

import (
	"reflect"
	"testing"

	"github.com/Santiago1809/teicos/internal/model"
)

// TestResolverSkillsAloneResolvesWithoutSDDOrEngram guards #3554: selecting
// only the Skills component must not force-install SDD (and transitively
// Engram). Standalone skill files have no hard dependency on SDD; the
// SDD-authored skills are installed by the SDD component itself.
//
// This test replaces the old TestResolverAddsMissingDependenciesInOrder,
// which encoded the previous (buggy) expansion of Skills into
// [Engram, SDD, Skills].
func TestResolverSkillsAloneResolvesWithoutSDDOrEngram(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
		Components: []model.ComponentID{model.ComponentSkills},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !reflect.DeepEqual(plan.Agents, []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}) {
		t.Fatalf("Resolve() agents = %v", plan.Agents)
	}

	if !reflect.DeepEqual(plan.OrderedComponents, []model.ComponentID{model.ComponentSkills}) {
		t.Fatalf("Resolve() ordered components = %v, want [skills] only", plan.OrderedComponents)
	}

	if len(plan.AddedDependencies) != 0 {
		t.Fatalf("Resolve() added dependencies = %v, want none", plan.AddedDependencies)
	}
}

func TestResolverEngramOnlyDoesNotForcePersona(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Components: []model.ComponentID{model.ComponentEngram},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !reflect.DeepEqual(plan.OrderedComponents, []model.ComponentID{model.ComponentEngram}) {
		t.Fatalf("Resolve() ordered components = %v", plan.OrderedComponents)
	}

	if len(plan.AddedDependencies) != 0 {
		t.Fatalf("Resolve() added dependencies = %v, want none", plan.AddedDependencies)
	}
}

func TestResolverSDDOnlyDoesNotForcePersona(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Components: []model.ComponentID{model.ComponentSkills},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	for _, dep := range plan.AddedDependencies {
		if dep == model.ComponentPersona {
			t.Fatalf("SDD-only selection should NOT force Persona, got AddedDependencies=%v", plan.AddedDependencies)
		}
	}
}

func TestResolverPersonaAndEngramWithoutSDD(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Components: []model.ComponentID{model.ComponentPersona, model.ComponentEngram},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !reflect.DeepEqual(plan.OrderedComponents, []model.ComponentID{model.ComponentPersona, model.ComponentEngram}) {
		t.Fatalf("Resolve() ordered components = %v, want [persona, engram]", plan.OrderedComponents)
	}

	if len(plan.AddedDependencies) != 0 {
		t.Fatalf("Resolve() added dependencies = %v, want none", plan.AddedDependencies)
	}
}

func TestResolverExcludesUnsupportedAgents(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Agents: []model.AgentID{model.AgentClaudeCode, model.AgentCursor, model.AgentID("unknown-agent")},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !reflect.DeepEqual(plan.Agents, []model.AgentID{model.AgentClaudeCode, model.AgentCursor}) {
		t.Fatalf("Resolve() agents = %v", plan.Agents)
	}

	if !reflect.DeepEqual(plan.UnsupportedAgents, []model.AgentID{model.AgentID("unknown-agent")}) {
		t.Fatalf("Resolve() unsupported agents = %v", plan.UnsupportedAgents)
	}
}
