package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v3"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

func addFile2Row(file string, row *xlsx.Row) (rows []string) {
	rows = textUtil.File2Array(file)
	xlsxUtil.AddArray2Row(rows, row)
	return
}

func addFamInfoSheet(excel *xlsx.File, sheetName string, sampleList []string) {
	var sheet = xlsxUtil.AddSheet(excel, sheetName)
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

func addCnv2Sheet(
	sheet *xlsx.Sheet, title, paths []string, sampleMap map[string]bool, filterSize, filterGene bool, stats map[string]int,
	key string) {
	cnvDb, _ := simple_util.LongFiles2MapArray(paths, "\t", nil)

	for _, item := range cnvDb {
		sample := item["Sample"]
		item["Primer"] = anno.CnvPrimer(item, sheet.Name)
		if sampleMap[sample] {
			var gene = item["OMIM_Gene"]

			var geneIDs []string
			for _, g := range strings.Split(gene, ";") {
				var id, ok = gene2id[g]
				if !ok {
					if g != "-" && g != "." {
						log.Fatalf("can not find gene id of [%s]\n", gene)
					}
				}
				geneIDs = append(geneIDs, id)
			}

			// 基因-疾病
			diseaseDb.Annos(item, "\n", geneIDs)
			// 突变频谱
			spectrumDb.Annos(item, "\n", geneIDs)

			if *cnvAnnot {
				anno.UpdateCnvAnnot(gene, item, gene2id, diseaseDb.Db)
			}

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
			xlsxUtil.AddMap2Row(item, title, sheet.AddRow())
		}
	}
}

func addSmnResult(sheet *xlsx.Sheet, title, paths []string, sampleMap map[string]bool) {
	smnDb, _ := simple_util.LongFiles2MapArray(paths, "\t", nil)

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
			xlsxUtil.AddMap2Row(item, title, sheet.AddRow())
		}
	}
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
	xlsxUtil.AddMap2Row(item, MTTitle, sheet.AddRow())
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

func appendLOHs(excel *xlsxUtil.File, lohs, lohSheetName string, sampleList []string) {
	for i, loh := range strings.Split(lohs, ",") {
		var sampleID = strconv.Itoa(i)
		if i < len(sampleList) {
			sampleID = sampleList[i]
		}
		excel.AppendSheet(*xlsxUtil.OpenFile(loh).File.Sheet[lohSheetName], sampleID+"-loh")
	}
}

func addExon() {
	if *exon != "" {
		anno.LoadGeneTrans(anno.GetPath("geneSymbol.transcript", dbPath, defaultConfig))
		var paths []string
		for _, path := range strings.Split(*exon, ",") {
			if osUtil.FileExists(path) {
				paths = append(paths, path)
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		addCnv2Sheet(
			tier1Xlsx.Sheet["exon_cnv"], exonCnvTitle, paths, sampleMap,
			false, *cnvFilter, stats, "exonCNV",
		)
		logTime("add exon cnv")
	}
}

func addLarge() {
	if *large != "" {
		var paths []string
		var pathMap = make(map[string]bool)
		for _, path := range strings.Split(*large, ",") {
			if osUtil.FileExists(path) {
				pathMap[path] = true
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		for path := range pathMap {
			paths = append(paths, path)
		}
		addCnv2Sheet(
			tier1Xlsx.Sheet["large_cnv"], largeCnvTitle, paths, sampleMap,
			*cnvFilter, false, stats, "largeCNV",
		)
		logTime("add large cnv")
	}
	if *smn != "" {
		var paths []string
		for _, path := range strings.Split(*smn, ",") {
			if osUtil.FileExists(path) {
				paths = append(paths, path)
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		addSmnResult(tier1Xlsx.Sheet["large_cnv"], largeCnvTitle, paths, sampleMap)
		logTime("add SMN1 result")
	}
}

func addExtra() {
	// extra sheet
	if *extra != "" {
		extraArray := strings.Split(*extra, ",")
		extraSheetArray := strings.Split(*extraSheetName, ",")
		if len(extraArray) != len(extraSheetArray) {
			log.Printf(
				"extra files not equal length to sheetnames:%+vvs.%+v",
				extraArray,
				extraSheetArray,
			)
		} else {
			for i := range extraArray {
				xlsxUtil.AddSlice2Sheet(
					textUtil.File2Slice(extraArray[i], "\t"),
					xlsxUtil.AddSheet(tier1Xlsx, extraSheetArray[i]),
				)
			}
		}
	}
}

func addQC() {
	parseQC()
	// QC Sheet
	updateQC(stats, qualitys[0])
	addQCSheet(tier1Xlsx, "quality", qualityColumn, qualitys)
	logTime("add qc")
}

func addLOH() {
	// append loh sheet
	if *loh != "" {
		appendLOHs(&xlsxUtil.File{File: tier1Xlsx}, *loh, *lohSheet, sampleList)
	}
}

func fillSheet() {
	parseList()
	addExon()
	addLarge()
	addExtra()
	addFamInfoSheet(tier1Xlsx, "fam_info", sampleList)
	addFV()
	addLOH()
	// need stats
	addQC()
}
func saveExcel() {
	if *save {
		if *wgs && *snv != "" {
			simpleUtil.CheckErr(wgsXlsx.Save(*prefix + ".WGS.xlsx"))
			logTime("save WGS")
		}

		// Tier1 excel
		tagStr := ""
		if *tag != "" {
			tagStr = textUtil.File2Array(*tag)[0]
		}
		var tier1Output string
		if isSMN1 && !*wesim {
			tier1Output = *prefix + ".Tier1" + tagStr + ".SMN1.xlsx"
		} else {
			tier1Output = *prefix + ".Tier1" + tagStr + ".xlsx"
		}
		simpleUtil.CheckErr(tier1Xlsx.Save(tier1Output))
		logTime("save Tier1")

		if *snv != "" {
			// Tier2 excel
			simpleUtil.CheckErr(tier2.save(), "Tier2 save fail")
			logTime("save Tier2")

			// Tier3 excel
			if !*noTier3 {
				simpleUtil.CheckErr(tier3Xlsx.Save(*prefix + ".Tier3.xlsx"))
				logTime("save Tier3")
			}
		}
	}
	logTime0("total work")
}
