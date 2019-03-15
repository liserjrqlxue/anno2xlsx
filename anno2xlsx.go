package main

import (
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/acmg2015/evidence"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	snv = flag.String(
		"snv",
		"",
		"input snv anno txt",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output xlsx prefix.tier{1,2,3}.xlsx, default is same to -snv",
	)
	geneDbFile = flag.String(
		"geneDb",
		dbPath+"基因库-更新版基因特征谱-加动态突变-20190110.xlsx.Sheet1.json.aes",
		"database of 突变频谱",
	)
	geneDiseaseDbFile = flag.String(
		"geneDisease",
		dbPath+"全外基因基因集整理OMIM-20190122.xlsx.Database.json.aes",
		"database of 基因-疾病数据库",
	)
	specVarList = flag.String(
		"specVarList",
		dbPath+"spec.var.list",
		"特殊位点库",
	)
	transInfo = flag.String(
		"transInfo",
		dbPath+"trans.exonCount.json.new.json",
		"info of transcript",
	)
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
	list = flag.String(
		"list",
		"proband,father,mother",
		"sample list for family mode, comma as sep",
	)
	exon = flag.String(
		"exon",
		"",
		"exonCnv file path, only write samples in -list",
	)
	large = flag.String(
		"large",
		"",
		"largeCnv file path, only write sample in -list",
	)
	smn = flag.String(
		"smn",
		"",
		"smn result file path, require -large and only write sample in -list",
	)
	gender = flag.String(
		"gender",
		"NA",
		"gender of proband, if M then change Hom to Hemi in XY not PAR region",
	)
	qc = flag.String(
		"qc",
		"",
		"coverage.report file to fill quality sheet",
	)
	ifRedis = flag.Bool(
		"redis",
		false,
		"if use redis server",
	)
	redisAddr = flag.String(
		"redisAddr",
		"192.168.136.114:6380",
		"redis Addr Option",
	)
	seqType = flag.String(
		"seqType",
		"SEQ2000",
		"redis key:[SEQ2000|SEQ500|Hiseq]",
	)
)

// family list
var sampleList []string

// to-do add exon count info of transcript
var exonCount = make(map[string]string)

// 突变频谱
var geneDb = make(map[string]string)

