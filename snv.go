package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/anno2xlsx/v2/hgvs"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	simple_util "github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v3"
)

// add filter_variants
func addFV() {
	// anno
	if *snv != "" {
		var data = loadData()

		stats["Total"] = len(data)

		cycle1(data)
		delDupVar(data)
		cycle2(data)
		// WGS
		wgsCycle(data)

		logTime("update info")
	}
}

func cycle1(data []map[string]string) {
	for _, item := range data {
		annotate1(item)
		cycle1Count++
		if cycle1Count%1000 == 0 {
			log.Printf("cycle1 progress %d/%d", cycle1Count, len(data))
		}
	}
	logTierStats(stats)
	logTime("load snv cycle 1")
}

func delDupVar(data []map[string]string) {
	for _, item := range data {
		if item["Tier"] == "Tier1" {
			var key = strings.Join([]string{item["#Chr"], item["Start"], item["Stop"], item["Ref"], item["Call"], item["Gene Symbol"]}, "\t")
			if countVar[key] > 1 {
				duplicateVar[key] = append(duplicateVar[key], item)
			}
		}
	}
	for key, items := range duplicateVar {
		var maxFunc = 0
		for _, item := range items {
			var score = anno.FuncInfo[item["Function"]]
			if score > maxFunc {
				maxFunc = score
			}
		}
		var minTrans = 0
		for _, item := range items {
			var transcript = item["Transcript"]
			var score = anno.FuncInfo[item["Function"]]
			if score < maxFunc {
				item["delete"] = "Y"
				deleteVar[key+"\t"+transcript] = true
				countVar[key]--
			} else {
				if transcriptLevel[transcript] > 0 {
					if minTrans == 0 {
						minTrans = transcriptLevel[transcript]
					}
					if minTrans > transcriptLevel[transcript] {
						minTrans = transcriptLevel[transcript]
					}
				}
			}
		}
		if minTrans > 0 {
			for _, item := range items {
				var transcript = item["Transcript"]
				if item["delete"] != "Y" && transcriptLevel[transcript] != minTrans {
					item["delete"] = "Y"
					deleteVar[key+"\t"+transcript] = true
					countVar[key]--
					log.Printf("Delete:%s\t%s\n", key, transcript)
				}
			}
		}
	}
}

func cycle2(data []map[string]string) {
	for _, item := range data {
		if item["Tier"] == "Tier1" {
			anno.InheritCheck(item, inheritDb)
		}
	}
	for _, item := range data {
		if item["Tier"] == "Tier1" {
			var key = strings.Join([]string{item["#Chr"], item["Start"], item["Stop"], item["Ref"], item["Call"], item["Gene Symbol"], item["Transcript"]}, "\t")
			if !deleteVar[key] {
				tier1Count++
				annotate2(item)
				// Tier1 Sheet
				xlsxUtil.AddMap2Row(item, filterVariantsTitle, tier1Xlsx.Sheet["filter_variants"].AddRow())
				if !*wgs {
					addTier2Row(tier2, item)
				}
			} else {
				if *wgs {
					tier1Db[item["MutationName"]] = true
					tier1GeneList[item["Gene Symbol"]] = true
				}
			}
		}
		// add to tier3
		if !*noTier3 {
			xlsxUtil.AddMap2Row(item, tier3Titles, tier3Sheet.AddRow())
		}
		cycle2Count++
		if cycle2Count%1000 == 0 {
			log.Printf("cycle1 progress %d/%d", cycle2Count, len(data))
		}
	}
	log.Printf("Tier1 Count : %d\n", tier1Count)
	if *wesim {
		simpleUtil.CheckErr(resultFile.Close())
	}
	logTime("load snv cycle 2")
}

