package rula

type Name struct {
	Plural   string
	Singular string
}

func (n *Name) String() string {
	return n.Singular
}

// A Resource is something that is used, consumed or produced
type Resource struct {
	ID   string
	Name Name
}

func (r *Resource) String() string {
	return r.Name.String()
}

// A Pool is a store of resources
type Pool struct {
	Resource *Resource
	Quantity int
	Capacity int
}

type PoolSet map[*Resource]*Pool

func (p PoolSet) SetCapacity(r *Resource, c int) {
	pool, ok := p[r]
	if !ok {
		p[r] = &Pool{Resource: r, Capacity: c}
		return
	}
	pool.Capacity = c
}

func (p PoolSet) AddPool(r *Resource, capacity, quantity int) {
	if r == nil {
		panic("nil resource supplied")
	}
	p[r] = &Pool{Resource: r, Capacity: capacity, Quantity: quantity}
}

func (p PoolSet) Quantity(r *Resource) int {
	if p == nil || r == nil {
		return 0
	}
	pool, ok := p[r]
	if !ok {
		return 0
	}
	return pool.Quantity
}

func (p PoolSet) Capacity(r *Resource) int {
	if p == nil || r == nil {
		return 0
	}
	pool, ok := p[r]
	if !ok {
		return 0
	}
	return pool.Capacity
}

// Add adds quantity q of resource r to the poolset returning the amount that
// could not be added. This will be 0 if there was a pool with sufficient capacity
func (p PoolSet) Add(r *Resource, q int) int {
	if p == nil || r == nil {
		return q
	}
	pool, ok := p[r]
	if !ok {
		return q
	}
	pool.Quantity += q

	if pool.Quantity > pool.Capacity {
		excess := pool.Quantity - pool.Capacity
		pool.Quantity = pool.Capacity
		return excess
	}
	return 0
}

// Set sets the quantity of resource r to be q  returning the amount that
// could not be added. This will be 0 if there was a pool with sufficient capacity
func (p PoolSet) Set(r *Resource, q int) int {
	if p == nil || r == nil {
		return q
	}
	pool, ok := p[r]
	if !ok {
		return q
	}
	pool.Quantity = q

	if pool.Quantity > pool.Capacity {
		excess := pool.Quantity - pool.Capacity
		pool.Quantity = pool.Capacity
		return excess
	}
	return 0
}

// Remove removes quantity q of resource r from the poolset returning the amount that
// could not be removed. This will be 0 if there was a pool with sufficient quantity. This
// method does not split the removal quantity, it will either remove all of q or 0.
func (p PoolSet) Remove(r *Resource, q int) int {
	if p == nil || r == nil {
		return q
	}
	pool, ok := p[r]
	if !ok {
		return q
	}

	if pool.Quantity < q {
		return q
	}

	pool.Quantity -= q

	return 0
}

func NewPoolSet() PoolSet {
	return map[*Resource]*Pool{}
}

// An Agent is something that consumes or produces resources. It could be a person, a building
// or even an entire country.
type Agent struct {
	Name      Name
	Pools     PoolSet
	Rules     []*Rule
	Relations map[Relation]*Agent
}

func NewAgent(name string) *Agent {
	return &Agent{
		Name:      Name{Singular: name},
		Pools:     NewPoolSet(),
		Rules:     []*Rule{},
		Relations: map[Relation]*Agent{},
	}
}

func (a *Agent) PrependRules(rules []*Rule) {
	nrules := append([]*Rule(nil), rules...)
	nrules = append(nrules, a.Rules...)
	a.Rules = nrules
}

func (a *Agent) AppendRules(rules []*Rule) {
	a.Rules = append(a.Rules, rules...)
}

func (a *Agent) SetCapacity(r *Resource, c int) {
	a.Pools.SetCapacity(r, c)
}

func (a *Agent) AddPool(r *Resource, capacity, quantity int) {
	a.Pools.AddPool(r, capacity, quantity)
}

func (a *Agent) AddRelation(r Relation, c *Agent) {
	a.Relations[r] = c
}

func (a *Agent) RuleContext() RuleContext {
	rc := RuleContext{
		Pools: map[Relation]PoolSet{
			RelationSelf: a.Pools,
		},
	}

	for r, ra := range a.Relations {
		rc.Pools[r] = ra.Pools
	}

	return rc
}

// A Global set of pools
type Global struct {
	Pools PoolSet
	Rules []*Rule
}

func NewGlobal(rules []*Rule) *Global {
	return &Global{
		Pools: NewPoolSet(),
		Rules: rules,
	}
}

func (g *Global) SetCapacity(r *Resource, c int) {
	g.Pools.SetCapacity(r, c)
}

// Rules operate on resources
type Rule struct {
	Name          string
	Period        int                 // Number of ticks between occurrences of the rule
	Preconditions []ResourceCondition // conjunctive, all must apply
	Inputs        []ResourceSpecifier
	Outputs       []ResourceSpecifier // Increments or decrements a resource
	Sets          []ResourceSpecifier // Sets a resource quantity to a specific value

	Manual     bool            // true if this rule can only be triggered manually, such as being target of an OnFail
	Repeat     int             // number of times to repeat the rule if possible
	RepeatFrom *ResourceSource // number of times to repeat the rule based on a resource count
	OnFail     *Rule           // a rule to trigger if a precondition fails or an input is missing, only triggered if first run of rule fails, not repeats
}

type ResourceSource struct {
	Relation Relation
	Resource *Resource
}

type ResourceSpecifier struct {
	Relation Relation
	Resource *Resource
	Quantity int
}

type ResourceCondition struct {
	ResourceSpecifier
	Op Op
}

type Op int

const (
	OpEquals             Op = 0
	OpGreaterThan        Op = 1
	OpGreaterThanOrEqual Op = 2
	OpLessThan           Op = 3
	OpLessThanOrEqual    Op = 4
)

type RuleState struct {
	LastRun int64
}

type Relation string

const (
	RelationSelf     Relation = "self"
	RelationGlobal   Relation = "global"
	RelationLocation Relation = "location"
)

type RuleContext struct {
	Pools map[Relation]PoolSet
}
