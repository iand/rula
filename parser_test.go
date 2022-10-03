package rula

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	ironOre = &Resource{Name: Name{Singular: "iron_ore"}}
	iron    = &Resource{Name: Name{Singular: "iron"}}
	workers = &Resource{Name: Name{Singular: "workers"}}
)

var ruleTests = []struct {
	spec  string
	rules []*Rule
}{

	{
		spec: `
rule test
	in iron_ore 3
	out iron 1
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: ironOre,
						Quantity: 3,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	in global iron_ore 3
	out location iron 1
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationGlobal,
						Resource: ironOre,
						Quantity: 3,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationLocation,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	if global iron_ore > 6
	in global iron_ore 3
	out self iron 1
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				Preconditions: []ResourceCondition{
					{
						ResourceSpecifier: ResourceSpecifier{
							Relation: RelationGlobal,
							Resource: ironOre,
							Quantity: 6,
						},
						Op: OpGreaterThan,
					},
				},
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationGlobal,
						Resource: ironOre,
						Quantity: 3,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	in iron_ore 3
	out iron 1
	every 5
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 5,
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: ironOre,
						Quantity: 3,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	in iron_ore 3
	out iron 1
	repeat 12
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				Repeat: 12,
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: ironOre,
						Quantity: 3,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	in iron_ore 3
	out iron 3
	onfail test2
end
rule test2
	in iron_ore 1
	out iron 1
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: ironOre,
						Quantity: 3,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 3,
					},
				},
				OnFail: &Rule{
					Name:   "test2",
					Period: 1,
					Inputs: []ResourceSpecifier{
						{
							Relation: RelationSelf,
							Resource: ironOre,
							Quantity: 1,
						},
					},
					Outputs: []ResourceSpecifier{
						{
							Relation: RelationSelf,
							Resource: iron,
							Quantity: 1,
						},
					},
				},
			},
			{
				Name:   "test2",
				Period: 1,
				Inputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: ironOre,
						Quantity: 1,
					},
				},
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	every 0
	out iron 1
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 0,
				Outputs: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	every 0
	set iron 1
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 0,
				Sets: []ResourceSpecifier{
					{
						Relation: RelationSelf,
						Resource: iron,
						Quantity: 1,
					},
				},
			},
		},
	},

	{
		spec: `
rule test
	repeat using workers
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				RepeatFrom: &ResourceSource{
					Relation: RelationSelf,
					Resource: workers,
				},
			},
		},
	},

	{
		spec: `
rule test
	repeat using location workers
end
`,

		rules: []*Rule{
			{
				Name:   "test",
				Period: 1,
				RepeatFrom: &ResourceSource{
					Relation: RelationLocation,
					Resource: workers,
				},
			},
		},
	},
}

func TestRuleParser(t *testing.T) {
	resources := []*Resource{
		ironOre,
		iron,
		workers,
	}

	p := NewRuleParser(resources)

	for _, tc := range ruleTests {
		t.Run("", func(t *testing.T) {
			rules, err := p.Parse(strings.NewReader(tc.spec))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tc.rules, rules); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

var resourceTests = []struct {
	spec      string
	resources []*Resource
	err       bool
}{

	{
		spec: `
resource iron_ore
end
		`,
		resources: []*Resource{
			{
				ID: "iron_ore",
				Name: Name{
					Singular: "iron_ore",
					Plural:   "iron_ore",
				},
			},
		},
	},
}

func TestResourceParser(t *testing.T) {
	p := NewResourceParser()

	for _, tc := range resourceTests {
		t.Run("", func(t *testing.T) {
			resources, err := p.Parse(strings.NewReader(tc.spec))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tc.resources, resources); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