func wgsCycle(data []map[string]string) {
	if *wgs {
		wgsXlsx = xlsx.NewFile()
		// MT sheet
		var MTSheet = xlsxUtil.AddSheet(wgsXlsx, "MT")
		xlsxUtil.AddArray2Row(MTTitle, MTSheet.AddRow())
		// intron sheet
		var intronSheet = xlsxUtil.AddSheet(wgsXlsx, "intron")
		xlsxUtil.AddArray2Row(filterVariantsTitle, intronSheet.AddRow())

		TIPdbPath := anno.GetPath("TIPdb", dbPath, defaultConfig)
		jsonUtil.JsonFile2Data(TIPdbPath, &TIPdb)
		MTdiseasePath := anno.GetPath("MTdisease", dbPath, defaultConfig)
		jsonUtil.JsonFile2Data(MTdiseasePath, &MTdisease)
		MTAFdbPath := anno.GetPath("MTAFdb", dbPath, defaultConfig)
		jsonUtil.JsonFile2Data(MTAFdbPath, &MTAFdb)

		inheritDb = make(map[string]map[string]int)
		for _, item := range data {
			anno.AddTier(item, stats, geneList, specVarDb, *trio, true, *allGene, anno.AFlist)
			// 遗传相符
			// only for Tier1
			if item["Tier"] == "Tier1" {
				anno.InheritCheck(item, inheritDb)
			}
		}
		logTime("load snv cycle 3")
		for _, item := range data {
			annotate4(item)

			if *wgs && isMT.MatchString(item["#Chr"]) {
				addMTRow(MTSheet, item)
			}
			if tier1GeneList[item["Gene Symbol"]] && item["Tier"] == "Tier1" {
				addTier2Row(tier2, item)

				if item["Function"] == "intron" && !tier1Db[item["MutationName"]] {
					intronRow := intronSheet.AddRow()
					for _, str := range filterVariantsTitle {
						intronRow.AddCell().SetString(item[str])
					}
				}
			}
		}
		logTime("load snv cycle 4")
	}
}

func annotate1(item map[string]string) {
	// score to prediction
	anno.Score2Pred(item)

	// update Function
	anno.UpdateFunction(item)

	// update FuncRegion
	anno.UpdateFuncRegion(item)

	var gene = item["Gene Symbol"]
	var id, ok = gene2id[gene]
	if !ok {
		if gene != "-" && gene != "." {
			log.Fatalf("can not find gene id of [%s]\n", gene)
		}
	}
	item["geneID"] = id

	chpo.Anno(item, id)
	// 基因-疾病
	diseaseDb.Anno(item, id)
	// 突变频谱
	spectrumDb.Anno(item, id)

	item["Gene"] = item["Omim Gene"]
	item["OMIM"] = item["OMIM_Phenotype_ID"]

	//anno.ParseSpliceAI(item)

	// ues acmg of go
	if *acmg {
		acmg2015.AddEvidences(item)
	}
	item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)

	anno.UpdateSnv(item, *gender)

	// 引物设计
	item["exonCount"] = exonCount[item["Transcript"]]
	item["引物设计"] = anno.PrimerDesign(item)

	// 变异来源
	if *trio2 {
		item["变异来源"] = anno.InheritFrom2(item, sampleList)
	}
	if *trio {
		item["变异来源"] = anno.InheritFrom(item, sampleList)
	}

	anno.AddTier(item, stats, geneList, specVarDb, *trio, false, *allGene, anno.AFlist)
	if *mt && isMT.MatchString(item["#Chr"]) {
		item["Tier"] = "Tier1"
		item["MTmut"] = getMhgvs(item)
		mtGnomAD.Anno(item, item["MTmut"])
	}

	if item["Tier"] == "Tier1" || item["Tier"] == "Tier2" {
		anno.UpdateSnvTier1(item)
		if *ifRedis {
			anno.UpdateRedis(item, redisDb, *seqType)
		}

		anno.UpdateAutoRule(item)
		anno.UpdateManualRule(item)
	}

	// only for Tier1
	if item["Tier"] == "Tier1" {
		// 遗传相符
		annotate1Tier1(item)
	}

	stats[item["#Chr"]]++
	if isHom.MatchString(item["Zygosity"]) {
		stats["Hom"]++
		stats["Hom:"+item["#Chr"]]++
	}
	stats[item["VarType"]]++
}

