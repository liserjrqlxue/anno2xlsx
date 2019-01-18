package main

// map[string]string update
import (
	"github.com/liserjrqlxue/simple-util"
	"regexp"
	//"github.com/liserjrqlxue/acmg2015"
	"strconv"
	"strings"
)

// Tier1 >1
// LoF 3
var FuncInfo = map[string]int{
	"splice-3":     3,
	"splice-5":     3,
	"inti-loss":    3,
	"alt-start":    3,
	"frameshift":   3,
	"nonsense":     3,
	"stop-gain":    3,
	"span":         3,
	"missense":     2,
	"cds-del":      2,
	"cds-indel":    2,
	"cds-ins":      2,
	"splice-10":    2,
	"splice+10":    2,
	"coding-synon": 1,
	"splice-20":    1,
	"splice+20":    1,
}

var long2short = map[string]string{
	"Pathogenic":             "P",
	"Likely Pathogenic":      "LP",
	"Uncertain Significance": "VUS",
	"Likely Benign":          "LB",
	"Benign":                 "B",
	"P":                      "P",
	"LP":                     "LP",
	"VUS":                    "VUS",
	"LB":                     "LB",
	"B":                      "B",
}

// regexp
var (
	isHgmd    = regexp.MustCompile("DM")
	isClinvar = regexp.MustCompile("Pathogenic|Likely_pathogenic")
	indexReg  = regexp.MustCompile(`\d+\.\s+`)
	//newlineReg = regexp.MustCompile(`\n+`)
	isDenovo  = regexp.MustCompile(`NA;NA$`)
	noProband = regexp.MustCompile(`^NA`)
)

func updateSnv(dataHash map[string]string) {

	// pHGVS= pHGVS1+"|"+pHGVS3
	dataHash["pHGVS"] = dataHash["pHGVS1"] + " | " + dataHash["pHGVS3"]

	score, err := strconv.ParseFloat(dataHash["dbscSNV_ADA_SCORE"], 32)
	if err != nil {
		dataHash["dbscSNV_ADA_pred"] = dataHash["dbscSNV_ADA_SCORE"]
	} else {
		if score >= 0.6 {
			dataHash["dbscSNV_ADA_pred"] = "D"
		} else {
			dataHash["dbscSNV_ADA_pred"] = "P"
		}
	}
	score, err = strconv.ParseFloat(dataHash["dbscSNV_RF_SCORE"], 32)
	if err != nil {
		dataHash["dbscSNV_RF_pred"] = dataHash["dbscSNV_RF_SCORE"]
	} else {
		if score >= 0.6 {
			dataHash["dbscSNV_RF_pred"] = "D"
		} else {
			dataHash["dbscSNV_RF_pred"] = "P"
		}
	}

	score, err = strconv.ParseFloat(dataHash["GERP++_RS"], 32)
	if err != nil {
		dataHash["GERP++_RS_pred"] = dataHash["GERP++_RS"]
	} else {
		if score >= 2 {
			dataHash["GERP++_RS_pred"] = "D"
		} else {
			dataHash["GERP++_RS_pred"] = "P"
		}
	}

	// 0-0.6 不保守  0.6-2.5 保守 ＞2.5 高度保守
	score, err = strconv.ParseFloat(dataHash["PhyloP Vertebrates"], 32)
	if err != nil {
		dataHash["PhyloP Vertebrates Pred"] = dataHash["PhyloP Vertebrates"]
	} else {
		if score >= 2.5 {
			dataHash["PhyloP Vertebrates Pred"] = "高度保守"
		} else if score > 0.6 {
			dataHash["PhyloP Vertebrates Pred"] = "保守"
		} else {
			dataHash["PhyloP Vertebrates Pred"] = "不保守"
		}
	}
	score, err = strconv.ParseFloat(dataHash["PhyloP Placental Mammals"], 32)
	if err != nil {
		dataHash["PhyloP Placental Mammals Pred"] = dataHash["PhyloP Placental Mammals"]
	} else {
		if score >= 2.5 {
			dataHash["PhyloP Placental Mammals Pred"] = "高度保守"
		} else if score > 0.6 {
			dataHash["PhyloP Placental Mammals Pred"] = "保守"
		} else {
			dataHash["PhyloP Placental Mammals Pred"] = "不保守"
		}
	}

	dataHash["烈性突变"] = "否"
	if FuncInfo[dataHash["Function"]] == 3 {
		dataHash["烈性突变"] = "是"
	}

	dataHash["HGMDorClinvar"] = "否"
	if isHgmd.MatchString(dataHash["HGMD Pred"]) {
		dataHash["HGMDorClinvar"] = "是"
	}
	if isClinvar.MatchString(dataHash["ClinVar Significance"]) {
		dataHash["HGMDorClinvar"] = "是"
	}

	dataHash["GnomAD homo"] = dataHash["GnomAD HomoAlt Count"]
	dataHash["GnomAD hemi"] = dataHash["GnomAD HemiAlt Count"]
	dataHash["纯合，半合"] = dataHash["GnomAD HomoAlt Count"] // + "|" + dataHash["GnomAD HemiAlt Count"]
	if len(strings.Split(dataHash["MutationName"], ":")) > 1 {
		dataHash["MutationNameLite"] = dataHash["Transcript"] + ":" + strings.Split(dataHash["MutationName"], ":")[1]
	} else {
		dataHash["MutationNameLite"] = dataHash["MutationName"]
	}

	//dataHash["突变频谱"] = geneDb[geneSymbol]

	dataHash["历史样本检出个数"] = dataHash["sampleMut"] + "/" + dataHash["sampleAll"]

	// remove index
	for _, k := range [2]string{"GeneralizationEN", "GeneralizationCH"} {
		sep := "\n\n"
		keys := strings.Split(dataHash[k], sep)
		for i := range keys {
			keys[i] = indexReg.ReplaceAllLiteralString(keys[i], "")
		}
		dataHash[k] = strings.Join(keys, sep)
	}

	dataHash["自动化判断"] = long2short[dataHash["ACMG"]]
	return
}

