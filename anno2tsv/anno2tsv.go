package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/simple-util"
	"os"
	"path/filepath"
	"time"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	pSep   = string(os.PathSeparator)
	dbPath = exPath + pSep + ".." + pSep + "db" + pSep
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
	gender = flag.String(
		"gender",
		"",
		"gender of sample",
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

var w1, w2, w3 *csv.Writer

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	flag.Parse()
	if *input == "" || *prefix == "" {
		flag.Usage()
		os.Exit(0)
	}

	// open file to write
	if *save {
		f1 := writeTsv(w1, *prefix+".解读表.tsv")
		defer simple_util.DeferClose(f1)
		f2 := writeTsv(w2, *prefix+".附表.tsv")
		defer simple_util.DeferClose(f2)
		f3 := writeTsv(w3, *prefix+".总表.tsv")
		defer simple_util.DeferClose(f3)
	}

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
	data, title := simple_util.File2MapArray(*input, "\t", nil)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load anno file")

	var stats = make(map[string]int)

	stats["Total"] = len(data)
	title = append(title, "突变频谱", "Tier",
		"pHGVS", "dbscSNV_ADA_pred", "dbscSNV_RF_pred", "GERP++_RS_pred", "PhyloP Vertebrates Pred", "PhyloP Placental Mammals Pred",
		"烈性突变", "HGMDorClinvar", "GnomAD homo", "GnomAD hemi", "纯合，半合", "MutationNameLite", "历史样本检出个数", "自动化判断")
	for _, value := range geneDiseaseDbColumn {
		title = append(title, value)
	}
	if *save {
		simple_util.CheckErr(w1.Write(title))
		simple_util.CheckErr(w2.Write(title))
		simple_util.CheckErr(w3.Write(title))
	}
	for _, item := range data {
		anno.UpdateSnv(item, *gender, false)
		gene := item["Gene Symbol"]
		// 突变频谱
		item["突变频谱"] = geneDb[gene]
		// 基因-疾病
		gDiseaseDb := geneDiseaseDb[gene]
		for key, value := range geneDiseaseDbColumn {
			item[value] = gDiseaseDb[key]
		}
		var arr []string
		for _, key := range title {
			arr = append(arr, item[key])
		}
		anno.AddTier(item, stats, geneList, specVarDb, *trio, false, false)

		if *save {
			simple_util.CheckErr(w1.Write(arr))
			simple_util.CheckErr(w2.Write(arr))
			simple_util.CheckErr(w3.Write(arr))
		}

	}
	w1.Flush()
	simple_util.CheckErr(w1.Error())
	w2.Flush()
	simple_util.CheckErr(w2.Error())
	w3.Flush()
	simple_util.CheckErr(w3.Error())

	logTierStats(stats)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "update info and write")

	logTime(ts, 0, step, "total work")
}

func logTierStats(stats map[string]int) {
	fmt.Printf("Total               Count : %7d\n", stats["Total"])
	fmt.Printf("  noProband         Count : %7d\n", stats["noProband"])

	fmt.Printf("Denovo              Hit   : %7d\n", stats["Denovo"])
	fmt.Printf("  Denovo B/LB       Hit   : %7d\n", stats["Denovo B/LB"])
	fmt.Printf("  Denovo Tier1      Hit   : %7d\n", stats["Denovo Tier1"])
	fmt.Printf("  Denovo Tier2      Hit   : %7d\n", stats["Denovo Tier2"])

	fmt.Printf("ACMG noB/LB         Hit   : %7d\n", stats["noB/LB"])
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
	rows, err := xlsxFh.GetRows(sheetName)
	simple_util.CheckErr(err)
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
	rows, err := xlsxFh.GetRows(sheetName)
	simple_util.CheckErr(err)
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

func writeTsv(w *csv.Writer, name string) *os.File {
	f, err := os.Create(name)
	simple_util.CheckErr(err)
	w = csv.NewWriter(f)
	w.Comma = '\t'
	w.UseCRLF = false
	return f
}
