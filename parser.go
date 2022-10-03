package rula

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/iand/loon"
)

/*

Uses loon for Rule file syntax (see github.com/iand/loon)

Line-oriented
Leading and trailing whitespace is ignored
Lines starting with # are comments and ignored

Rule declaration:

  rule <id>
  	declares a new rule

  end
  	ends a rule declaration

Directives:

  in <relation>? <resource> <quantity>
  	declares an input with optional relation, resource name and quantity. the
  	rule will not run if there are not enough resources in
  	the related resource pool

  if <relation>? <resource> <op> <quantity>
  	declares a condition. the rule will only run if the condition
  	holds before any inputs are consumed.
  	op is one of =, >, <, >=, <=

  out <relation>? <resource> <quantity>
  	declares that a resource should be altered by specific quantity (may be negative) upon successful rule evaluation

  set <relation>? <resource> <quantity>
  	declares that a resource should be set to specific quantity upon successful rule evaluation

  every <ticks>
  	number of ticks between invocations of the rule. Set to 0 to
  	prevent this rule running automatically. defaults to 1

  repeat <count>
  	number of times each rule should attempt to run on invocation

  repeat using <relation>? <resource>
  	number of times each rule should attempt to run on invocation, using a resource as the count

  onfail <id>
  	id of a rule to run if preconditions or inputs fail to be satisfied




*/

type RuleParser struct {
	rm map[string]*Resource
}

func NewRuleParser(resources []*Resource) *RuleParser {
	p := &RuleParser{
		rm: make(map[string]*Resource),
	}

	for _, r := range resources {
		p.rm[strings.ToLower(r.Name.Singular)] = r
	}

	return p
}