// 基因-疾病
var geneList = make(map[string]bool)
var geneDiseaseDb = make(map[string]map[string]string)
var geneDiseaseDbColumn = map[string]string{
	"Gene/Locus":                 "Omim Gene",
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
var googleUrl = "https://www.google.com.hk/#q="

var quality = make(map[string]string)

var qualityKeyMap = map[string]string{
	"原始数据产出（Mb）":        "[Total] Raw Data(Mb)",
	"目标区长度（bp）":         "[Target] Len of region",
	"目标区覆盖度":            "[Target] Coverage (>0x)",
	"目标区平均深度（X）":        "[Target] Average depth(rmdup)",
	"目标区平均深度>4X位点所占比例":  "[Target] Coverage (>=4x)",
	"目标区平均深度>10X位点所占比例": "[Target] Coverage (>=10x)",
	"目标区平均深度>20X位点所占比例": "[Target] Coverage (>=20x)",
	"目标区平均深度>30X位点所占比例": "[Target] Coverage (>=30x)",
	"bam文件路径":           "bamPath",
}

var codeKey []byte

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

var redisDb *redis.Client

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	flag.Parse()
	if *snv == "" && *exon == "" && *large == "" && *smn == "" {
		flag.Usage()
		fmt.Println("\nshold have at least one input:-snv,-exon,-large,-smn")
		os.Exit(0)
	}
	if *prefix == "" {
		if *snv == "" {
			flag.Usage()
			fmt.Println("\nshold have -prefix for output")
			os.Exit(0)
		}
		*prefix = *snv
	}
	sampleList = strings.Split(*list, ",")
	var sampleMap = make(map[string]bool)
	for _, sample := range sampleList {
		sampleMap[sample] = true
	}
	quality["样本编号"] = sampleList[0]

	// load coverage.report
	if *qc != "" {
		loadQC(*qc, quality)
		for k, v := range qualityKeyMap {
			quality[k] = quality[v]
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load coverage.report")
	}

	// redis
	if *ifRedis {
		redisDb = redis.NewClient(&redis.Options{
			Addr: *redisAddr,
		})
		pong, err := redisDb.Ping().Result()
		fmt.Println("connect redis:", pong, err)
	}

	// load tier template
	var tiers = make(map[string]xlsxTemplate)
	var tierSheet = map[string]string{
		"Tier1": "filter_variants",
		"Tier2": "附表",
		"Tier3": "总表",
	}
	for _, key := range []string{"Tier1", "Tier3"} {
		var tier = xlsxTemplate{
			flag:      key,
			template:  templatePath + key + ".xlsx",
			sheetName: tierSheet[key],
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
	var tier2 = xlsxTemplate{
		flag:      "Tier2",
		sheetName: tierSheet["Tier2"],
		sheet:     tiers["Tier1"].xlsx.Sheet[tierSheet["Tier2"]],
	}
	if *trio {
		tier2.output = *prefix + ".Tier2.xlsx"
		tier2.xlsx, err = xlsx.OpenFile(templatePath + "Tier2.xlsx")
		simple_util.CheckErr(err)
	} else {
		tier2.xlsx = tiers["Tier1"].xlsx
	}
	tier2.sheet = tier2.xlsx.Sheet[tierSheet["Tier2"]]
	for _, cell := range tier2.sheet.Row(0).Cells {
		tier2.title = append(tier2.title, cell.String())
	}
	tiers["Tier2"] = tier2

	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load template")

	// exonCount
	exonCount = simple_util.JsonFile2Map(*transInfo)

	// 突变频谱
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDbExt := simple_util.Json2MapMap(simple_util.File2Decode(*geneDbFile, codeKey))
	for k := range geneDbExt {
		geneDb[k] = geneDbExt[k]["突变/致病多样性-补充/更正"]
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load 突变频谱")

	// 基因-疾病
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDiseaseDb = simple_util.Json2MapMap(simple_util.File2Decode(*geneDiseaseDbFile, codeKey))
	for key := range geneDiseaseDb {
		geneList[key] = true
	}
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
	if *snv != "" {
		var data []map[string]string
		if isGz.MatchString(*snv) {
			data, _ = simple_util.Gz2MapArray(*snv, "\t", isComment)
		} else {
			data, _ = simple_util.File2MapArray(*snv, "\t", isComment)
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load anno file")

		var stats = make(map[string]int)

		stats["Total"] = len(data)
		for _, item := range data {

			// score to prediction
			anno.Score2Pred(item)

			// ues acmg of go
			evidence.ComparePS4(item)
			evidence.ComparePM4(item)
			evidence.ComparePP3(item, true) // PP3 更改条件，更严格了，非splice未考虑保守性
			//evidence.CompareBA1(item,true) // BA1 更改条件，去除PVFD，新增ESP6500
			//evidence.CompareBS1(item,true) // BS1 更改条件，去除PVFD，也没有对阈值1%进行修正
			evidence.CompareBP3(item)
			evidence.CompareBP4(item, true) // BP4 更改条件，更严格了，非splice未考虑保守性
			evidence.CompareBP7(item, true) // BP 更改条件，更严格了，考虑PhyloP,以及无记录预测按不满足条件来做
			item["自动化判断"] = acmg2015.PredACMG2015(item)

			anno.UpdateSnv(item, *gender)

			gene := item["Gene Symbol"]
			// 突变频谱
			item["突变频谱"] = geneDb[gene]
			// 基因-疾病
			updateDisease(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
			item["Gene"] = item["Omim Gene"]

			// 引物设计
			item["exonCount"] = exonCount[item["Transcript"]]
			item["引物设计"] = anno.PrimerDesign(item)

			// 变异来源
			if *trio {
				item["变异来源"] = anno.InheritFrom(item, sampleList)
			}

			anno.AddTier(item, stats, geneList, specVarDb, *trio)

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
			if item["Tier"] == "Tier1" {
				anno.InheritCheck(item, inheritDb)
			}
		}
		for _, item := range data {
			// 遗传相符
			if item["Tier"] == "Tier1" {
				item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
				if *trio {
					item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
				}
				if item["遗传相符"] == "相符" {
					stats["遗传相符"]++
				}
			}

			// add to excel
			for flg := range tierSheet {
				if tier2xlsx[flg][item["Tier"]] {
					tierRow := tiers[flg].sheet.AddRow()
					for _, str := range tiers[flg].title {
						if str == "一键搜索链接" {
							cell := tierRow.AddCell()
							hyperlink := googleUrl + strings.Replace(item[str], "\"", "%22", -1) //  escape "
							if len(hyperlink) > 255 {
								cell.SetString(item[str])
							} else {
								cell.SetFormula("HYPERLINK(\"" + hyperlink + "\",\"" + strings.Replace(item[str], "\"", "\"\"", -1) + "\")")
							}
						} else {
							tierRow.AddCell().SetString(item[str])
						}
					}
				}
			}
		}

		logTierStats(stats)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "update info")
	} else {
		//tiers["Tier1"].xlsx.Sheet[tierSheet["Tier1"]].Hidden = true
		//tiers["Tier1"].xlsx.Sheet[tierSheet["Tier2"]].Hidden = true
	}

	// QC Sheet
	qcSheet := tiers["Tier1"].xlsx.Sheet["quality"]
	if qcSheet != nil {
		for _, row := range qcSheet.Rows {
			key := row.Cells[0].Value
			row.AddCell().SetString(quality[key])
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add qc")
	}
	//qcSheet.Cols[1].Width = 12

	if *exon != "" {
		addCnv2Sheet(tiers["Tier1"].xlsx.Sheet["exon_cnv"], *exon, sampleMap)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add exon cnv")
	} else {
		//tiers["Tier1"].xlsx.Sheet["exon_cnv"].Hidden = true
	}

	if *large != "" {
		addCnv2Sheet(tiers["Tier1"].xlsx.Sheet["large_cnv"], *large, sampleMap)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add large cnv")
	}
	if *smn != "" {
		addSmnResult(tiers["Tier1"].xlsx.Sheet["large_cnv"], *smn, sampleMap)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add SMN1 result")
	}
	if *large == "" && *smn == "" {
		//tiers["Tier1"].xlsx.Sheet["large_cnv"].Hidden = true
	}

	addFamInfoSheet(tiers["Tier1"].xlsx, "fam_info", sampleList)

	if *save {
		err = tiers["Tier1"].xlsx.Save(tiers["Tier1"].output)
		simple_util.CheckErr(err)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier1")
	}

	if *save && *trio {
		err = tiers["Tier2"].xlsx.Save(tiers["Tier2"].output)
		simple_util.CheckErr(err)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier1")
	}

	if *save && *snv != "" {
		err = tiers["Tier3"].xlsx.Save(tiers["Tier3"].output)
		simple_util.CheckErr(err)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier3")
	}
	logTime(ts, 0, step, "total work")
}

func logTierStats(stats map[string]int) {
	fmt.Printf("Total               Count : %7d\n", stats["Total"])
	if stats["Total"] == 0 {
		return
	}
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
	fmt.Printf("    遗传相符        Count : %7d\n", stats["遗传相符"])
	fmt.Printf("  Tier2             Count : %7d\n", stats["Tier2"])
	fmt.Printf("  Tier3             Count : %7d\n", stats["Tier3"])
}

func logTime(timeList []time.Time, step1, step2 int, message string) {
	trim := 3*8 - 1
	str := simple_util.FormatWidth(trim, message, ' ')
	fmt.Printf("%s\ttook %7.3fs to run.\n", str, timeList[step2].Sub(timeList[step1]).Seconds())
}

func addFamInfoSheet(excel *xlsx.File, sheetName string, sampleList []string) {
	sheet, err := excel.AddSheet(sheetName)
	simple_util.CheckErr(err)

	sheet.AddRow().AddCell().SetString("SampleID")

	for _, sample := range sampleList {
		sheet.AddRow().AddCell().SetString(sample)
	}
}

func addCnv2Sheet(sheet *xlsx.Sheet, path string, sampleMap map[string]bool) {
	cnvDb, _ := simple_util.File2MapArray(path, "\t", nil)

	// title
	var title []string
	for _, cell := range sheet.Row(0).Cells {
		title = append(title, cell.Value)
	}

	for _, item := range cnvDb {
		sample := item["Sample"]
		if sampleMap[sample] {
			gene := item["OMIM_Gene"]
			updateDiseaseMultiGene(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
			item["OMIM_Phenotype_ID"] = item["OMIM"]
			row := sheet.AddRow()
			for _, key := range title {
				row.AddCell().SetString(item[key])
			}
		}
	}
}

func addSmnResult(sheet *xlsx.Sheet, path string, sampleMap map[string]bool) {
	smnDb, _ := simple_util.File2MapArray(path, "\t", nil)

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
			item["Chr"] = "5"
			item["Start"] = "70241892"
			item["End"] = "70242003"
			item["Gene"] = "SMN1"
			item["OMIM_Gene"] = "SMN1"
			item["SMN1_result"] = item["SMN1_ex7_cn"]
			if item["SMN1_ex7_cn"] == "0" {
				item["SMN1_result"] = "Y"
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

func loadQC(file string, quality map[string]string) {
	report := simple_util.File2Array(file)
	for _, line := range report {
		if isSharp.MatchString(line) {
			if m := isBamPath.FindStringSubmatch(line); m != nil {
				quality["bamPath"] = m[1]
			}
		} else {
			m := strings.Split(line, "\t")
			quality[strings.TrimSpace(m[0])] = strings.TrimSpace(m[1])
		}
	}
}
