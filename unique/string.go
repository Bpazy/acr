package unique

func Strings(rules []string) []string {
	ruleSet := map[string]bool{}
	var rules2 []string
	for _, rule := range rules {
		if _, ok := ruleSet[rule]; ok {
			continue
		}
		ruleSet[rule] = true
		rules2 = append(rules2, rule)
	}
	return rules2
}