func (p *RuleParser) Parse(r io.Reader) ([]*Rule, error) {
	type rulespec struct {
		Rule
		onFailRuleName string
	}
	var rulespecs []*rulespec
	ruleIndex := map[string]*rulespec{}

	var rule *rulespec

	pp := loon.NewParser(r)
	for pp.Next() {

		obj := pp.Object()

		if obj.Type != "rule" {
			return nil, fmt.Errorf("unexpected token at line %d (expecting a rule to be started)", obj.Line)
		}

		rule = &rulespec{
			Rule: Rule{
				Name:   obj.Name,
				Period: 1,
			},
		}

		for _, dir := range obj.Directives {
			switch dir.Name {
			case "in", "out", "set":
				if len(dir.Args) != 2 && len(dir.Args) != 3 {
					return nil, fmt.Errorf("malformed resource specifier at line %d: %s %s", dir.Line, dir.Name, dir.ArgText)
				}

				relation := RelationSelf
				if len(dir.Args) == 3 {
					relation = Relation(strings.ToLower(dir.Args[0]))
					dir.Args = dir.Args[1:]
				}

				resname := strings.ToLower(dir.Args[0])

				res, ok := p.rm[resname]
				if !ok {
					return nil, fmt.Errorf("unknown resource at line %d: %q", dir.Line, resname)
				}

				quantity, err := strconv.Atoi(dir.Args[1])
				if err != nil {
					return nil, fmt.Errorf("invalid quantity at line %d: %q", dir.Line, err)
				}

				specifier := ResourceSpecifier{
					Relation: relation,
					Resource: res,
					Quantity: quantity,
				}

				if dir.Name == "in" {
					rule.Inputs = append(rule.Inputs, specifier)
				} else if dir.Name == "set" {
					rule.Sets = append(rule.Sets, specifier)
				} else {
					rule.Outputs = append(rule.Outputs, specifier)
				}

			case "if":
				if len(dir.Args) != 3 && len(dir.Args) != 4 {
					return nil, fmt.Errorf("malformed resource condition at line %d: %s %s", dir.Line, dir.Name, dir.ArgText)
				}

				relation := RelationSelf
				if len(dir.Args) == 4 {
					relation = Relation(strings.ToLower(dir.Args[0]))
					dir.Args = dir.Args[1:]
				}

				resname := strings.ToLower(dir.Args[0])

				res, ok := p.rm[resname]
				if !ok {
					return nil, fmt.Errorf("unknown resource at line %d: %q", dir.Line, resname)
				}

				var op Op
				switch dir.Args[1] {
				case "=":
					op = OpEquals
				case ">":
					op = OpGreaterThan
				case "<":
					op = OpLessThan
				case ">=":
					op = OpGreaterThanOrEqual
				case "<=":
					op = OpLessThanOrEqual
				default:
					return nil, fmt.Errorf("unknown operator at line %d: %s", dir.Line, dir.Args[2])
				}

				quantity, err := strconv.Atoi(dir.Args[2])
				if err != nil {
					return nil, fmt.Errorf("invalid quantity at line %d: %v", dir.Line, err)
				}

				cond := ResourceCondition{
					ResourceSpecifier: ResourceSpecifier{
						Relation: relation,
						Resource: res,
						Quantity: quantity,
					},
					Op: op,
				}

				rule.Preconditions = append(rule.Preconditions, cond)
			case "every":
				if len(dir.Args) != 1 {
					return nil, fmt.Errorf("malformed every directive at line %d: %s %s", dir.Line, dir.Name, dir.ArgText)
				}
				period, err := strconv.Atoi(dir.Args[0])
				if err != nil {
					return nil, fmt.Errorf("invalid period at line %d: %v", dir.Line, err)
				}
				rule.Period = period
			case "repeat":
				if len(dir.Args) == 0 || len(dir.Args) > 3 {
					return nil, fmt.Errorf("malformed repeat directive at line %d: %s %s", dir.Line, dir.Name, dir.ArgText)
				}

				if len(dir.Args) == 1 {
					count, err := strconv.Atoi(dir.Args[len(dir.Args)-1])
					if err != nil {
						return nil, fmt.Errorf("invalid repeat at line %d: %v", dir.Line, err)
					}

					rule.Repeat = count
				} else if dir.Args[0] == "using" {
					dir.Args = dir.Args[1:]

					// must be repeat using <relation>? <resource>
					relation := RelationSelf
					if len(dir.Args) == 2 {
						relation = Relation(strings.ToLower(dir.Args[0]))
						dir.Args = dir.Args[1:]
					}

					resname := strings.ToLower(dir.Args[0])
					res, ok := p.rm[resname]
					if !ok {
						return nil, fmt.Errorf("unknown resource at line %d: %q", obj.Line, resname)
					}

					rule.RepeatFrom = &ResourceSource{
						Relation: relation,
						Resource: res,
					}

				} else {
					return nil, fmt.Errorf("malformed repeat at line %d: %s %s", dir.Line, dir.Name, dir.ArgText)
				}

			case "onfail":
				if len(dir.Args) != 1 {
					return nil, fmt.Errorf("malformed onfail directive at line %d: %s %s", dir.Line, dir.Name, dir.ArgText)
				}
				rule.onFailRuleName = dir.Args[0]
			default:
				return nil, fmt.Errorf("unknown directive at line %d: %s", dir.Line, dir.Name)
			}
		}

		rulespecs = append(rulespecs, rule)
		ruleIndex[rule.Name] = rule
	}

	if pp.Err() != nil {
		return nil, pp.Err()
	}

	var rules []*Rule
	for _, r := range rulespecs {
		if r.onFailRuleName != "" {
			onFail, exists := ruleIndex[r.onFailRuleName]
			if !exists {
				return nil, fmt.Errorf("%s: unknown onfail rule: %q", r.Name, r.onFailRuleName)
			}
			r.Rule.OnFail = &onFail.Rule
		}
		rules = append(rules, &r.Rule)
	}

	return rules, nil
}

type ResourceParser struct{}

func NewResourceParser() *ResourceParser {
	p := &ResourceParser{}

	return p
}

func (p *ResourceParser) Parse(r io.Reader) ([]*Resource, error) {
	var resources []*Resource

	var res *Resource

	pp := loon.NewParser(r)
	for pp.Next() {
		obj := pp.Object()

		if obj.Type != "resource" {
			return nil, fmt.Errorf("unexpected token at line %d (expecting a resource to be started)", obj.Line)
		}

		res = &Resource{
			ID: strings.TrimSpace(obj.Name),
			Name: Name{
				Singular: strings.TrimSpace(obj.Name),
				Plural:   strings.TrimSpace(obj.Name),
			},
		}
		for _, dir := range obj.Directives {
			switch dir.Name {
			case "singular":
				res.Name.Singular = dir.ArgText
			case "plural":
				res.Name.Plural = dir.ArgText
			default:
				return nil, fmt.Errorf("unknown directive at line %d: %s", dir.Line, dir.Name)
			}
		}

		resources = append(resources, res)

	}

	if pp.Err() != nil {
		return nil, pp.Err()
	}
	return resources, nil
}
