package anno

import (
	"strconv"
	"strings"
)

var autoRuleScores = map[string]int{
	"PVS1": 8, "PS1": 4, "PS4": 4, "PM1": 2, "PM2": 2, "PM4": 2, "PM5": 2, "PP2": 1, "PP3": 1, "PP5": 1,
	"BA1": -8, "BS1": -4, "BS2": -4, "BP1": -1, "BP3": -1, "BP4": -1, "BP6": -1, "BP7": -1,
}
var autoRuleKey = []string{
	"PVS1", "PS1", "PS4", "PM1", "PM2", "PM4", "PM5", "PP2", "PP3", "PP5", "BA1",
	"BS1", "BS2", "BP1", "BP3", "BP4", "BP6", "BP7",
}

var manualRuleKey = []string{
	"PVS1", "PS1", "PM5", "PS2", "PS3", "PM3", "PM6", "PP1",
	"PP4", "BS3", "BS4", "BP2", "BP5",
}

//UpdateAutoRule update auto rules of acmg2015
func UpdateAutoRule(item map[string]string) {
	var autoRuleScroe int
	var autoRuleName, autoIsChecked []string
	if item["AutoPVS1 Adjusted Strength"] != "" {
		item["PVS1"] = ""
	}
	switch item["AutoPVS1 Adjusted Strength"] {
	case "VeryStrong":
		autoRuleName = append(autoRuleName, "PVS1")
		autoIsChecked = append(autoIsChecked, "1")
		autoRuleScroe += 8
	case "Strong":
		autoRuleName = append(autoRuleName, "PVS1_Strong")
		autoIsChecked = append(autoIsChecked, "1")
		autoRuleScroe += 4
	case "Moderate":
		autoRuleName = append(autoRuleName, "PVS1_Moderate")
		autoIsChecked = append(autoIsChecked, "1")
		autoRuleScroe += 2
	case "Supporting":
		autoRuleName = append(autoRuleName, "PVS1_Supporting")
		autoIsChecked = append(autoIsChecked, "1")
		autoRuleScroe += 1
	}
	for _, key := range autoRuleKey {
		if item[key] != "" && item[key] != "0" {
			autoRuleName = append(autoRuleName, key)
			autoIsChecked = append(autoIsChecked, "1")
			autoRuleScroe += autoRuleScores[key]
		}
	}
	item["autoRuleName"] = strings.Join(autoRuleName, "\n")
	item["autoIsChecked"] = strings.Join(autoIsChecked, "\n")
	item["autoRuleScore"] = strconv.Itoa(autoRuleScroe)
}

//UpdateManualRule update manualRuleName and manualExplaination
func UpdateManualRule(item map[string]string) {
	var manualRuleName, manualExplaination []string
	for _, key := range manualRuleKey {
		if item[key] != "" && item[key] != "0" {
			manualRuleName = append(manualRuleName, key)
			manualExplaination = append(manualExplaination, item[key])
		}
	}
	item["manualRuleName"] = strings.Join(manualRuleName, "\n")
	item["manualExplaination"] = strings.Join(manualExplaination, "\n")
}
