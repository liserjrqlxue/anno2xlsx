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
	"log"
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
	productID = flag.String(
		"product",
		"",
		"product ID",
	)
	snv = flag.String(
		"snv",
		"",
		"input snv anno txt, comma as sep",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output xlsx prefix.tier{1,2,3}.xlsx, default is same to first file of -snv",
	)
	logfile = flag.String(
		"log",
		"",
		"output log to log.log, default is prefix.log",
	)
	geneDbFile = flag.String(
		"geneDb",
		dbPath+"基因库-更新版基因特征谱-加动态突变-20190110.xlsx.Sheet1.json.aes",
		"database of 突变频谱",
	)
	geneDiseaseDbFile = flag.String(
		"geneDisease",
		"",
		"database of 基因-疾病数据库",
	)
	geneDiseaseDbTitle = flag.String(
		"geneDiseaseTitle",
		dbPath+"基因-疾病数据库.Title.json",
		"Title map of 基因-疾病数据库",
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
		"exonCnv files path, comma as sep, only write samples in -list",
	)
	large = flag.String(
		"large",
		"",
		"largeCnv file path, comma as sep, only write sample in -list",
	)
	smn = flag.String(
		"smn",
		"",
		"smn result file path, comma as sep, require -large and only write sample in -list",
	)
	gender = flag.String(
		"gender",
		"NA",
		"gender of sample list, comma as sep, if M then change Hom to Hemi in XY not PAR region",
	)
	qc = flag.String(
		"qc",
		"",
		"coverage.report file to fill quality sheet, comma as sep, same order with -list",
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
	cnvFilter = flag.Bool(
		"cnvFilter",
		false,
		"if filter cnv result",
	)
	annoMT = flag.Bool(
		"annoMT",
		false,
		"if anno MT",
	)
	wgs = flag.Bool(
		"wgs",
		false,
		"if anno wgs (include -annoMT)",
	)
	config = flag.String(
		"config",
		filepath.Join(exPath, "etc", "config.json"),
		"default config file, config will be overwrite by flag",
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
var geneDiseaseDbColumn = make(map[string]string)

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

func (xt *xlsxTemplate) save() error {
	return xt.xlsx.Save(xt.output)
}

var tierSheet = map[string]string{
	"Tier1": "filter_variants",
	"Tier3": "总表",
}

var tier1GeneList = make(map[string]bool)

var err error

func newXlsxTemplate(flag string) xlsxTemplate {
	var tier = xlsxTemplate{
		flag:      flag,
		template:  templatePath + flag + ".xlsx",
		sheetName: tierSheet[flag],
		output:    *prefix + "." + flag + ".xlsx",
	}
	tier.xlsx, err = xlsx.OpenFile(tier.template)
	simple_util.CheckErr(err)
	tier.sheet = tier.xlsx.Sheet[tier.sheetName]
	for _, cell := range tier.sheet.Row(0).Cells {
		tier.title = append(tier.title, cell.String())
	}
	return tier
}

var qualitys []map[string]string

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

// tier2
var isEnProduct = map[string]bool{
	"DX0700": true,
	"DX1335": true,
	"DX0458": false,
	"DX1616": false,
	"DX1515": false,
	"DX1617": false,
	"RC0029": false,
}

var transEN = map[string]string{
	"是":    "Yes",
	"否":    "No",
	"备注说明": "Note",
}

type templateInfo struct {
	cols      []string
	titles    [2][]string
	noteTitle [2]string
	note      [2][]string
}

var codeKey []byte

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

var redisDb *redis.Client

var isSMN1 bool

var snvs []string

// MT
var MTxlsx *xlsx.File
var MTsheet *xlsx.Sheet
var TIPdb = make(map[string]Variant)
var MTdisease = make(map[string]Variant)
var MTAFdb = make(map[string]Variant)

// WGS
var intronSheet *xlsx.Sheet

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
		} else {
			snvs = strings.Split(*snv, ",")
			*prefix = snvs[0]
		}
	}
	if *snv == "" {
		if *prefix == "" {
			flag.Usage()
			fmt.Println("\nshold have -prefix for output")
			os.Exit(0)
		}
	} else {
		snvs = strings.Split(*snv, ",")
		if *prefix == "" {
			*prefix = snvs[0]
		}
	}

	// parser etc/config.json
	defalutConfig := simple_util.JsonFile2Map(*config)
	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = filepath.Join(exPath, "db", defalutConfig["*geneDiseaseDbFile"])
	}

	if *wgs {
		*annoMT = true
	}

	if *logfile == "" {
		*logfile = *prefix + ".log"
	}
	logFile, err := os.Create(*logfile)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(logFile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Log file:%v \n", *logfile)

	sampleList = strings.Split(*list, ",")
	var sampleMap = make(map[string]bool)
	for _, sample := range sampleList {
		sampleMap[sample] = true
		quality := make(map[string]string)
		quality["样本编号"] = sample
		qualitys = append(qualitys, quality)
	}

	// load coverage.report
	if *qc != "" {
		loadQC(*qc, qualitys)
		for _, quality := range qualitys {
			for k, v := range qualityKeyMap {
				quality[k] = quality[v]
			}
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
	tier1 := newXlsxTemplate("Tier1")
	tier3 := newXlsxTemplate("Tier3")

	// tier2
	var tier2 = xlsxTemplate{
		flag:      "Tier2",
		sheetName: *productID + "_" + sampleList[0],
	}
	tier2.output = *prefix + "." + tier2.flag + ".xlsx"
	tier2.xlsx = xlsx.NewFile()

	var tier2TemplateInfo templateInfo
	tier2Template, err := xlsx.OpenFile(templatePath + "Tier2.xlsx")
	simple_util.CheckErr(err)
	tier2Infos, err := tier2Template.ToSlice()
	simple_util.CheckErr(err)
	for i, item := range tier2Infos[0] {
		if i > 0 {
			tier2TemplateInfo.cols = append(tier2TemplateInfo.cols, item[0])
			tier2TemplateInfo.titles[0] = append(tier2TemplateInfo.titles[0], item[1])
			tier2TemplateInfo.titles[1] = append(tier2TemplateInfo.titles[0], item[2])
		}
	}
	for _, item := range tier2Infos[1] {
		tier2TemplateInfo.note[0] = append(tier2TemplateInfo.note[0], item[0])
		tier2TemplateInfo.note[1] = append(tier2TemplateInfo.note[1], item[1])
	}

	tier2.sheet, err = tier2.xlsx.AddSheet(tier2.sheetName)
	simple_util.CheckErr(err)
	tier2row := tier2.sheet.AddRow()
	for i, col := range tier2TemplateInfo.cols {
		tier2.title = append(tier2.title, col)
		var title string
		if isEnProduct[*productID] {
			title = tier2TemplateInfo.titles[0][i]
		} else {
			title = tier2TemplateInfo.titles[1][i]
		}
		tier2row.AddCell().SetString(title)
	}

	var tier2NoteSheetName = "备注说明"
	var tier2Note []string
	if isEnProduct[*productID] {
		tier2NoteSheetName = transEN[tier2NoteSheetName]
		tier2Note = tier2TemplateInfo.note[1]
	} else {
		tier2Note = tier2TemplateInfo.note[0]
	}
	tier2NoteSheet, err := tier2.xlsx.AddSheet(tier2NoteSheetName)
	simple_util.CheckErr(err)
	for _, line := range tier2Note {
		tier2NoteSheet.AddRow().AddCell().SetString(line)
	}

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
	geneDiseaseDbTitleInfo := simple_util.JsonFile2MapMap(*geneDiseaseDbTitle)
	for key, item := range geneDiseaseDbTitleInfo {
		geneDiseaseDbColumn[key] = item["Key"]
	}
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

	// QC Sheet
	qcSheet := tier1.xlsx.Sheet["quality"]
	if qcSheet != nil {
		for _, row := range qcSheet.Rows {
			key := row.Cells[0].Value
			for _, quality := range qualitys {
				row.AddCell().SetString(quality[key])
			}
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add qc")
	}
	//qcSheet.Cols[1].Width = 12

	if *exon != "" {
		var paths []string
		for _, path := range strings.Split(*exon, ",") {
			if simple_util.FileExists(path) {
				paths = append(paths, path)
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		if *cnvFilter {
			addCnv2Sheet(tier1.xlsx.Sheet["exon_cnv"], paths, sampleMap, false, true)
		} else {
			addCnv2Sheet(tier1.xlsx.Sheet["exon_cnv"], paths, sampleMap, false, false)

		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add exon cnv")
	} else {
		//tiers["Tier1"].xlsx.Sheet["exon_cnv"].Hidden = true
	}

	if *large != "" {
		var paths []string
		for _, path := range strings.Split(*large, ",") {
			if simple_util.FileExists(path) {
				paths = append(paths, path)
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		if *cnvFilter {
			addCnv2Sheet(tier1.xlsx.Sheet["large_cnv"], paths, sampleMap, true, false)
		} else {
			addCnv2Sheet(tier1.xlsx.Sheet["large_cnv"], paths, sampleMap, false, false)
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add large cnv")
	}
	if *smn != "" {
		var paths []string
		for _, path := range strings.Split(*smn, ",") {
			if simple_util.FileExists(path) {
				paths = append(paths, path)
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		addSmnResult(tier1.xlsx.Sheet["large_cnv"], paths, sampleMap)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add SMN1 result")
	}
	if *large == "" && *smn == "" {
		//tiers["Tier1"].xlsx.Sheet["large_cnv"].Hidden = true
	}
	addFamInfoSheet(tier1.xlsx, "fam_info", sampleList)

	// anno
	if *snv != "" {
		var data []map[string]string
		for _, snv := range snvs {
			if isGz.MatchString(snv) {
				d, _ := simple_util.Gz2MapArray(snv, "\t", isComment)
				data = append(data, d...)
			} else {
				d, _ := simple_util.File2MapArray(snv, "\t", isComment)
				data = append(data, d...)
			}
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load anno file")

		var stats = make(map[string]int)

		stats["Total"] = len(data)
		for _, item := range data {

			// score to prediction
			anno.Score2Pred(item)

			// update Function
			anno.UpdateFunction(item)

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
			item["OMIM"] = item["OMIM_Phenotype_ID"]
			item["death age"] = item["hpo_cn"]

			// 引物设计
			item["exonCount"] = exonCount[item["Transcript"]]
			item["引物设计"] = anno.PrimerDesign(item)

			// 变异来源
			if *trio {
				item["变异来源"] = anno.InheritFrom(item, sampleList)
			}

			anno.AddTier(item, stats, geneList, specVarDb, *trio, *wgs)

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
				tier1GeneList[item["Gene Symbol"]] = true
			}
		}

		if *annoMT {
			MTxlsx = xlsx.NewFile()
			// MT sheet
			MTsheet, err = MTxlsx.AddSheet("MT")
			simple_util.CheckErr(err)
			rowMT := MTsheet.AddRow()
			for _, key := range MTTitle {
				rowMT.AddCell().SetString(key)
			}
			// intron sheet
			if *wgs {
				intronSheet, err = MTxlsx.AddSheet("intron")
				simple_util.CheckErr(err)
				rowIntron := intronSheet.AddRow()
				for _, key := range tier1.title {
					rowIntron.AddCell().SetString(key)
				}
			}

			simple_util.JsonFile2Data(dbPath+"线粒体数据库-fll-20190418.xlsx.MitoTIP-tRNA预测打分.json.db", &TIPdb)
			//fmt.Printf("%+v\n",TIPdb)
			simple_util.JsonFile2Data(dbPath+"线粒体数据库-fll-20190418.xlsx.疾病-位点库.json.db", &MTdisease)
			simple_util.JsonFile2Data(dbPath+"线粒体数据库-fll-20190418.xlsx.频率库.json.db", &MTAFdb)
		}
		var isMT = regexp.MustCompile(`MT|chrM`)
		for _, item := range data {
			// 遗传相符
			if item["Tier"] == "Tier1" {
				item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
				if item["遗传相符"] == "相符" {
					stats["遗传相符"]++
				}
				if *trio {
					item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
				}
				item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio)

				tier1Row := tier1.sheet.AddRow()
				for _, str := range tier1.title {
					tier1Row.AddCell().SetString(item[str])
				}

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
					default:
						tier2Row.AddCell().SetString(item[str])
					}

				}
			}

			if *annoMT && isMT.MatchString(item["#Chr"]) {
				addMTRow(MTsheet, item)
			}
			if *wgs && tier1GeneList[item["Gene Symbol"]] && item["Tier"] == "intron" {
				intronRow := intronSheet.AddRow()
				for _, str := range tier1.title {
					intronRow.AddCell().SetString(item[str])
				}
			}
			// add to tier3
			tier3Row := tier3.sheet.AddRow()
			for _, str := range tier3.title {
				tier3Row.AddCell().SetString(item[str])
			}
		}

		logTierStats(stats)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "update info")
	}

	if *save {
		if *annoMT {
			simple_util.CheckErr(MTxlsx.Save(*prefix + ".MT.xlsx"))
			err = tier1.save()
			simple_util.CheckErr(err)
			ts = append(ts, time.Now())
			step++
			logTime(ts, step-1, step, "save MT")
		}
	}

	if *save {
		if isSMN1 {
			tier1.output = *prefix + ".Tier1.SMN1.xlsx"
		}
		err = tier1.save()
		simple_util.CheckErr(err)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier1")
	}

	if *save {
		simple_util.CheckErr(tier2.save(), "Tier2 save fail")
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier2")
	}

	if *save && *snv != "" {
		simple_util.CheckErr(tier3.save())
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier3")
	}
	logTime(ts, 0, step, "total work")
}
