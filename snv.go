package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	"github.com/tealeg/xlsx/v3"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/anno2xlsx/v2/hgvs"
)

// add filter_variants
func addFV() {
	// anno
	if *snv != "" {
		var step0 = step
		var data = loadData()

		stats["Total"] = len(data)

		cycle1(data)
		cycle2(data)
		// WGS
		wgsCycle(data)

		ts = append(ts, time.Now())
		step++
		logTime(ts, step0, step, "update info")
	}
}

func cycle1(data []map[string]string) {
	for _, item := range data {
		annotate1(item)
	}
	logTierStats(stats)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load snv cycle 1")
}

func cycle2(data []map[string]string) {
	for _, item := range data {
		if item["Tier"] == "Tier1" {
			annotate2(item)

			// Tier1 Sheet
			xlsxUtil.AddMap2Row(item, filterVariantsTitle, tier1Xlsx.Sheet["filter_variants"].AddRow())

			if !*wgs {
				addTier2Row(tier2, item)
			} else {
				tier1Db[item["MutationName"]] = true
				tier1GeneList[item["Gene Symbol"]] = true
			}
		}
		// add to tier3
		if !*noTier3 {
			xlsxUtil.AddMap2Row(item, tier3Titles, tier3Sheet.AddRow())
		}
	}
	if *wesim {
		simpleUtil.CheckErr(resultFile.Close())
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load snv cycle 2")
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
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load snv cycle 3")
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
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load snv cycle 4")
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

	anno.UpdateSnv(item, *gender, *debug)

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

	// 遗传相符
	// only for Tier1
	annotate1Tier1(item)

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
	if item["Tier"] == "Tier1" {
		anno.InheritCheck(item, inheritDb)
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
	}
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