func addTier(item map[string]string, stats map[string]int) {
	gene := item["Gene Symbol"]
	gDiseaseDb := geneDiseaseDb[gene]
	// Tier
	if isDenovo.MatchString(item["Zygosity"]) {
		stats["Denovo"]++
	}
	if noProband.MatchString(item["Zygosity"]) {
		stats["noProband"]++
		return
	}
	if item["ACMG"] != "Benign" && item["ACMG"] != "Likely Benign" {
		stats["noB/LB"]++
		if isDenovo.MatchString(item["Zygosity"]) {
			stats["isDenovo noB/LB"]++
			if checkAF(item, 0.01) {
				stats["low AF"]++
				stats["Denovo AF"]++
				if gDiseaseDb != nil {
					stats["OMIM Gene"]++
					stats["Denovo Gene"]++
					if FuncInfo[item["Function"]] > 1 {
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["Denovo Function"]++
					} else if FuncInfo[item["Function"]] > 0 {
						//pp3,err:=strconv.Atoi(item["PP3"])
						//if err==nil && pp3>0{
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["Denovo Function"]++
					} else {
						item["Tier"] = "Tier2"
						stats["noFunction"]++
						stats["Denovo noFunction"]++
					}
				} else {
					item["Tier"] = "Tier2"
					stats["noB/LB AF noGene"]++
					stats["Denovo noGene"]++
				}
			} else {
				item["Tier"] = "Tier2"
				stats["noB/LB noAF"]++
				stats["Denovo noAF"]++
			}
			if item["Tier"] == "Tier1" {
				stats["Denovo Tier1"]++
			} else {
				stats["Denovo Tier2"]++
			}
		} else {
			stats["noDenovo noB/LB"]++
			if checkAF(item, 0.01) {
				stats["low AF"]++
				stats["noDenovo AF"]++
				if gDiseaseDb != nil {
					stats["OMIM Gene"]++
					stats["noDenovo Gene"]++
					if FuncInfo[item["Function"]] > 1 {
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["noDenovo Function"]++
					} else if FuncInfo[item["Function"]] > 0 {
						//pp3,err:=strconv.Atoi(item["PP3"])
						//if err==nil && pp3>0{
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["noDenovo Function"]++
						//}
					} else {
						item["Tier"] = "Tier3"
						stats["noFunction"]++
						stats["noDenovo noFunction"]++
					}
				} else {
					item["Tier"] = "Tier3"
					stats["noB/LB AF noGene"]++
					stats["noDenovo noGene"]++
				}
			} else {
				item["Tier"] = "Tier3"
				stats["noB/LB noAF"]++
				stats["noDenovo noAF"]++
			}
		}
	} else if isDenovo.MatchString(item["Zygosity"]) {
		stats["Denovo B/LB"]++
	}

	if isHgmd.MatchString(item["HGMD Pred"]) || isClinvar.MatchString(item["ClinVar Significance"]) {
		stats["HGMD/ClinVar"]++
		if checkAF(item, 0.01) {
			item["Tier"] = "Tier1"
			stats["HGMD/ClinVar Tier1"]++
		} else {
			if item["Tier"] != "Tier1" {
				item["Tier"] = "Tier2"
			}
			stats["HGMD/ClinVar Tier2"]++
		}
	}

	if item["Tier"] == "Tier1" {
		stats["Tier1"]++
	} else if item["Tier"] == "Tier2" {
		stats["Tier2"]++
	} else if item["Tier"] == "Tier3" {
		stats["Tier3"]++
	} else {
		return
	}
	stats["Retain"]++
	return
}

var AFlist = []string{
	"GnomAD EAS AF",
	"GnomAD AF",
	"1000G ASN AF",
	"1000G EAS AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC EAS AF",
	"ExAC AF",
	"PVFD AF",
	"Panel AlleleFreq",
}

func checkAF(item map[string]string, threshold float64) bool {
	for _, key := range AFlist {
		af := item[key]
		if af == "" || af == "." {
			continue
		}
		AF, err := strconv.ParseFloat(af, 64)
		simple_util.CheckErr(err)
		if AF > threshold {
			return false
		}
	}
	return true
}
