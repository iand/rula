# SCRIPTING

# 2021-08-31

Some thoughts on types of resources:

 - Currently they are linear counts with an upper bound (capacity).
 - Could add other types of resources with different aggregation functions. 
 - For example, could have an s-curve resource with diminishing returns as we get closer to upper and lower bounds.
 - Booleans can already be modeled with 0/1



Rules can be used to set trigger flags:

	# Determine if agent is hungry
	rule is_hungry
	if food = 0
	set self hungry 1

	# Increase unrest while hungry
    rule unrest
    if self hungry = 1
    out self unrest 1









2020-06-16
----------

Sketch of scripting system


See https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md

NOTE: every script file must have a namespace

Some definitions:

 * **modifier** is a quantity that ranges from -100 to +100(?) and is interpreted as a % (100+x/100)

 * **trigger** is an effect that is applied whenever a condition is met and remains until the trigger conditions fail to be met

 * **rule** is an effect that is applied repeatedly whenever a condition is met 

 * **event** is an effect that is triggered and presents one or more choices to the player

 * **decision** is an effect that is triggered with a single action that the player may choose to enact



* on <action_id> - script will fire when action fires. Actions could be things like a caravan arriving



## Structure 

Each rule/event/decision/trigger has the same basic structure. They mostly differ in their mode of activation, i.e. whether they
are automatic, user activated or prompt the user for action.

 * **prerequisites** - conditions that must be met before the rule becomes available. Decisions are hidden from the user until the prerequisites are met
 * **conditions** - conditions that must be met for the rule to execute
 * **effects** - the effects that are applied when the rule is executed






## Conditions

  if <relation>? <resource> <op> <quantity>
    declares a condition. the rule will only run if the condition
    holds before any inputs are consumed.
    op is one of =, >, <, >=, <=

  hasflag <relation>? <flag>
    declares a condition. the rule will only run if the condition
    holds before any inputs are consumed.
    op is one of =, >, <, >=, <=


## Rules

  rule <id>
  	declares a new rule

  end
  	ends a rule declaration

  in <relation>? <resource> <quantity>
  	declares an input with optional relation, resource name and quantity. the
  	rule will not run if there are not enough resources in
  	the related resource pool

  out <relation>? <resource> <quantity>
  	declares an output with optional relation, resource name and quantity

  if <relation>? <resource> <op> <quantity>
  	declares a condition. the rule will only run if the condition
  	holds before any inputs are consumed.
  	op is one of =, >, <, >=, <=

  every <ticks>
  	number of ticks between invocations of the rule. Set to 0 to
  	prevent this rule running automatically. defaults to 1

  repeat <count>
  	number of times each rule should attempt to run on invocation

  repeat using <relation>? <resource>
  	number of times each rule should attempt to run on invocation, using a resource as the count

  onfail <id>
  	id of a rule to run if preconditions or inputs fail to be satisfied


Events

  event <id>
  	declares a new event
	
  end
  	ends an event declaration


  TODO: conditions

  mtth <ticks>
  	mean time to happen in ticks


  effect <relation> <modifier> <duration>
    apply the given modifier for the duration. Duration is in ticks or the keyword 'forever'



Modifiers





