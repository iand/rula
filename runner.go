package rula

import (
	"fmt"
	"log"
)

type Runner struct {
	ruleStates map[*Rule]RuleState
}

func NewRunner() *Runner {
	return &Runner{
		ruleStates: map[*Rule]RuleState{},
	}
}

func (ru *Runner) Run(rules []*Rule, tick int64, ctx RuleContext) error {
	for _, r := range rules {
		if r.Period == 0 {
			continue
		}

		if err := ru.RunRule(r, tick, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (ru *Runner) RunRule(rule *Rule, tick int64, ctx RuleContext) error {
	state := ru.ruleStates[rule]
	if state.LastRun+int64(rule.Period) > tick {
		return nil
	}

	defer func() {
		state.LastRun = tick
		ru.ruleStates[rule] = state
	}()

	rounds := 1

	if rule.RepeatFrom != nil {
		poolset, ok := ctx.Pools[rule.RepeatFrom.Relation]
		if !ok {
			log.Printf("rule %q failed: no repeat poolset of type %v", rule.Name, rule.RepeatFrom.Relation)
			return nil
		}
		pool := poolset[rule.RepeatFrom.Resource]
		if pool == nil {
			rounds = 0
		} else {
			rounds = pool.Quantity
		}
		log.Printf("rule %q rounds: %d", rule.Name, rounds)

	} else {
		rounds = rule.Repeat + 1
	}

	runOnce := false
	for rounds > 0 {
		ok, err := ru.canRun(rule, ctx)
		if err != nil {
			log.Printf("rule %q failed: %v", rule.Name, err)
			return err
		}
		if !ok {
			if !runOnce && rule.OnFail != nil {
				return ru.RunRule(rule.OnFail, tick, ctx)
			}
			return nil
		}

		runOnce = true
		// Adjust inputs
		for _, in := range rule.Inputs {
			poolset, ok := ctx.Pools[in.Relation]
			if !ok {
				log.Printf("rule %q failed: no input poolset of type %v", rule.Name, in.Relation)
				return nil
			}

			excess := poolset.Remove(in.Resource, in.Quantity)
			if excess > 0 {
				log.Printf("rule %q failed: not enough resource of type %v", rule.Name, in.Resource)
				return nil
			}
		}

		// Adjust outputs
		for _, out := range rule.Outputs {
			poolset, ok := ctx.Pools[out.Relation]
			if !ok {
				// fail, no scope of the required type
				log.Printf("rule %q failed: no output poolset of type %v", rule.Name, out.Relation)
				return nil
			}

			// Any excess is lost
			poolset.Add(out.Resource, out.Quantity)
		}

		// Adjust outputs
		for _, s := range rule.Sets {
			poolset, ok := ctx.Pools[s.Relation]
			if !ok {
				// fail, no scope of the required type
				log.Printf("rule %q failed: no set poolset of type %v", rule.Name, s.Relation)
				return nil
			}

			// Any excess is lost
			poolset.Set(s.Resource, s.Quantity)
		}

		rounds--
	}

	return nil
}

func (ru *Runner) canRun(rule *Rule, ctx RuleContext) (bool, error) {
	for _, c := range rule.Preconditions {
		poolset, ok := ctx.Pools[c.Relation]
		if !ok {
			// fail, no scope of the required type
			return false, fmt.Errorf("rule %q failed: no precondition poolset of type %v", rule.Name, c.Relation)
		}

		q := poolset.Quantity(c.Resource)
		switch c.Op {
		case OpEquals:
			if q != c.Quantity {
				log.Printf("rule %q: cannot run for resource %s, %d != %d", rule.Name, c.Resource, q, c.Quantity)
				return false, nil
			}
		case OpGreaterThan:
			if !(q > c.Quantity) {
				log.Printf("rule %q: cannot run for resource %s, %d not > %d", rule.Name, c.Resource, q, c.Quantity)
				return false, nil
			}
		case OpGreaterThanOrEqual:
			if !(q >= c.Quantity) {
				log.Printf("rule %q: cannot run for resource %s, %d not >= %d", rule.Name, c.Resource, q, c.Quantity)
				return false, nil
			}
		case OpLessThan:
			if !(q < c.Quantity) {
				log.Printf("rule %q: cannot run for resource %s, %d not < %d", rule.Name, c.Resource, q, c.Quantity)
				return false, nil
			}
		case OpLessThanOrEqual:
			if !(q <= c.Quantity) {
				log.Printf("rule %q: cannot run for resource %s, %d not <= %d", rule.Name, c.Resource, q, c.Quantity)
				return false, nil
			}
		default:
			// fail, unknown operation
			return false, fmt.Errorf("rule %q failed: unknown operation %v", rule.Name, c.Op)
		}
	}

	// Check inputs are available
	for _, in := range rule.Inputs {
		poolset, ok := ctx.Pools[in.Relation]
		if !ok {
			// fail, no scope of the required type
			return false, fmt.Errorf("rule %q failed: no input poolset of type %v", rule.Name, in.Relation)
		}

		if in.Quantity > poolset.Quantity(in.Resource) {
			// fail, not enough input
			log.Printf("rule %q failed: not enough of resource %q, got %d wanted %d", rule.Name, in.Resource, poolset.Quantity(in.Resource), in.Quantity)
			return false, nil
		}
	}

	return true, nil
}
