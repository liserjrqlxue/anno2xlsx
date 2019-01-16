package main

import (
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	pSep   = string(os.PathSeparator)
	dbPath = exPath + pSep + "db" + pSep
)

// flag
var (
	input = flag.String(
		"input",
		"",
		"input anno txt",
	)
	output = flag.String(
		"output",
		"",
		"output xlsx name",
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
		dbPath+"全外基因-疾病集-20190109.xlsx",
		"database of 基因-疾病数据库",
	)
	geneDiseaseSheet = flag.String(
		"geneDiseaseSheet",
		"Database",
		"sheet name of geneDiseaseDbExcel",
	)
	save = flag.Bool(
		"save",
		true,
		"if save to excel")
)

// regexp
var (
	isHgmd    = regexp.MustCompile("DM")
	isClinvar = regexp.MustCompile("Pathogenic|Likely_pathogenic")
	//indexReg   = regexp.MustCompile(`\d+\.\s+`)
	//newlineReg = regexp.MustCompile(`\n+`)
)

// 突变频谱
var geneDb = make(map[string]string)

// 基因-疾病
var geneDiseaseDb = make(map[string]map[string]string)
var geneDiseaseDbColumn = []string{
	"Disease NameEN",
	"Disease NameCH",
	"Inheritance",
	"GeneralizationEN",
	"GeneralizationCH",
	"SystemSort",
}

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

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	flag.Parse()
	if *input == "" || *output == "" {
		flag.Usage()
		os.Exit(0)
	}

	// 突变频谱
	geneDb = loadGeneDb(*geneDbExcel, *geneDbSheet)
	ts = append(ts, time.Now())
	step++
	fmt.Printf("load 突变频谱 \ttook %v to run.\n", ts[step].Sub(ts[step-1]))

	// 基因-疾病
	geneDiseaseDb = loadGeneDiseaseDb(*geneDiseaseDbExcel, *geneDiseaseSheet)
	ts = append(ts, time.Now())
	step++
	fmt.Printf("load 基因-疾病 \ttook %v to run.\n", ts[step].Sub(ts[step-1]))

	// anno
	data, title := simple_util.File2MapArray(*input, "\t")
	title = append(title, "Tier", "突变频谱")
	title = append(title, geneDiseaseDbColumn...)
	ts = append(ts, time.Now())
	step++
	fmt.Printf("load anno file \ttook %v to run.\n", ts[step].Sub(ts[step-1]))

	var stats = make(map[string]int)
	outputXlsx := xlsx.NewFile()
	outputSheet, err := outputXlsx.AddSheet("filter_variant")
	simple_util.CheckErr(err)

	var outputRow = outputSheet.AddRow()

	for _, str := range title {
		outputCell := outputRow.AddCell()
		outputCell.SetString(str)
	}

	stats["Total"] = len(data)
	for _, item := range data {
		gene := item["Gene Symbol"]
		// 突变频谱
		item["突变频谱"] = geneDb[gene]
		// 基因-疾病
		gDiseaseDb := geneDiseaseDb[gene]
		for _, key := range geneDiseaseDbColumn {
			item[key] = gDiseaseDb[key]
		}

		// Tier
		if isHgmd.MatchString(item["HGMD Pred"]) || isClinvar.MatchString(item["ClinVar Significance"]) {
			stats["HGMD/ClinVar"]++
			if checkAF(item, 0.01) {
				item["Tier"] = "Tier1"
				stats["HGMD/ClinVar Tier1"]++
			} else {
				item["Tier"] = "Tier2"
				stats["HGMD/ClinVar Tier2"]++
			}
		} else if item["ACMG"] == "Benign" || item["ACMG"] == "Likely Benign" {
			stats["B/LB"]++
			continue
		}

		if item["ACMG"] != "Benign" && item["ACMG"] != "Likely Benign" {
			stats["noB/LB"]++
			if checkAF(item, 0.01) {
				stats["low AF"]++
				if gDiseaseDb != nil {
					stats["OMIM Gene"]++
					if FuncInfo[item["Function"]] > 1 {
						item["Tier"] = "Tier1"
						stats["Function"]++
					} else if FuncInfo[item["Function"]] > 0 {
						//pp3,err:=strconv.Atoi(item["PP3"])
						//if err==nil && pp3>0{
						item["Tier"] = "Tier1"
						stats["Function"]++
						//}
					}
				} else {
					stats["noB/LB AF noGene"]++
				}
			} else {
				stats["noB/LB noAF"]++
			}
			if item["Tier"] != "Tier1" {
				item["Tier"] = "Tier2"
			}
		}

		if item["Tier"] == "Tier1" {
			stats["Tier1"]++
		} else if item["Tier"] == "Tier2" {
			stats["Tier2"]++
		}
		stats["Retain"]++

		outputRow = outputSheet.AddRow()
		for _, str := range title {
			outputCell := outputRow.AddCell()
			outputCell.SetString(item[str])
		}
	}

	fmt.Printf("Total         Count  : %d\n", stats["Total"])
	fmt.Printf("HGMD/ClinVar  Hit    : %d\n", stats["HGMD/ClinVar"])
	fmt.Printf(" HGMD/ClinVar Tier1  : %d\n", stats["HGMD/ClinVar Tier1"])
	fmt.Printf(" HGMD/ClinVar Tier2  : %d\n", stats["HGMD/ClinVar Tier2"])
	fmt.Printf("ACMG         Hit     : %d\n", stats["noB/LB"])
	fmt.Printf(" lowAF       Hit     : %d\n", stats["low AF"])
	fmt.Printf("  OMIM Gene  Hit     : %d\n", stats["OMIM Gene"])
	fmt.Printf("   Function  Hit     : %d\n", stats["Function"])
	fmt.Printf("Retain       Count   : %d\n", stats["Retain"])
	fmt.Printf(" Tier1       Count   : %d\n", stats["Tier1"])
	fmt.Printf(" Tier2       Count   : %d\n", stats["Tier2"])
	ts = append(ts, time.Now())
	step++
	fmt.Printf("create excel \ttook %v to run.\n", ts[step].Sub(ts[step-1]))

	if *save {
		err = outputXlsx.Save(*output)
		simple_util.CheckErr(err)
		ts = append(ts, time.Now())
		step++
		fmt.Printf("save excel \ttook %v to run.\n", ts[step].Sub(ts[step-1]))
	}

	fmt.Printf("total work \ttook %v to run.\n", ts[step].Sub(ts[0]))
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

