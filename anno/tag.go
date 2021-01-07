package anno

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	//isBLB        = regexp.MustCompile(`B|LB`)
	isClinVarPLP = regexp.MustCompile(`Pathogenic|Likely_pathogenic`)
	//isHgmdDM     = regexp.MustCompile(`DM$|DM\|`)
	isHgmdDMplus = regexp.MustCompile(`DM`)
	//isHgmdDMQ= regexp.MustCompile(`DM\?`)
	isPP3  = regexp.MustCompile(`PP3`)
	isZero = regexp.MustCompile(`^0$|^\.$|^$`)
)

var keys = []string{
	"ExAC HomoAlt Count",
	"PVFD Homo Count",
	"GnomAD HomoAlt Count",
	"1000G EAS AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC EAS AF",
	"ExAC AF",
	"GnomAD EAS AF",
	"GnomAD AF",
	"PVFD AF",
	"Panel AlleleFreq",
}

func tag1Trio(tagMap map[string]bool, item map[string]string) {
	if item["遗传相符-经典trio"] == "相符" {
		tagMap["T1"] = true
	}
	if item["遗传相符-非经典trio"] == "相符" {
		inherit := item["ModeInheritance"]
		if isAR.MatchString(inherit) || isXL.MatchString(inherit) || isYL.MatchString(inherit) {
			tagMap["1"] = true
		} else if isAD.MatchString(inherit) {
			var flag = true
			for _, key := range keys {
				if !isZero.MatchString(item[key]) {
					flag = false
				}
			}
			if flag {
				tagMap["1"] = true
			}
		}
	}
}

func tag1Single(tagMap map[string]bool, item map[string]string) {
	if item["遗传相符"] == "相符" {
		inherit := item["ModeInheritance"]
		if isAR.MatchString(inherit) || isXL.MatchString(inherit) || isYL.MatchString(inherit) {
			tagMap["1"] = true
		} else if isAD.MatchString(inherit) {
			var flag = true
			for _, key := range keys {
				if !isZero.MatchString(item[key]) {
					flag = false
				}
			}
			if flag {
				tagMap["1"] = true
			}
		}
	}
}

func tag1(tagMap map[string]bool, item map[string]string, specVarDb map[string]bool, isTrio, isTrio2 bool) {
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}

	if freq <= 0.01 ||
		specVarDb[item["MutationName"]] ||
		isHgmdDMplus.MatchString(item["HGMD Pred"]) ||
		isClinVarPLP.MatchString(item["ClinVar Significance"]) {
	} else {
		return
	}

	if isTrio || isTrio2 {
		tag1Trio(tagMap, item)
	} else {
		tag1Single(tagMap, item)
	}
}

func tag2(tagMap map[string]bool, item map[string]string, specVarDb map[string]bool) {
	var flag1 bool
	if specVarDb[item["MutationName"]] {
		flag1 = true
	}
	if isHgmdDMplus.MatchString(item["HGMD Pred"]) {
		flag1 = true
	}
	if isClinVarPLP.MatchString(item["ClinVar Significance"]) {
		flag1 = true
	}
	if flag1 {
		tagMap["2"] = true
	}
}

func tag3(tagMap map[string]bool, item map[string]string) {
	var flag1, flag2 bool
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq <= 0.01 {
		flag1 = true
	}
	if item["烈性突变"] == "是" {
		flag2 = true
	}
	if flag1 && flag2 {
		tagMap["3"] = true
	}
}

var tag4Func = map[string]bool{
	"stop-loss": true,
	"cds-del":   true,
	"cds-indel": true,
	"cds-ins":   true,
}

func tag4(tagMap map[string]bool, item map[string]string) {
	var flag1, flag2 bool
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq <= 0.01 {
		flag1 = true
	}
	if isPP3.MatchString(item["autoRuleName"]) {
		flag2 = true
	}
	if tag4Func[item["Function"]] && (item["RepeatTag"] == "." || item["RepeatTag"] == "") {
		flag2 = true
	}

	if flag1 && flag2 {
		tagMap["4"] = true
	}
}

func tag5(tagMap map[string]bool, item map[string]string) {
	if item["SecondaryFinding_Var_致病等级"] != "" {
		tagMap["5"] = true
	}
}

//UpdateTags get Tags of item
func UpdateTags(item map[string]string, specVarDb map[string]bool, isTrio, isTrio2 bool) string {
	var tagMap = make(map[string]bool)
	tag1(tagMap, item, specVarDb, isTrio, isTrio2)
	tag2(tagMap, item, specVarDb)
	tag3(tagMap, item)
	tag4(tagMap, item)
	tag5(tagMap, item)
	var tags []string
	for _, t := range supportTag {
		if tagMap[t] {
			tags = append(tags, t)
		}
	}
	return strings.Join(tags, ";")
}
