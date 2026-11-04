package agent

import "testing"

// These tests derive from the registry and assert the three parallel
// structures (supportedAgents, agentAliases, GetAgent switch) agree.
// Unlike the pinned expectation in TestSupportedAgents, deriving is sound
// here because they exercise the factory and alias map — real behavior,
// not a copy of the list.
func TestAgentAliasesMapToCanonicalNames(t *testing.T) {
	supported := make(map[string]bool, len(supportedAgents))
	for _, name := range supportedAgents {
		if name == "" {
			t.Fatal("empty canonical agent name in supportedAgents")
		}
		supported[name] = true
	}

	for alias, canonical := range agentAliases {
		if !supported[canonical] {
			t.Errorf("alias %q maps to non-canonical name %q", alias, canonical)
		}
	}
}

func TestEveryCanonicalAgentResolvesWithCanonicalName(t *testing.T) {
	for _, name := range supportedAgents {
		a, err := GetAgent(name, "", "", nil)
		if err != nil {
			t.Errorf("GetAgent(%q) failed: %v", name, err)

			continue
		}
		if a.Name() != name {
			t.Errorf("GetAgent(%q).Name() = %q, want canonical name", name, a.Name())
		}
	}
}

func TestEveryAliasResolvesLikeItsCanonicalAgent(t *testing.T) {
	for alias, canonical := range agentAliases {
		a, err := GetAgent(alias, "", "", nil)
		if err != nil {
			t.Errorf("GetAgent(%q) failed: %v", alias, err)

			continue
		}
		want, err := GetAgent(canonical, "", "", nil)
		if err != nil {
			t.Errorf("GetAgent(%q) failed: %v", canonical, err)

			continue
		}
		if a.Name() != want.Name() {
			t.Errorf("alias %q resolves to %q, but canonical %q resolves to %q", alias, a.Name(), canonical, want.Name())
		}
	}
}