func loadGeneDb(excelFile, sheetName string) map[string]string {
	xlsxFh, err := excelize.OpenFile(excelFile)
	simple_util.CheckErr(err)
	rows := xlsxFh.GetRows(sheetName)
	var title []string
	var geneDb = make(map[string]string)

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
	return geneDb
}

func loadGeneDiseaseDb(excelFile, sheetName string) map[string]map[string]string {
	xlsxFh, err := excelize.OpenFile(excelFile)
	simple_util.CheckErr(err)
	rows := xlsxFh.GetRows(sheetName)
	var title []string
	var geneDiseaseDb = make(map[string]map[string]string)

	for i, row := range rows {
		if i == 0 {
			title = row
		} else {
			var dataHash = make(map[string]string)
			for j, cell := range row {
				dataHash[title[j]] = cell
			}
			gene := dataHash["Gene/Locus"]
			if geneDiseaseDb[gene] == nil {
				geneDiseaseDb[gene] = dataHash
			} else {
				//var newDataHash=make(map[string]string)
				for _, key := range title {
					//newDataHash[key]=geneDiseaseDb[gene][key]+"\n"+dataHash[key]
					geneDiseaseDb[gene][key] = geneDiseaseDb[gene][key] + "\n" + dataHash[key]
				}
			}
		}
	}

	return geneDiseaseDb
}
