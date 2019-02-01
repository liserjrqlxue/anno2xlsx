package main

import (
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"os"
	"path/filepath"
	"time"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	pSep         = string(os.PathSeparator)
	dbPath       = exPath + pSep + "db" + pSep
	templatePath = exPath + pSep + "template" + pSep
)

// flag
var (
	input = flag.String(
		"input",
		"",
		"input anno txt",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output xlsx prefix.tier{1,2,3}.xlsx",
	)
	geneDbExcel = flag.String(
		"geneDb",
		dbPath+"基因库-更新版基因特征谱-加动态突变-20190110.xlsx",
		"database of 突变频谱",
	)
	geneDbSheet = flag.String(
		"geneDbSheet",
		"Sheet1",
		"sheet name of 突变频谱 database in excel",
	)
	geneDiseaseDbExcel = flag.String(
		"geneDisease",
		dbPath+"全外基因基因集整理OMIM-20190122.xlsx",
		"database of 基因-疾病数据库",
	)
	geneDiseaseSheet = flag.String(
		"geneDiseaseSheet",
		"Database",
		"sheet name of geneDiseaseDbExcel",
	)
	specVarList = flag.String(
		"specVarList",
		dbPath+"spec.var.list",
		"特殊位点库")
	save = flag.Bool(
		"save",
		true,
		"if save to excel",
	)
	trio = flag.Bool(
		"trio",
		false,
		"if trio mode",
	)
)

// 突变频谱
var geneDb = make(map[string]string)

// 基因-疾病
var geneList = make(map[string]bool)
var geneDiseaseDb = make(map[string]map[string]string)
var geneDiseaseDbColumn = map[string]string{
	"Gene/Locus":                 "Gene",
	"Phenotype MIM number":       "OMIM",
	"Disease NameEN":             "DiseaseNameEN",
	"Disease NameCH":             "DiseaseNameCH",
	"Alternative Disease NameEN": "AliasEN",
	"Location":                   "Location",
	"Gene/Locus MIM number":      "Gene/Locus MIM number",
	"Inheritance":                "ModeInheritance",
	"GeneralizationEN":           "GeneralizationEN",
	"GeneralizationCH":           "GeneralizationCH",
	"SystemSort":                 "SystemSort",
}

// 特殊位点库
var specVarDb = make(map[string]bool)

// 遗传相符
var inheritDb = make(map[string]map[string]int)

type xlsxTemplate struct {
	flag      string
	template  string
	xlsx      *xlsx.File
	sheetName string
	sheet     *xlsx.Sheet
	title     []string
	output    string
}

var tier2xlsx = map[string]map[string]bool{
	"Tier1": {
		"Tier1": true,
	},
	"Tier2": {
		"Tier1": true,
		"Tier2": true,
	},
	"Tier3": {
		"Tier1": true,
		"Tier2": true,
		"Tier3": true,
	},
}

var err error

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	flag.Parse()
	if *input == "" || *prefix == "" {
		flag.Usage()
		os.Exit(0)
	}

	// load tier template
	var tiers = make(map[string]xlsxTemplate)
	var tierSheet = map[string]string{
		"Tier1": "filter_variants",
		"Tier2": "附表",
		"Tier3": "总表",
	}
	for key, value := range tierSheet {
		var tier = xlsxTemplate{
			flag:      key,
			template:  templatePath + key + ".xlsx",
			sheetName: value,
			output:    *prefix + "." + key + ".xlsx",
		}
		tier.xlsx, err = xlsx.OpenFile(tier.template)
		simple_util.CheckErr(err)
		tier.sheet = tier.xlsx.Sheet[tier.sheetName]
		for _, cell := range tier.sheet.Row(0).Cells {
			tier.title = append(tier.title, cell.String())
		}
		tiers[key] = tier
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load template")

	// 突变频谱
	loadGeneDb(*geneDbExcel, *geneDbSheet, geneDb)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load 突变频谱")

	// 基因-疾病
	loadGeneDiseaseDb(*geneDiseaseDbExcel, *geneDiseaseSheet, geneDiseaseDb, geneList)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load 基因-疾病")

	// 特殊位点库
	for _, key := range simple_util.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load 特殊位点库")

	// anno
	data, _ := simple_util.File2MapArray(*input, "\t")
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load anno file")

	var stats = make(map[string]int)

	stats["Total"] = len(data)
	for _, item := range data {
		anno.UpdateSnv(item)
		gene := item["Gene Symbol"]
		// 突变频谱
		item["突变频谱"] = geneDb[gene]
		// 基因-疾病
		gDiseaseDb := geneDiseaseDb[gene]
		for key, value := range geneDiseaseDbColumn {
			item[value] = gDiseaseDb[key]
		}

		anno.AddTier(item, stats, geneList, specVarDb, *trio)

		// 遗传相符
		// only for Tier1
		if item["Tier"] == "Tier1" {
			anno.InheritCheck(item, inheritDb)
		}
	}
	for _, item := range data {
		// 遗传相符
		item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
		// add to excel
		for flg := range tierSheet {
			if tier2xlsx[flg][item["Tier"]] {
				tierRow := tiers[flg].sheet.AddRow()
				for _, str := range tiers[flg].title {
					tierRow.AddCell().SetString(item[str])
				}
			}
		}
	}

	logTierStats(stats)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "update info")

	if *save {
		for flg := range tierSheet {
			err = tiers[flg].xlsx.Save(tiers[flg].output)
			simple_util.CheckErr(err)
			ts = append(ts, time.Now())
			step++
			logTime(ts, step-1, step, "save "+flg)
		}
	}
	logTime(ts, 0, step, "total work")
}

