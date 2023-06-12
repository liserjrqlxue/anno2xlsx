package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	"github.com/tealeg/xlsx/v3"
)

// add filter_variants
func addFV() {
	anno.HomFixRatioThreshold = homFixRatioThreshold
	// anno
	if *snv != "" {
		var data = loadData()

		stats["Total"] = len(data)

		cycle1(data)
		delDupVar(data)
		cycle2(data)
		// WGS
		if *wgs {
			wgsCycle(data)
		}

		logTime("update info")
	}
}

func cycle1(data []map[string]string) {
	for _, item := range data {
		annotate1(item)
		cycle1Count++
		if cycle1Count%20000 == 0 {
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
				continue
			}
			if transcriptLevel[transcript] > 0 {
				if minTrans == 0 {
					minTrans = transcriptLevel[transcript]
				}
				if minTrans > transcriptLevel[transcript] {
					minTrans = transcriptLevel[transcript]
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

var (
	isBLB        = regexp.MustCompile(`B|LB`)
	isClinVarBLB = regexp.MustCompile(`Benign|Likely_benign`)
	isHLA        = regexp.MustCompile(`^HLA-`)
)

var nonCodeFunction = map[string]bool{
	"utr-3":        true,
	"utr-5":        true,
	"intron":       true,
	"promoter":     true,
	"ncRNA":        true,
	"coding-synon": true,
	"splice+10":    true,
	"splice-10":    true,
	"splice+20":    true,
	"splice-20":    true,
}

func tier1Filter(item map[string]string) bool {
	if isHLA.MatchString(item["Gene Symbol"]) {
		return false
	}
	if item["筛选标签"] == "" {
		if anno.CheckAF(item, []string{"A.Ratio"}, 0.1) || nonCodeFunction[item["Function"]] {
			return false
		}
		if isBLB.MatchString(item["自动化判断"]) && isClinVarBLB.MatchString(item["ClinVar Significance"]) {
			return false
		}
	}
	return true
}

func cycle2(data []map[string]string) {
	for _, item := range data {
		if item["Tier"] == "Tier1" {
			anno.InheritCheck(item, inheritDb)
		}
	}
	for _, item := range data {
		item["ClinVar Significance"] = anno.AddClnsigConf(item)
		if item["Tier"] == "Tier1" {
			tier1Db[item["MutationName"]] = true
			var key = strings.Join([]string{item["#Chr"], item["Start"], item["Stop"], item["Ref"], item["Call"], item["Gene Symbol"], item["Transcript"]}, "\t")
			if !deleteVar[key] && tier1Filter(item) {
				tier1Count++
				annotate2(item)
				// Tier1 Sheet
				xlsxUtil.AddMap2Row(item, filterVariantsTitle, tier1Xlsx.Sheet["filter_variants"].AddRow())
				tier1Data = append(tier1Data, selectMap(item, filterVariantsTitle))
				if !*wgs {
					addTier2Row(tier2, item)
				}
			}
		}
		// add to tier3
		if outputTier3 {
			SteamWriterSetStringMap2Row(tier3SW, 1, tier3RowID, item, tier3Titles)
			tier3RowID++
		}
		cycle2Count++
		if cycle2Count%50000 == 0 {
			log.Printf("cycle2 progress %d/%d", cycle2Count, len(data))
		}
	}
	log.Printf("Tier1 Count : %d\n", tier1Count)
	if *wesim {
		simpleUtil.CheckErr(resultFile.Close())
	}
	logTime("load snv cycle 2")
}

func wgsCycle(data []map[string]string) {
	wgsXlsx = xlsx.NewFile()
	// MT sheet
	var MTSheet = xlsxUtil.AddSheet(wgsXlsx, "MT")
	xlsxUtil.AddArray2Row(MTTitle, MTSheet.AddRow())
	// intron sheet
	var intronSheet = xlsxUtil.AddSheet(wgsXlsx, "intron")
	xlsxUtil.AddArray2Row(filterVariantsTitle, intronSheet.AddRow())

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
	var extraIntronCount = 0
	for _, item := range data {

		if item["Tier"] == "Tier1" {
			annotate4(item)
			addTier2Row(tier2, item)

			if item["Function"] != "no-change" && !tier1Db[item["MutationName"]] {
				extraIntronCount++
				intronRow := intronSheet.AddRow()
				for _, str := range filterVariantsTitle {
					intronRow.AddCell().SetString(item[str])
				}
			}
		}

		if isMT.MatchString(item["#Chr"]) {
			xlsxUtil.AddMap2Row(item, MTTitle, MTSheet.AddRow())
		}
	}
	log.Printf("add %d extra intron variant for wgs", extraIntronCount)
	logTime("load snv cycle 4")
}

func annotate1(item map[string]string) {
	// inhouse_AF -> frequency
	item["frequency"] = item["inhouse_AF"]
	// 历史验证假阳次数
	item["历史验证假阳次数"] = fpDb[item["Transcript"]+":"+strings.Replace(item["cHGVS"], " ", "", -1)]["重复数"]

	// score to prediction
	anno.Score2Pred(item)

	// update Function
	anno.UpdateFunction(item)

	// update FuncRegion
	anno.UpdateFuncRegion(item)

	// gene symbol -> geneID
	var gene = item["Gene Symbol"]
	var id, ok = gene2id[gene]
	if !ok {
		if gene != "-" && gene != "." {
			log.Fatalf("can not find gene id of [%s]\n", gene)
		}
	}
	item["geneID"] = id

	// CHPO
	chpo.Anno(item, id)
	// 基因-疾病
	diseaseDb.Anno(item, id)
	// 突变频谱
	spectrumDb.Anno(item, id)

	var multiKeys = anno.GetKeys(item["Transcript"], item["cHGVS"])
	// ACMG SF
	acmgSecondaryFindingDb.AnnoMultiKey(item, multiKeys)
	// 孕前数据库
	prePregnancyDb.AnnoMultiKey(item, multiKeys)
	// 新生儿数据库
	newBornDb.AnnoMultiKey(item, multiKeys)
	// 耳聋数据库
	hearingLossDb.AnnoMultiKey(item, multiKeys)
	// PHGDTag
	var phgdTag []string
	for _, db := range phgdTagDb {
		var v1 = item[db[1]]
		if v1 != "" {
			v1 = db[0] + ":" + v1
			var v2 = item[db[2]]
			if v2 != "" {
				v1 += ":" + v2
			}
			phgdTag = append(phgdTag, v1)
		}
	}
	item[phgdTagKey] = strings.Join(phgdTag, phgdTagSep)

	item["Gene"] = item["Omim Gene"]
	item["OMIM"] = item["OMIM_Phenotype_ID"]

	//anno.ParseSpliceAI(item)

	// ues acmg of go
	if *acmg {
		if item["cHGVS_org"] == "" {
			item["cHGVS_org"] = item["cHGVS"]
		}
		acmg2015.AddEvidences(item)
		item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)
	}

	anno.UpdateSnv(item, *gender)

	// 引物设计
	item["exonCount"] = exonCount[item["Transcript"]]
	item["引物设计"] = anno.PrimerDesign(item)

	// flank + HGVSc
	if item["HGVSc"] != "" {
		item["flank"] += " " + item["HGVSc"]
	}

	// 变异来源
	if *trio2 {
		item["变异来源"] = anno.InheritFrom2(item, sampleList)
	}
	if *trio {
		item["变异来源"] = anno.InheritFrom(item, sampleList)
	}

	anno.AddTier(item, stats, geneList, specVarDb, *trio, false, *allGene, anno.AFlist)
	if *allTier1 {
		item["Tier"] = "Tier1"
	}

	if item["Tier"] == "Tier1" || item["Tier"] == "Tier2" {
		if *ifRedis {
			anno.UpdateRedis(item, redisDb, *seqType)
		}
		anno.UpdateSnvTier1(item)

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

func annotate1Tier1(item map[string]string) {
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
	var familyTag = "single"
	if *trio || *trio2 {
		familyTag = "trio"
	} else if *couple {
		familyTag = "couple"
	}
	item["familyTag"] = anno.FamilyTag(item, inheritDb, familyTag)
	item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio, *trio2)

	anno.Format(item)

	// WESIM
	if *wesim {
		annotate2IM(item)
	}
}

func annotate2IM(item map[string]string) {
	var zygo = item["Zygosity"]
	if acmgSFGene[item["Gene Symbol"]] {
		item["IsACMG59"] = "Y"
	} else {
		item["IsACMG59"] = "N"
	}

	var inheritance = strings.Split(item["ModeInheritance"], "\n")
	var disease []string
	if isEN {
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
	item["DiseaseName/ModeInheritance"] = strings.Join(inheritance, "<br>")

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

func annotate4(item map[string]string) {
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

func loadData() (data []map[string]string) {
	for _, f := range snvs {
		if isGz.MatchString(f) {
			d, _ := textUtil.Gz2MapArray(f, "\t", isComment)
			data = append(data, d...)
		} else {
			d, _ := textUtil.File2MapArray(f, "\t", isComment)
			data = append(data, d...)
		}
	}
	logTime("load anno file")
	return
}
