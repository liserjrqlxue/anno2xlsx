package anno

import (
	"regexp"
	"strconv"
	"strings"
)

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
	indexReg = regexp.MustCompile(`\d+\.\s+`)

	isARorXR = regexp.MustCompile(`AR|XR`)
	isAR     = regexp.MustCompile(`AR`)
	isAD     = regexp.MustCompile(`AD`)
	isXL     = regexp.MustCompile(`XL`)

	isHet  = regexp.MustCompile(`^Het`)
	isHom  = regexp.MustCompile(`^Hom`)
	isHemi = regexp.MustCompile(`^Hemi`)
	isNA   = regexp.MustCompile(`^NA`)

	isHetHetHet = regexp.MustCompile(`^Het;Het;Het`)
	isHetHetNA  = regexp.MustCompile(`^Het;Het;NA`)
	isHetNAHet  = regexp.MustCompile(`^Het;NA;Het`)
	isHetNANA   = regexp.MustCompile(`^Het;NA;NA`)

	isHomInherit  = regexp.MustCompile(`^Hom;Het;Het|^Hom;Het;NA|^Hom;NA;Het|^Hom;NA;NA`)
	isHemiInherit = regexp.MustCompile(`^Hemi;Het;NA|^Hemi;NA;Het|^Hemi;NA;NA|^Het;NA;NA`)
)

func UpdateSnv(dataHash map[string]string) {
	// Zygosity format
	dataHash["Zygosity"] = zygosityFormat(dataHash["Zygosity"])

	// pHGVS= pHGVS1+"|"+pHGVS3
	if dataHash["pHGVS1"] != "" && dataHash["pHGVS3"] != "" {
		dataHash["pHGVS"] = dataHash["pHGVS1"] + " | " + dataHash["pHGVS3"]
	}

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

func InheritCheck(item map[string]string, inheritDb map[string]map[string]int) {
	geneSymbol := item["Gene Symbol"]
	inherit := item["ModeInheritance"]
	zygosity := item["Zygosity"]
	var db = make(map[string]int)
	if inheritDb[geneSymbol] == nil {
		inheritDb[geneSymbol] = db
	}
	if isARorXR.MatchString(inherit) {
		if isHet.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag1"]++
		}
		if isHetHetNA.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag110"]++
		}
		if isHetNAHet.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag101"]++
		}
		if isHetNANA.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag100"]++
		}
	}
}

func InheritCoincide(item map[string]string, inheritDb map[string]map[string]int, isTrio bool) string {
	geneSymbol := item["Gene Symbol"]
	inherit := item["ModeInheritance"]
	zygosity := item["Zygosity"]
	if isTrio {
		if isNA.MatchString(zygosity) {
			return "NA"
		}
		if isAD.MatchString(inherit) && isHetNANA.MatchString(zygosity) {
			return "相符"
		}
		if isXL.MatchString(inherit) && isHemiInherit.MatchString(zygosity) {
			return "相符"
		}
		if isAR.MatchString(inherit) {
			if isHomInherit.MatchString(zygosity) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag110"] > 0 &&
				inheritDb[geneSymbol]["flag101"] > 0 &&
				(isHetHetNA.MatchString(zygosity) || isHetNAHet.MatchString(zygosity)) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag110"] > 0 &&
				inheritDb[geneSymbol]["flag100"] > 0 &&
				(isHetHetNA.MatchString(zygosity) || isHetNANA.MatchString(zygosity)) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag101"] > 0 &&
				inheritDb[geneSymbol]["flag100"] > 0 &&
				(isHetNAHet.MatchString(zygosity) || isHetNANA.MatchString(zygosity)) {
				return "相符"
			}
			if isHetHetHet.MatchString(zygosity) ||
				(inheritDb[geneSymbol]["flag100"] >= 2 && isHetNANA.MatchString(zygosity)) {
				return "不确定"
			}
		}
		return "不相符"
	} else {
		if (isHet.MatchString(zygosity) && isARorXR.MatchString(inherit)) ||
			(isHom.MatchString(zygosity) && isAD.MatchString(inherit)) ||
			(isHemi.MatchString(zygosity) && isXL.MatchString(inherit)) {
			return "相符"
		} else if isARorXR.MatchString(zygosity) && inheritDb[geneSymbol]["flag1"] >= 2 {
			return "不确定"
		} else {
			return "不相符"
		}
	}
}

func zygosityFormat(zygosity string) string {
	zygosity = strings.Replace(zygosity, "het-ref", "Het", -1)
	zygosity = strings.Replace(zygosity, "het-alt", "Het", -1)
	zygosity = strings.Replace(zygosity, "hom-alt", "Hom", -1)
	zygosity = strings.Replace(zygosity, "hem-alt", "Hemi", -1)
	zygosity = strings.Replace(zygosity, "hemi-alt", "Hemi", -1)
	return zygosity
}