func logTierStats(stats map[string]int) {
	fmt.Printf("Total               Count : %7d\n", stats["Total"])
	if *trio {
		fmt.Printf("  noProband         Count : %7d\n", stats["noProband"])

		fmt.Printf("Denovo              Hit   : %7d\n", stats["Denovo"])
		fmt.Printf("  Denovo B/LB       Hit   : %7d\n", stats["Denovo B/LB"])
		fmt.Printf("  Denovo Tier1      Hit   : %7d\n", stats["Denovo Tier1"])
		fmt.Printf("  Denovo Tier2      Hit   : %7d\n", stats["Denovo Tier2"])
	}

	fmt.Printf("ACMG noB/LB         Hit   : %7d\n", stats["noB/LB"])
	if *trio {
		fmt.Printf("  +isDenovo         Hit   : %7d\n", stats["isDenovo noB/LB"])
		fmt.Printf("    +isAF           Hit   : %7d\n", stats["Denovo AF"])
		fmt.Printf("      +isGene       Hit   : %7d\n", stats["Denovo Gene"])
		fmt.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["Denovo Function"])
		fmt.Printf("        +noFunction Hit   : %7d\n", stats["Denovo noFunction"])
		fmt.Printf("      +noGene       Hit   : %7d\n", stats["Denovo noGene"])
		fmt.Printf("    +noAF           Hit   : %7d\n", stats["Denovo noAF"])
		fmt.Printf("  +noDenovo         Hit   : %7d\n", stats["noDenovo noB/LB"])
		fmt.Printf("    +isAF           Hit   : %7d\n", stats["noDenovo AF"])
		fmt.Printf("      +isGene       Hit   : %7d\n", stats["noDenovo Gene"])
		fmt.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["noDenovo Function"])
		fmt.Printf("        +noFunction Hit   : %7d\n", stats["noDenovo noFunction"])
		fmt.Printf("      +noGene       Hit   : %7d\n", stats["noDenovo noGene"])
		fmt.Printf("    +noAF           Hit   : %7d\n", stats["noDenovo noAF"])
	} else {
		fmt.Printf("    +isAF           Hit   : %7d\n", stats["isAF"])
		fmt.Printf("      +isGene       Hit   : %7d\n", stats["isGene"])
		fmt.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["isFunction"])
		fmt.Printf("        +noFunction Hit   : %7d\n", stats["noFunction"])
		fmt.Printf("      +noGene       Hit   : %7d\n", stats["noGene"])
		fmt.Printf("    +noAF           Hit   : %7d\n", stats["noAF"])
	}

	fmt.Printf("HGMD/ClinVar        Hit   : %7d\n", stats["HGMD/ClinVar"])
	fmt.Printf("  isAF              Hit   : %7d\n", stats["HGMD/ClinVar isAF"])
	fmt.Printf("    noMT            Hit   : %7d\tTier1\n", stats["HGMD/ClinVar noMT T1"])
	fmt.Printf("  noAF              Hit   : %7d\n", stats["HGMD/ClinVar noAF"])
	fmt.Printf("    noMT            Hit   : %7d\tTier2\n", stats["HGMD/ClinVar noMT T2"])

	fmt.Printf("SpecVar             Hit   : %7d\n", stats["SpecVar"])

	fmt.Printf("Retain              Count : %7d\n", stats["Retain"])
	fmt.Printf("  Tier1             Count : %7d\n", stats["Tier1"])
	fmt.Printf("  Tier2             Count : %7d\n", stats["Tier2"])
	fmt.Printf("  Tier3             Count : %7d\n", stats["Tier3"])
}

func logTime(timeList []time.Time, step1, step2 int, message string) {
	trim := 3*8 - 1
	str := simple_util.FormatWidth(trim, message, ' ')
	fmt.Printf("%s\ttook %7.3fs to run.\n", str, timeList[step2].Sub(timeList[step1]).Seconds())
}

func loadGeneDb(excelFile, sheetName string, geneDb map[string]string) {
	xlsxFh, err := excelize.OpenFile(excelFile)
	simple_util.CheckErr(err)
	rows := xlsxFh.GetRows(sheetName)
	var title []string

	for i, row := range rows {
		if i == 0 {
			title = row
		} else {
			var dataHash = make(map[string]string)
			for j, cell := range row {
				dataHash[title[j]] = cell
			}
			if geneDb[dataHash["基因名"]] == "" {
				geneDb[dataHash["基因名"]] = dataHash["突变/致病多样性-补充/更正"]
			} else {
				geneDb[dataHash["基因名"]] = geneDb[dataHash["基因名"]] + ";" + dataHash["突变/致病多样性-补充/更正"]
			}
		}
	}
	return
}

func loadGeneDiseaseDb(excelFile, sheetName string, geneDiseaseDb map[string]map[string]string, geneList map[string]bool) {
	xlsxFh, err := excelize.OpenFile(excelFile)
	simple_util.CheckErr(err)
	rows := xlsxFh.GetRows(sheetName)
	var title []string

	for i, row := range rows {
		if i == 0 {
			title = row
		} else {
			var dataHash = make(map[string]string)
			for j, cell := range row {
				dataHash[title[j]] = cell
			}
			gene := dataHash["Gene/Locus"]
			geneList[gene] = true
			if geneDiseaseDb[gene] == nil {
				geneDiseaseDb[gene] = dataHash
			} else {
				for _, key := range title {
					geneDiseaseDb[gene][key] = geneDiseaseDb[gene][key] + "\n" + dataHash[key]
				}
			}
		}
	}
	return
}
