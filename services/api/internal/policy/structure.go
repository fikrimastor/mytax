package policy

import "errors"

// validateStructure checks the rule and group graph. Evaluate and ValidateRelease both use it,
// so a package that cannot be evaluated can never be released.
func validateStructure(p Package) (map[string]Rule, map[string]Group, error) {
	groups := map[string]Group{}
	for _, g := range p.Groups {
		if g.ID == "" || g.LimitSen < 0 {
			return nil, nil, errors.New("invalid group")
		}
		if _, ok := groups[g.ID]; ok {
			return nil, nil, errors.New("duplicate group")
		}
		groups[g.ID] = g
	}
	for _, g := range p.Groups {
		if g.Parent == "" {
			continue
		}
		parent, ok := groups[g.Parent]
		if !ok {
			return nil, nil, errors.New("unknown parent group")
		}
		if g.LimitSen > parent.LimitSen {
			return nil, nil, errors.New("child group limit exceeds parent limit")
		}
		seen := map[string]bool{g.ID: true}
		for id := g.Parent; id != ""; id = groups[id].Parent {
			if seen[id] {
				return nil, nil, errors.New("group parent cycle")
			}
			seen[id] = true
		}
	}
	rules := map[string]Rule{}
	for _, rule := range p.Rules {
		if rule.ID == "" || rule.LimitSen < 0 {
			return nil, nil, errors.New("invalid rule")
		}
		if _, ok := rules[rule.ID]; ok {
			return nil, nil, errors.New("duplicate rule")
		}
		if rule.Group != "" {
			if _, ok := groups[rule.Group]; !ok {
				return nil, nil, errors.New("unknown group")
			}
		}
		if rule.FixedAmount && (rule.LimitSen == 0 || rule.Group != "") {
			return nil, nil, errors.New("fixed-amount rule needs a positive limit and no group")
		}
		rules[rule.ID] = rule
	}
	return rules, groups, nil
}

// groupChain returns id followed by every ancestor group. validateStructure has ruled out cycles.
func groupChain(groups map[string]Group, id string) []string {
	var chain []string
	for ; id != ""; id = groups[id].Parent {
		chain = append(chain, id)
	}
	return chain
}
