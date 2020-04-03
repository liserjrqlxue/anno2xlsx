package main

import (
	"github.com/liserjrqlxue/goUtil/textUtil"
	"log"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v2"

	"github.com/liserjrqlxue/anno2xlsx/anno"
)

func addFamInfoSheet(excel *xlsx.File, sheetName string, sampleList []string) {
	var sheet, err = excel.AddSheet(sheetName)
	simple_util.CheckErr(err)

	sheet.AddRow().AddCell().SetString("SampleID")

	for _, sample := range sampleList {
		sheet.AddRow().AddCell().SetString(sample)
	}
}

func addQCSheet(excel *xlsx.File, sheetName string, qualityColumn []string, qualitys []map[string]string) {
	sheet, err := excel.AddSheet(sheetName)
	simple_util.CheckErr(err)

	for _, key := range qualityColumn {
		row := sheet.AddRow()
		row.AddCell().SetString(key)
		for _, item := range qualitys {
			row.AddCell().SetString(item[key])
		}
	}
}

func addTxt2Sheet(excel *xlsx.File, sheetName, file string) {
	sheet, err := excel.AddSheet(sheetName)
	simple_util.CheckErr(err)

	slice := simple_util.File2Slice(file, "\t")
	for _, line := range slice {
		row := sheet.AddRow()
		for _, val := range line {
			row.AddCell().SetString(val)
		}
	}
}

func addCnv2Sheet(
	sheet *xlsx.Sheet, paths []string, sampleMap map[string]bool, filterSize, filterGene bool, stats map[string]int,
	key, titleFile string) {
	cnvDb, _ := simple_util.LongFiles2MapArray(paths, "\t", nil)

	// title
	title := textUtil.File2Array(titleFile)
	titleRow := sheet.Row(0)
	titleCells := titleRow.Cells
	for i, v := range title {
		if i < len(titleCells) {
			titleRow.Cells[i].SetString(v)
		} else {
			titleRow.AddCell().SetString(v)
		}
	}

	for _, item := range cnvDb {
		sample := item["Sample"]
		item["Primer"] = anno.CnvPrimer(item, sheet.Name)
		if sampleMap[sample] {
			gene := item["OMIM_Gene"]
			updateDiseMultiGene(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
			// 突变频谱
			updateGeneDb(gene, item, geneDb)
			item["OMIM"] = item["OMIM_Phenotype_ID"]
			stats[key]++
			if item["OMIM"] != "" {
				stats["Tier1"+key]++
			}
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
			for _, str := range title {
				row.AddCell().SetString(item[str])
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

func updateDiseMultiGene(geneLst string, item, geneDisDbCol map[string]string, geneDisDb map[string]map[string]string) {
	genes := strings.Split(geneLst, ";")
	// 基因-疾病
	for key, value := range geneDisDbCol {
		var vals []string
		for _, gene := range genes {
			singelGeneDb, ok := geneDisDb[gene]
			if ok {
				vals = append(vals, singelGeneDb[key])
			}
		}
		if len(vals) > 0 {
			item[value] = strings.Join(vals, "\n")
		}
	}
}

func updateGeneDb(geneList string, item, geneDb map[string]string) {
	genes := strings.Split(geneList, ";")
	// 突变频谱
	var vals []string
	for _, gene := range genes {
		vals = append(vals, geneDb[gene])
	}
	item["突变频谱"] = strings.Join(vals, "\n")
}

//Variant struct for anno info
type variant struct {
	Chr   string                 `json:"Chromosome"`
	Ref   string                 `json:"Ref"`
	Alt   string                 `json:"Alt"`
	Start int                    `json:"Start"`
	End   int                    `json:"End"`
	Info  map[string]interface{} `json:"Info"`
}

func addMTRow(sheet *xlsx.Sheet, item map[string]string) {
	rowMT := sheet.AddRow()
	ref := item["Ref"]
	alt := item["Call"]
	if ref == "." {
		ref = ""
	}
	if alt == "." {
		alt = ""
	}
	key := strings.Join([]string{"MT", item["Start"], item["Stop"], ref, alt}, "\t")
	mut, ok := TIPdb[key]
	if ok {
		for _, str := range []string{"Mito TIP"} {
			item[str] = mut.Info[str].(string)
		}
	}
	mut, ok = MTdisease[key]
	if ok {
		for _, str := range []string{"Disease", "pmid", "title", "Status"} {
			item[str] = mut.Info[str].(string)
		}
	}
	mut, ok = MTAFdb[key]
	if ok {
		for _, str := range []string{"# in HG branch with variant", "Total # HG branch seqs", "Fequency in HG branch(%)"} {
			item[str] = strconv.FormatFloat(mut.Info[str].(float64), 'f', 5, 64)
		}

	}
	for _, str := range MTTitle {
		rowMT.AddCell().SetString(item[str])
	}
}

func addTier2Row(tier2 xlsxTemplate, item map[string]string) {
	tier2Row := tier2.sheet.AddRow()
	for _, str := range tier2.title {
		switch str {
		case "HGMDorClinvar":
			if isEnProduct[*productID] {
				tier2Row.AddCell().SetString(transEN[item[str]])
			} else {
				tier2Row.AddCell().SetString(item[str])
			}
		case "DiseaseName/ModeInheritance":
			inheritance := strings.Split(item["ModeInheritance"], "\n")
			var disease []string
			if isEnProduct[*productID] {
				disease = strings.Split(item["DiseaseNameEN"], "\n")
			} else {
				disease = strings.Split(item["DiseaseNameCH"], "\n")
			}
			if len(disease) == len(inheritance) {
				for i, text := range disease {
					inheritance[i] = text + "/" + inheritance[i]
				}
			} else {
				log.Fatalf("Disease error:%s\t%v vs %v\n", item["Gene Symbol"], disease, inheritance)
			}
			tier2Row.AddCell().SetString(strings.Join(inheritance, "\n"))
			item["DiseaseName/ModeInheritance"] = strings.Join(inheritance, "<br>")
		default:
			tier2Row.AddCell().SetString(item[str])
		}
	}
}