func getMhgvs(item map[string]string) string {
	var pos = simpleUtil.HandleError(strconv.Atoi(item["Start"])).(int) + 1
	var ref = item["Ref"]
	var alt = item["Call"]
	if ref == "." {
		ref = ""
	}
	if alt == "." {
		alt = ""
	}
	return hgvs.GetMhgvs(pos, []byte(ref), []byte(alt))
}

func annotate1Tier1(item map[string]string) {
	tier1GeneList[item["Gene Symbol"]] = true
	if anno.FuncInfo[item["Function"]] >= 3 {
		stats["Tier1LoF"]++
	}
	if isHom.MatchString(item["Zygosity"]) {
		stats["Tier1Hom"]++
	}
	stats["Tier1"+item["VarType"]]++

	if *academic {
		revel.anno(item)
	}

	var key = strings.Join([]string{item["#Chr"], item["Start"], item["Stop"], item["Ref"], item["Call"], item["Gene Symbol"]}, "\t")
	countVar[key]++
}

func annotate2(item map[string]string) {
	// 遗传相符
	item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
	item["遗传相符-经典trio"] = anno.InheritCoincide(item, inheritDb, true)
	item["遗传相符-非经典trio"] = anno.InheritCoincide(item, inheritDb, false)
	if item["遗传相符"] == "相符" {
		stats["遗传相符"]++
	}
	// familyTag
	if *trio || *trio2 {
		item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
	} else if *couple {
		item["familyTag"] = anno.FamilyTag(item, inheritDb, "couple")
	} else {
		item["familyTag"] = anno.FamilyTag(item, inheritDb, "single")
	}
	item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio, *trio2)

	anno.Format(item)

	// WESIM
	annotate2IM(item)
	var key = strings.Join([]string{item["#Chr"], item["Start"], item["Stop"], item["Ref"], item["Call"], item["Gene Symbol"]}, "\t")
	if countVar[key] > 1 {
		log.Printf("Duplicate:%s\t%s\t%s\t%s\n", key, item["Transcript"], item["cHGVS"], item["Function"])
		duplicateVar[key] = append(duplicateVar[key], item)
	}
}

func annotate2IM(item map[string]string) {
	if *wesim {
		var zygo = item["Zygosity"]
		if acmg59Gene[item["Gene Symbol"]] {
			item["IsACMG59"] = "Y"
		} else {
			item["IsACMG59"] = "N"
		}
		if *trio {
			zygosity := strings.Split(item["Zygosity"], ";")
			zygosity = append(zygosity, "NA", "NA")
			item["Zygosity"] = zygosity[0]
			item["Genotype of Family Member 1"] = zygosity[1]
			item["Genotype of Family Member 2"] = zygosity[2]
		}
		var resultArray []string
		for _, key := range resultColumn {
			resultArray = append(resultArray, item[key])
		}
		_, err = fmt.Fprintln(resultFile, strings.Join(resultArray, "\t"))
		simpleUtil.CheckErr(err)
		item["Zygosity"] = zygo
	}
}

func annotate4(item map[string]string) {
	if item["Tier"] == "Tier1" {
		// 遗传相符
		item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
		item["遗传相符-经典trio"] = anno.InheritCoincide(item, inheritDb, true)
		item["遗传相符-非经典trio"] = anno.InheritCoincide(item, inheritDb, false)
		if item["遗传相符"] == "相符" {
			stats["遗传相符"]++
		}
		// familyTag
		if *trio {
			item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
		}
		item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio, *trio2)
	}
}

func loadData() (data []map[string]string) {
	for _, f := range snvs {
		if isGz.MatchString(f) {
			d, _ := simple_util.Gz2MapArray(f, "\t", isComment)
			data = append(data, d...)
		} else {
			d, _ := simple_util.File2MapArray(f, "\t", isComment)
			data = append(data, d...)
		}
	}
	logTime("load anno file")
	return
}
