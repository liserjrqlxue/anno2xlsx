package main

import (
	"flag"
	"fmt"
	"github.com/brentp/bix"
	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/acmg2015/evidence"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
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

// version
var buildStamp, gitHash, goVersion string

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
		"",
		"database of 突变频谱",
	)
	geneDiseaseDbFile = flag.String(
		"geneDisease",
		"",
		"database of 基因-疾病数据库",
	)
	geneDiseaseDbTitle = flag.String(
		"geneDiseaseTitle",
		"",
		"Title map of 基因-疾病数据库",
	)
	specVarList = flag.String(
		"specVarList",
		"",
		"特殊位点库",
	)
	transInfo = flag.String(
		"transInfo",
		"",
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
		"",
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
	wgs = flag.Bool(
		"wgs",
		false,
		"if anno wgs",
	)
	config = flag.String(
		"config",
		filepath.Join(exPath, "etc", "config.json"),
		"default config file, config will be overwrite by flag",
	)
	wesim = flag.Bool(
		"wesim",
		false,
		"if wesim, output result.tsv",
	)
	acmg = flag.Bool(
		"acmg",
		false,
		"if use new ACMG, fix PVS1, PS1,PS4, PM1,PM2,PM4,PM5 PP2,PP3, BA1, BS1,BS2, BP1,BP3,BP4,BP7",
	)
	cpuprofile = flag.String(
		"cpuprofile",
		"",
		"cpu profile",
	)
	memprofile = flag.String(
		"memprofile",
		"",
		"mem profile",
	)
	noTier3 = flag.Bool(
		"noTier3",
		false,
		"if not output Tier3.xlsx",
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

// WESIM
var resultColumn, qualityColumn []string
var resultFile, qcFile *os.File

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
var qualityKeyMap = make(map[string]string)

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
	isMT      = regexp.MustCompile(`MT|chrM`)
)

var redisDb *redis.Client

var isSMN1 bool

var snvs []string

// WGS
var WGSxlsx *xlsx.File
var TIPdb = make(map[string]Variant)
var MTdisease = make(map[string]Variant)
var MTAFdb = make(map[string]Variant)
var MTTitle []string

// ACMG
// PVS1
var LOFList map[string]int
var transcriptInfo map[string][]evidence.Region

// PS1 & PM5
var ClinVarMissense, ClinVarPHGVSlist, HGMDMissense, HGMDPHGVSlist, ClinVarAAPosList, HGMDAAPosList map[string]int

// PM1
var tbx *bix.Bix
var dbNSFPDomain, PfamDomain map[string]bool

// PP2
var ClinVarPP2GeneList, HgmdPP2GeneList map[string]float64

// BS2
var lateOnsetList map[string]int

// BP1
var ClinVarBP1GeneList, HgmdBP1GeneList map[string]float64

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	logVersion()
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
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

	if *logfile == "" {
		*logfile = *prefix + ".log"
	}
	logFile, err := os.Create(*logfile)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(logFile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Log file:%v \n", *logfile)
	logVersion()

	// parser etc/config.json
	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})

	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = getStrVal("redisServer", defaultConfig)
		}
		redisDb = redis.NewClient(&redis.Options{
			Addr: *redisAddr,
		})
		pong, err := redisDb.Ping().Result()
		log.Println("connect redis:", pong, err)
		if err != nil {
			*ifRedis = false
			log.Printf("Error connect redis[%+v], skip\n", err)
		}
	}

	if *acmg {
		// PVS1
		simple_util.JsonFile2Data(getPath("LOFList", defaultConfig), &LOFList)
		simple_util.JsonFile2Data(getPath("transcriptInfo", defaultConfig), &transcriptInfo)

		// PS1 & PM5
		simple_util.JsonFile2Data(getPath("ClinVarPathogenicMissense", defaultConfig), &ClinVarMissense)
		simple_util.JsonFile2Data(getPath("ClinVarPHGVSlist", defaultConfig), &ClinVarPHGVSlist)
		simple_util.JsonFile2Data(getPath("HGMDPathogenicMissense", defaultConfig), &HGMDMissense)
		simple_util.JsonFile2Data(getPath("HGMDPHGVSlist", defaultConfig), &HGMDPHGVSlist)
		simple_util.JsonFile2Data(getPath("ClinVarAAPosList", defaultConfig), &ClinVarAAPosList)
		simple_util.JsonFile2Data(getPath("HGMDAAPosList", defaultConfig), &HGMDAAPosList)

		// PM1
		simple_util.JsonFile2Data(getPath("PM1dbNSFPDomain", defaultConfig), &dbNSFPDomain)
		simple_util.JsonFile2Data(getPath("PM1PfamDomain", defaultConfig), &PfamDomain)
		tbx, err = bix.New(getPath("PathogenicLite", defaultConfig))
		simple_util.CheckErr(err, "load tabix")

		// PP2
		simple_util.JsonFile2Data(getPath("ClinVarPP2GeneList", defaultConfig), &ClinVarPP2GeneList)
		simple_util.JsonFile2Data(getPath("HgmdPP2GeneList", defaultConfig), &HgmdPP2GeneList)

		// BS2
		lateOnsetList = evidence.GetLateOnsetList(getPath("LateOnset", defaultConfig))

		// BP1
		simple_util.JsonFile2Data(getPath("ClinVarBP1GeneList", defaultConfig), &ClinVarBP1GeneList)
		simple_util.JsonFile2Data(getPath("HgmdBP1GeneList", defaultConfig), &HgmdBP1GeneList)
	}

	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = getPath("geneDiseaseDbFile", defaultConfig)
	}
	if *geneDiseaseDbTitle == "" {
		*geneDiseaseDbTitle = getPath("geneDiseaseDbTitle", defaultConfig)
	}
	if *geneDbFile == "" {
		*geneDbFile = getPath("geneDbFile", defaultConfig)
	}
	if *specVarList == "" {
		*specVarList = getPath("specVarList", defaultConfig)
	}
	if *transInfo == "" {
		*transInfo = getPath("transInfo", defaultConfig)
	}

	if *wesim {
		for _, key := range defaultConfig["resultColumn"].([]interface{}) {
			resultColumn = append(resultColumn, key.(string))
		}
		resultFile, err = os.Create(*prefix + ".result.tsv")
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(resultFile)
		fmt.Fprintln(resultFile, strings.Join(resultColumn, "\t"))

		for _, key := range defaultConfig["qualityColumn"].([]interface{}) {
			qualityColumn = append(qualityColumn, key.(string))
		}
		qcFile, err = os.Create(*prefix + ".qc.tsv")
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(qcFile)
		fmt.Fprintln(qcFile, strings.Join(qualityColumn, "\t"))
	}

	if *wgs {
		for _, key := range defaultConfig["MTTitle"].([]interface{}) {
			MTTitle = append(MTTitle, key.(string))
		}
		for k, v := range defaultConfig["qualityKeyMapWGS"].(map[string]interface{}) {
			qualityKeyMap[k] = v.(string)
		}
	} else {
		for k, v := range defaultConfig["qualityKeyMapWES"].(map[string]interface{}) {
			qualityKeyMap[k] = v.(string)
		}
	}

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
		loadQC(*qc, qualitys, *wgs)
		for _, quality := range qualitys {
			for k, v := range qualityKeyMap {
				quality[k] = quality[v]
			}
			if *wesim {
				var qcArray []string
				for _, key := range qualityColumn {
					qcArray = append(qcArray, quality[key])
				}
				fmt.Fprintln(qcFile, strings.Join(qcArray, "\t"))
			}
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load coverage.report")
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
	logTime(ts, step-1, step, "load mutation spectrum")

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
	logTime(ts, step-1, step, "load Gene-Disease DB")

	// 特殊位点库
	for _, key := range simple_util.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load Special mutation DB")

	// QC Sheet
	qcSheet := tier1.xlsx.Sheet["quality"]
	if qcSheet != nil {
		for _, row := range qcSheet.Rows {
			var qcArray []string
			key := row.Cells[0].Value
			qcArray = append(qcArray, key)
			for _, quality := range qualitys {
				row.AddCell().SetString(quality[key])
				qcArray = append(qcArray, quality[key])
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
		var step0 = step
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
			if *acmg {
				evidence.ComparePVS1(item, LOFList, transcriptInfo, tbx)
				item["PVS1"] = evidence.CheckPVS1(item, LOFList, transcriptInfo, tbx)
				item["PS1"] = evidence.CheckPS1(item, ClinVarMissense, ClinVarPHGVSlist, HGMDMissense, HGMDPHGVSlist)
				item["PM5"] = evidence.CheckPM5(item, ClinVarPHGVSlist, ClinVarAAPosList, HGMDPHGVSlist, HGMDAAPosList)
				item["PS4"] = evidence.CheckPS4(item)
				item["PM1"] = evidence.CheckPM1(item, dbNSFPDomain, PfamDomain, tbx)
				item["PM2"] = evidence.CheckPM2(item)
				item["PM4"] = evidence.CheckPM4(item)
				item["PP2"] = evidence.CheckPP2(item, ClinVarPP2GeneList, HgmdPP2GeneList)
				item["PP3"] = evidence.CheckPP3(item)
				item["BA1"] = evidence.CheckBA1(item) // BA1 更改条件，去除PVFD，新增ESP6500
				item["BS1"] = evidence.CheckBS1(item) // BS1 更改条件，去除PVFD，也没有对阈值1%进行修正
				item["BS2"] = evidence.CheckBS2(item, lateOnsetList)
				item["BP1"] = evidence.CheckBP1(item, ClinVarBP1GeneList, HgmdBP1GeneList)
				item["BP3"] = evidence.CheckBP3(item)
				item["BP4"] = evidence.CheckBP4(item) // BP4 更改条件，更严格了，非splice未考虑保守性
				item["BP7"] = evidence.CheckBP7(item) // BP 更改条件，更严格了，考虑PhyloP,以及无记录预测按不满足条件来做
			}

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

			anno.AddTier(item, stats, geneList, specVarDb, *trio, false)

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
		logTierStats(stats)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load snv cycle 1")
		for _, item := range data {
			if item["Tier"] == "Tier1" {
				// 遗传相符
				item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
				if item["遗传相符"] == "相符" {
					stats["遗传相符"]++
				}
				// familyTag
				if *trio {
					item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
				}
				item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio)

				anno.FloatFormat(item)

				// Tier1 Sheet
				tier1Row := tier1.sheet.AddRow()
				for _, str := range tier1.title {
					tier1Row.AddCell().SetString(item[str])
				}

				if !*wgs {
					addTier2Row(tier2, item)
				}

				// WESIM
				if *wesim {
					var resultArray []string
					for _, key := range resultColumn {
						resultArray = append(resultArray, item[key])
					}
					fmt.Fprintln(resultFile, strings.Join(resultArray, "\t"))
				}

				tier1GeneList[item["Gene Symbol"]] = true
			}

			// add to tier3
			if !*noTier3 {
				tier3Row := tier3.sheet.AddRow()
				for _, str := range tier3.title {
					tier3Row.AddCell().SetString(item[str])
				}
			}
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load snv cycle 2")

		// WGS
		if *wgs {
			WGSxlsx = xlsx.NewFile()
			// MT sheet
			MTsheet, err := WGSxlsx.AddSheet("MT")
			simple_util.CheckErr(err)
			rowMT := MTsheet.AddRow()
			for _, key := range MTTitle {
				rowMT.AddCell().SetString(key)
			}
			// intron sheet
			intronSheet, err := WGSxlsx.AddSheet("intron")
			simple_util.CheckErr(err)
			rowIntron := intronSheet.AddRow()
			for _, key := range tier1.title {
				rowIntron.AddCell().SetString(key)
			}

			TIPdbPath := getPath("TIPdb", defaultConfig)
			simple_util.JsonFile2Data(TIPdbPath, &TIPdb)
			MTdiseasePath := getPath("MTdisease", defaultConfig)
			simple_util.JsonFile2Data(MTdiseasePath, &MTdisease)
			MTAFdbPath := getPath("MTAFdb", defaultConfig)
			simple_util.JsonFile2Data(MTAFdbPath, &MTAFdb)

			inheritDb = make(map[string]map[string]int)
			for _, item := range data {
				anno.AddTier(item, stats, geneList, specVarDb, *trio, true)
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
				if item["Tier"] == "Tier1" {
					// 遗传相符
					item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
					if item["遗传相符"] == "相符" {
						stats["遗传相符"]++
					}
					// familyTag
					if *trio {
						item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
					}
					item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio)
				}
				if *wgs && isMT.MatchString(item["#Chr"]) {
					addMTRow(MTsheet, item)
				}
				if tier1GeneList[item["Gene Symbol"]] && item["Tier"] == "Tier1" {
					addTier2Row(tier2, item)

					if item["Function"] == "intron" {
						intronRow := intronSheet.AddRow()
						for _, str := range tier1.title {
							intronRow.AddCell().SetString(item[str])
						}
					}
				}
			}
			ts = append(ts, time.Now())
			step++
			logTime(ts, step-1, step, "load snv cycle 4")
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step0, step, "update info")
	}

	if *save {
		if *wgs && *snv != "" {
			simple_util.CheckErr(WGSxlsx.Save(*prefix + ".WGS.xlsx"))
			err = tier1.save()
			simple_util.CheckErr(err)
			ts = append(ts, time.Now())
			step++
			logTime(ts, step-1, step, "save WGS")
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

	if *save && *snv != "" && !*noTier3 {
		simple_util.CheckErr(tier3.save())
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier3")
	}
	logTime(ts, 0, step, "total work")

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer simple_util.DeferClose(f)
	}
}
