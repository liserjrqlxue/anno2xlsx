package anno

import "strings"

var AutoRuleKey = []string{
	"PVS1", "PS1", "PS4", "PM1", "PM2", "PM4", "PM5", "PP2", "PP3", "PP5", "BA1",
	"BS1", "BS2", "BP1", "BP3", "BP4", "BP6", "BP7",
}

var ManualRuleKey = []string{
	"PVS1", "PS1", "PM5", "PS2", "PS3", "PM3", "PM6", "PP1",
	"PP4", "BS3", "BS4", "BP2", "BP5",
}

func UpdateAutoRule(item map[string]string) {
	var autoRuleName, autoIsChecked []string
	for _, key := range AutoRuleKey {
		if item[key] != "" && item[key] != "0" {
			autoRuleName = append(autoRuleName, key)
			autoIsChecked = append(autoIsChecked, "1")
		}
	}
	item["autoRuleName"] = strings.Join(autoRuleName, "\n")
	item["autoIsChecked"] = strings.Join(autoIsChecked, "\n")
}

func UpdateManualRule(item map[string]string) {
	var manualRuleName, manualExplaination []string
	for _, key := range ManualRuleKey {
		if item[key] != "" && item[key] != "0" {
			manualRuleName = append(manualRuleName, key)
			manualExplaination = append(manualExplaination, item[key])
		}
	}
	item["manualRuleName"] = strings.Join(manualRuleName, "\n")
	item["manualExplaination"] = strings.Join(manualExplaination, "\n")
}
