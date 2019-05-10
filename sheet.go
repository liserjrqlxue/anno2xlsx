package main

import (
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func addFamInfoSheet(excel *xlsx.File, sheetName string, sampleList []string) {
	sheet, err := excel.AddSheet(sheetName)
	simple_util.CheckErr(err)

	sheet.AddRow().AddCell().SetString("SampleID")

	for _, sample := range sampleList {
		sheet.AddRow().AddCell().SetString(sample)
	}
}

func addCnv2Sheet(sheet *xlsx.Sheet, paths []string, sampleMap map[string]bool, filterSize, filterGene bool) {
	cnvDb, _ := simple_util.LongFiles2MapArray(paths, "\t", nil)

	// title
	var title []string
	for _, cell := range sheet.Row(0).Cells {
		title = append(title, cell.Value)
	}

	for _, item := range cnvDb {
		sample := item["Sample"]
		item["Primer"] = anno.CnvPrimer(item, sheet.Name)
		if sampleMap[sample] {
			gene := item["OMIM_Gene"]
			updateDiseaseMultiGene(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
			item["OMIM"] = item["OMIM_Phenotype_ID"]
			if filterGene && item["OMIM"] == "" {
				continue
			}
			if filterSize {
				length, err := strconv.ParseFloat(item["Len(Kb)"], 16)
				if err != nil {
					log.Printf(
						"can not ParseFloat of Len(Kb)[%s] for Summary[%s]\n",
						item["Len(Kb)"], item["Summary"],
					)
				} else if length < 1000 {
					continue
				}
			}
			row := sheet.AddRow()
			for _, key := range title {
				row.AddCell().SetString(item[key])
			}
		}
	}
}

func addSmnResult(sheet *xlsx.Sheet, paths []string, sampleMap map[string]bool) {
	smnDb, _ := simple_util.LongFiles2MapArray(paths, "\t", nil)

	// title
	var title []string
	for _, cell := range sheet.Row(0).Cells {
		title = append(title, cell.Value)
	}

	for _, item := range smnDb {
		sample := item["SampleID"]
		if sampleMap[sample] {
			item["Sample"] = item["SampleID"]
			item["Copy_Num"] = item["SMN1_ex7_cn"]
			item["Detect"] = item["SMN1_ex7_cn"]
			item["Chr"] = "chr5"
			item["Start"] = "70241892"
			item["End"] = "70242003"
			item["Gene"] = "SMN1"
			item["OMIM_Gene"] = "SMN1"
			item["SMN1_result"] = item["SMN1_ex7_cn"]
			if item["SMN1_ex7_cn"] == "0" {
				item["SMN1_result"] = "Hom"
				isSMN1 = true
			}
			row := sheet.AddRow()
			for _, key := range title {
				row.AddCell().SetString(item[key])
			}
		}
	}
}

func updateDisease(gene string, item, geneDiseaseDbColumn map[string]string, geneDiseaseDb map[string]map[string]string) {
	// 基因-疾病
	gDiseaseDb := geneDiseaseDb[gene]
	for key, value := range geneDiseaseDbColumn {
		item[value] = gDiseaseDb[key]
	}
}

func updateDiseaseMultiGene(geneList string, item, geneDiseaseDbColumn map[string]string, geneDiseaseDb map[string]map[string]string) {
	genes := strings.Split(geneList, ";")
	// 基因-疾病
	for key, value := range geneDiseaseDbColumn {
		var vals []string
		for _, gene := range genes {
			geneDb, ok := geneDiseaseDb[gene]
			if ok {
				vals = append(vals, geneDb[key])
			}
			//fmt.Println(gene,":",key,":",vals)
		}
		if len(vals) > 0 {
			item[value] = strings.Join(vals, "\n")
		}
	}
}

var isSharp = regexp.MustCompile(`^#`)
var isBamPath = regexp.MustCompile(`^## Files : (\S+)`)

func loadQC(files string, quality []map[string]string) {
	file := strings.Split(files, ",")
	for i, in := range file {
		report := simple_util.File2Array(in)
		for _, line := range report {
			if isSharp.MatchString(line) {
				if m := isBamPath.FindStringSubmatch(line); m != nil {
					quality[i]["bamPath"] = m[1]
				}
			} else {
				m := strings.Split(line, "\t")
				quality[i][strings.TrimSpace(m[0])] = strings.TrimSpace(m[1])
			}
		}
	}
}

var MTTitle = []string{
	"#Chr",
	"Start",
	"Stop",
	"Ref",
	"Call",
	"MutationName",
	"Disease",
	"pmid",
	"title",
	"Status",
	"Mito TIP",
}

type Variant struct {
	Chr   string                 `json:"Chromosome"`
	Ref   string                 `json:"Ref"`
	Alt   string                 `json:"Alt"`
	Start int                    `json:"Start"`
	End   int                    `json:"End"`
	Info  map[string]interface{} `json:"Info"`
}

func addMTRow(sheet *xlsx.Sheet, item map[string]string) {
	rowMT := sheet.AddRow()
	key := strings.Join([]string{"MT", item["Start"], item["Stop"], item["Ref"], item["Call"]}, "\t")
	mut, ok := TIPdb[key]
	if ok {
		for _, key := range []string{"Mito TIP"} {
			item[key] = mut.Info[key].(string)
		}
	}
	mut, ok = MTdisease[key]
	if ok {
		for _, key := range []string{"Disease", "pmid", "title", "Status"} {
			item[key] = mut.Info[key].(string)
		}
	}
	for _, str := range MTTitle {
		rowMT.AddCell().SetString(item[str])
	}
}
