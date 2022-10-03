package rula

import (
	"strings"
	"testing"
)

func BenchmarkRunRule(b *testing.B) {
	rule := `
rule test
	in self iron_ore 1
	out self iron_ore 1
end
`

	resources := []*Resource{
		ironOre,
	}

	p := NewRuleParser(resources)

	rules, err := p.Parse(strings.NewReader(rule))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	ctx := RuleContext{
		Pools: map[Relation]PoolSet{
			RelationSelf: {
				ironOre: {Resource: ironOre, Capacity: 1<<63 - 1, Quantity: 1000},
			},
		},
	}

	runner := NewRunner()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.Run(rules, int64(i), ctx)
	}
}
