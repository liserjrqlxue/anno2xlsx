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
	isPP3 = regexp.MustCompile(`PP3`)

	isD = regexp.MustCompile(`D`)

	Tag1AFThreshold = 0.05
)

var keys = []string{
	"ExAC HomoAlt Count",
	"PVFD Homo Count",
	"GnomAD HomoAlt Count",
}

func is0(str string) bool {
	var f, e = strconv.ParseFloat(str, 64)
	if e != nil || f <= 0 {
		return true
	}
	return false
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
				if !is0(item[key]) {
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
				if !is0(item[key]) {
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

	if freq <= Tag1AFThreshold ||
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
	var flag bool
	if specVarDb[item["MutationName"]] {
		flag = true
	}
	if isHgmdDMplus.MatchString(item["HGMD Pred"]) {
		flag = true
	}
	if isClinVarPLP.MatchString(item["ClinVar Significance"]) {
		flag = true
	}
	if flag {
		tagMap["2"] = true
	}
}

func tag3(tagMap map[string]bool, item map[string]string) {
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq <= 0.01 && item["烈性突变"] == "是" {
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
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq > 0.01 {
		return
	}
	if isPP3.MatchString(item["autoRuleName"]) {
		tagMap["4"] = true
		return
	}
	if tag4Func[item["Function"]] && (item["RepeatTag"] == "." || item["RepeatTag"] == "") {
		tagMap["4"] = true
		return
	}
	if item["Function"] != "no-change" && isD.MatchString(item["SpliceAI Pred"]) {
		tagMap["4"] = true
		return
	}
}

func tag5(tagMap map[string]bool, item map[string]string) {
	if item["SecondaryFinding_Var_致病等级"] != "" {
		tagMap["5"] = true
	}
}

func tag6(tagMap map[string]bool, item map[string]string) {
	if item["孕前致病性"] != "" {
		tagMap["6"] = true
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
	tag6(tagMap, item)
	var tags []string
	for _, t := range supportTag {
		if tagMap[t] {
			tags = append(tags, t)
		}
	}
	return strings.Join(tags, ";")
}
