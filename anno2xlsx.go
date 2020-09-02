package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/brentp/bix"
	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/acmg2015/evidence"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	simple_util "github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v2"

	"github.com/liserjrqlxue/anno2xlsx/anno"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "db")
	templatePath = filepath.Join(exPath, "template")
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
	trio2 = flag.Bool(
		"trio2",
		false,
		"if no standard trio mode but proband-father-mother",
	)
	couple = flag.Bool(
		"couple",
		false,
		"if couple mode",
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
	loh = flag.String(
		"loh",
		"",
		"loh result excel path, comma as sep, use sampleID in -list to create sheetName in order",
	)
	lohSheet = flag.String(
		"lohSheet",
		"LOH_annotation",
		"loh sheet name to append",
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
	kinship = flag.String(
		"kinship",
		"",
		"kinship result for trio only",
	)
	karyotype = flag.String(
		"karyotype",
		"",
		"karyotype files to fill quality sheet's 核型预测, comma as sep")
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
	filterVariants = flag.String(
		"filter_variants",
		filepath.Join(exPath, "etc", "Tier1.filter_variants.txt"),
		"overwrite template/tier1.xlsx filter_variants sheet columns' title",
	)
	exonCnv = flag.String(
		"exon_cnv",
		filepath.Join(exPath, "etc", "Tier1.exon_cnv.txt"),
		"overwrite template/tier1.xlsx exon_cnv sheet columns' title",
	)
	largeCnv = flag.String(
		"large_cnv",
		filepath.Join(exPath, "etc", "Tier1.large_cnv.txt"),
		"overwrite template/tier1.xlsx large_cnv sheet columns' title",
	)
	tier3Title = flag.String(
		"tier3Title",
		filepath.Join(exPath, "etc", "Tier3.总表.txt"),
		"overwrite template/tier3.xlsx 总表 sheet columns' title",
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
	autoPVS1 = flag.Bool(
		"autoPVS1",
		false,
		"if use autoPVS1",
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
	debug = flag.Bool(
		"debug",
		false,
		"if print some log",
	)
	allGene = flag.Bool(
		"allgene",
		false,
		"if not filter gene",
	)
	extra = flag.String(
		"extra",
		"",
		"extra file path to excel, comma as sep",
	)
	extraSheetName = flag.String(
		"extraSheet",
		"",
		"extra sheet name, comma as sep, same order with -extra",
	)
	tag = flag.String(
		"tag",
		"",
		"read tag from file, add to tier1 file name:[prefix].Tier1[tag].xlsx",
	)
	filterStat = flag.String(
		"filterStat",
		"",
		"filter.stat files to calculate reads QC, comma as sep",
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

var tier1GeneList = make(map[string]bool)

// WESIM
var resultColumn, qualityColumn []string
var resultFile, qcFile *os.File

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

var acmg59Gene = make(map[string]bool)

// WGS
var wgsXlsx *xlsx.File
var TIPdb = make(map[string]variant)
var MTdisease = make(map[string]variant)
var MTAFdb = make(map[string]variant)
var MTTitle []string
var tier1Db = make(map[string]bool)

// ACMG
// PVS1
var LOFList map[string]int
var transcriptInfo map[string][]evidence.Region

// PM1
var tbx *bix.Bix

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
		simpleUtil.CheckErr(pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
	}
	if *snv == "" && *exon == "" && *large == "" && *smn == "" && *loh == "" {
		flag.Usage()
		fmt.Println("\nshold have at least one input:-snv,-exon,-large,-smn,-loh")
		os.Exit(0)
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
	simpleUtil.CheckErr(err)
	defer simpleUtil.DeferClose(logFile)
	log.Printf("Log file         : %v\n", *logfile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Log file         : %v\n", *logfile)
	log.Printf("Git Commit Hash  : %s\n", gitHash)
	logVersion()

	// parser etc/config.json
	defaultConfig := jsonUtil.JsonFile2Interface(*config).(map[string]interface{})

	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = anno.GetStrVal("redisServer", defaultConfig)
		}
		redisDb = redis.NewClient(&redis.Options{
			Addr: *redisAddr,
		})
		pong, err := redisDb.Ping().Result()
		log.Printf("Connect [%s]:%s\n", redisDb.String(), pong)
		if err != nil {
			log.Fatalf("Error [%+v]\n", err)
		}
	}

	if *acmg {
		// PS1 PM5
		evidence.LoadPS1PM5(
			anno.GetPath("PS1PM5.MutationName.count", dbPath, defaultConfig),
			anno.GetPath("PS1PM5.pHGVS1.count", dbPath, defaultConfig),
			anno.GetPath("PS1PM5.AApos.count", dbPath, defaultConfig),
		)
		// PM1
		evidence.LoadPM1(
			anno.GetPath("PM1InterproDomain", dbPath, defaultConfig),
			anno.GetPath("PM1PfamIdDomain", dbPath, defaultConfig),
		)

		// PVS1
		if !*autoPVS1 {
			jsonUtil.JsonFile2Data(anno.GetPath("LOFList", dbPath, defaultConfig), &LOFList)
			jsonUtil.JsonFile2Data(anno.GetPath("transcriptInfo", dbPath, defaultConfig), &transcriptInfo)
		}

		// PM1
		tbx, err = bix.New(anno.GetPath("PathogenicLite", dbPath, defaultConfig))
		simpleUtil.CheckErr(err, "load tabix")

		// PP2
		evidence.LoadPP2(
			anno.GetPath("PP2GeneList", dbPath, defaultConfig),
		)

		// BS2
		evidence.LoadBS2(
			anno.GetPath("LateOnset", dbPath, defaultConfig),
		)

		// BP1
		evidence.LoadBP1(
			anno.GetPath("BP1GeneList", dbPath, defaultConfig),
		)

		// BA1
		evidence.LoadBA1(
			anno.GetPath("BA1ExceptionList", dbPath, defaultConfig),
		)
	}

	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = anno.GetPath("geneDiseaseDbFile", dbPath, defaultConfig)
	}
	if *geneDiseaseDbTitle == "" {
		*geneDiseaseDbTitle = anno.GetPath("geneDiseaseDbTitle", dbPath, defaultConfig)
	}
	if *geneDbFile == "" {
		*geneDbFile = anno.GetPath("geneDbFile", dbPath, defaultConfig)
	}
	geneDbKey := anno.GetStrVal("geneDbKey", defaultConfig)
	if *specVarList == "" {
		*specVarList = anno.GetPath("specVarList", dbPath, defaultConfig)
	}
	if *transInfo == "" {
		*transInfo = anno.GetPath("transInfo", dbPath, defaultConfig)
	}

	if *wgs {
		for _, key := range defaultConfig["qualityColumnWGS"].([]interface{}) {
			qualityColumn = append(qualityColumn, key.(string))
		}
	} else {
		for _, key := range defaultConfig["qualityColumn"].([]interface{}) {
			qualityColumn = append(qualityColumn, key.(string))
		}
	}

	if *wesim {
		acmg59GeneList := textUtil.File2Array(anno.GetPath("Acmg59Gene", dbPath, defaultConfig))
		for _, gene := range acmg59GeneList {
			acmg59Gene[gene] = true
		}

		for _, key := range defaultConfig["resultColumn"].([]interface{}) {
			resultColumn = append(resultColumn, key.(string))
		}
		if *trio {
			resultColumn = append(resultColumn, "Genotype of Family Member 1", "Genotype of Family Member 2")
		}
		resultFile, err = os.Create(*prefix + ".result.tsv")
		simpleUtil.CheckErr(err)
		defer simpleUtil.DeferClose(resultFile)
		_, err = fmt.Fprintln(resultFile, strings.Join(resultColumn, "\t"))
		simpleUtil.CheckErr(err)

		qcFile, err = os.Create(*prefix + ".qc.tsv")
		simpleUtil.CheckErr(err)
		defer simpleUtil.DeferClose(qcFile)
		_, err = fmt.Fprintln(qcFile, strings.Join(qualityColumn, "\t"))
		simpleUtil.CheckErr(err)
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

	var karyotypeMap = make(map[string]string)
	if *karyotype != "" {
		karyotypeMap, err = simple_util.Files2Map(*karyotype, "\t", true)
		simpleUtil.CheckErr(err)
	}
	// load coverage.report
	if *qc != "" {
		loadQC(*qc, *kinship, qualitys, *wgs)
		for _, quality := range qualitys {
			for k, v := range qualityKeyMap {
				quality[k] = quality[v]
			}
			quality["核型预测"] = karyotypeMap[quality["样本编号"]]
			if *wesim {
				var qcArray []string
				for _, key := range qualityColumn {
					qcArray = append(qcArray, quality[key])
				}
				_, err = fmt.Fprintln(qcFile, strings.Join(qcArray, "\t"))
				simpleUtil.CheckErr(err)
			}
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load coverage.report")
	}
	loadFilterStat(*filterStat, qualitys[0])

	// load tier template
	var tier1Xlsx = xlsx.NewFile()
	xlsxUtil.AddSheets(tier1Xlsx, []string{"filter_variants", "exon_cnv", "large_cnv"})
	filterVariantsTitle := addFile2Row(*filterVariants, tier1Xlsx.Sheet["filter_variants"].AddRow())
	exonCnvTitle := addFile2Row(*exonCnv, tier1Xlsx.Sheet["exon_cnv"].AddRow())
	largeCNVTitle := addFile2Row(*largeCnv, tier1Xlsx.Sheet["large_cnv"].AddRow())

	// create Tier3.xlsx
	var tier3Xlsx = xlsx.NewFile()
	var tier3Sheet = xlsxUtil.AddSheet(tier3Xlsx, "总表")
	var tier3Titles []string
	if !*noTier3 {
		tier3Titles = addFile2Row(*tier3Title, tier3Sheet.AddRow())
	}

	// tier2
	var tier2 = xlsxTemplate{
		flag:      "Tier2",
		sheetName: *productID + "_" + sampleList[0],
	}
	tier2.output = *prefix + "." + tier2.flag + ".xlsx"
	tier2.xlsx = xlsx.NewFile()

	var tier2TemplateInfo templateInfo
	tier2Template, err := xlsx.OpenFile(filepath.Join(templatePath, "Tier2.xlsx"))
	simpleUtil.CheckErr(err)
	tier2Infos, err := tier2Template.ToSlice()
	simpleUtil.CheckErr(err)
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
	simpleUtil.CheckErr(err)
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
	simpleUtil.CheckErr(err)
	for _, line := range tier2Note {
		tier2NoteSheet.AddRow().AddCell().SetString(line)
	}

	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load template")

	// exonCount
	exonCount = jsonUtil.JsonFile2Map(*transInfo)

	// 突变频谱
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDbExt := jsonUtil.Json2MapMap(simple_util.File2Decode(*geneDbFile, codeKey))
	for k := range geneDbExt {
		geneDb[k] = geneDbExt[k][geneDbKey]
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load mutation spectrum")

	// 基因-疾病
	geneDiseaseDbTitleInfo := jsonUtil.JsonFile2MapMap(*geneDiseaseDbTitle)
	for key, item := range geneDiseaseDbTitleInfo {
		geneDiseaseDbColumn[key] = item["Key"]
	}
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDiseaseDb = jsonUtil.Json2MapMap(simple_util.File2Decode(*geneDiseaseDbFile, codeKey))
	for key := range geneDiseaseDb {
		geneList[key] = true
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load Gene-Disease DB")

	// 特殊位点库
	for _, key := range textUtil.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load Special mutation DB")

	var stats = make(map[string]int)
	var isHom = regexp.MustCompile(`^Hom`)

	if *exon != "" {
		anno.LoadGeneTrans(anno.GetPath("geneSymbol.transcript", dbPath, defaultConfig))
		var paths []string
		for _, path := range strings.Split(*exon, ",") {
			if simple_util.FileExists(path) {
				paths = append(paths, path)
			} else {
				log.Printf("ERROR:not exists or not a file:%v \n", path)
			}
		}
		addCnv2Sheet(
			tier1Xlsx.Sheet["exon_cnv"], exonCnvTitle, paths, sampleMap,
			false, *cnvFilter, stats, "exonCNV",
		)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add exon cnv")
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
		addCnv2Sheet(
			tier1Xlsx.Sheet["large_cnv"], largeCNVTitle, paths, sampleMap,
			*cnvFilter, false, stats, "largeCNV",
		)
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
		addSmnResult(tier1Xlsx.Sheet["large_cnv"], largeCNVTitle, paths, sampleMap)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add SMN1 result")
	}
	addFamInfoSheet(tier1Xlsx, "fam_info", sampleList)

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
					simple_util.File2Slice(extraArray[i], "\t"),
					xlsxUtil.AddSheet(tier1Xlsx, extraSheetArray[i]),
				)
			}
		}
	}

	// anno
	if *snv != "" {
		var step0 = step
		var data []map[string]string
		for _, f := range snvs {
			if isGz.MatchString(f) {
				d, _ := simple_util.Gz2MapArray(f, "\t", isComment)
				data = append(data, d...)
			} else {
				d, _ := simple_util.File2MapArray(f, "\t", isComment)
				data = append(data, d...)
			}
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load anno file")

		stats["Total"] = len(data)
		for _, item := range data {

			// score to prediction
			anno.Score2Pred(item)

			// update Function
			anno.UpdateFunction(item)

			// update FuncRegion
			anno.UpdateFuncRegion(item)

			gene := item["Gene Symbol"]
			// 基因-疾病
			anno.UpdateDisease(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
			item["Gene"] = item["Omim Gene"]
			item["OMIM"] = item["OMIM_Phenotype_ID"]
			item["death age"] = item["hpo_cn"]

			//anno.AnnoSpliceAI(item)

			// ues acmg of go
			if *acmg {
				if !*autoPVS1 {
					item["PVS1"] = evidence.CheckPVS1(item, LOFList, transcriptInfo, tbx)
				}
				item["PS1"] = evidence.CheckPS1(item)
				item["PM5"] = evidence.CheckPM5(item)
				item["PS4"] = evidence.CheckPS4(item)
				item["PM1"] = evidence.CheckPM1(item, tbx)
				item["PM2"] = evidence.CheckPM2(item)
				item["PM4"] = evidence.CheckPM4(item)
				item["PP2"] = evidence.CheckPP2(item)
				item["PP3"] = evidence.CheckPP3(item, *autoPVS1)
				item["BA1"] = evidence.CheckBA1(item) // BA1 更改条件，去除PVFD，新增ESP6500
				item["BS1"] = evidence.CheckBS1(item) // BS1 更改条件，去除PVFD，也没有对阈值1%进行修正
				item["BS2"] = evidence.CheckBS2(item)
				item["BP1"] = evidence.CheckBP1(item)
				item["BP3"] = evidence.CheckBP3(item)
				item["BP4"] = evidence.CheckBP4(item) // BP4 更改条件，更严格了，非splice未考虑保守性
				item["BP7"] = evidence.CheckBP7(item) // BP 更改条件，更严格了，考虑PhyloP,以及无记录预测按不满足条件来做
			}
			item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)

			anno.UpdateSnv(item, *gender, *debug)

			// 突变频谱
			item["突变频谱"] = geneDb[gene]

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
				if anno.FuncInfo[item["Function"]] >= 3 {
					stats["Tier1LoF"]++
				}
				if isHom.MatchString(item["Zygosity"]) {
					stats["Tier1Hom"]++
				}
				stats["Tier1"+item["VarType"]]++
			}
			stats[item["#Chr"]]++
			if isHom.MatchString(item["Zygosity"]) {
				stats["Hom"]++
				stats["Hom:"+item["#Chr"]]++
			}
			stats[item["VarType"]]++
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
				if *trio || *trio2 {
					item["familyTag"] = anno.FamilyTag(item, inheritDb, "trio")
				} else if *couple {
					item["familyTag"] = anno.FamilyTag(item, inheritDb, "couple")
				} else {
					item["familyTag"] = anno.FamilyTag(item, inheritDb, "single")
				}
				item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio)

				anno.Format(item)

				// Tier1 Sheet
				xlsxUtil.AddMap2Row(item, filterVariantsTitle, tier1Xlsx.Sheet["filter_variants"].AddRow())

				if !*wgs {
					addTier2Row(tier2, item)
				} else {
					tier1Db[item["MutationName"]] = true
				}

				// WESIM
				if *wesim {
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
				}

				tier1GeneList[item["Gene Symbol"]] = true
			}

			// add to tier3
			if !*noTier3 {
				xlsxUtil.AddMap2Row(item, tier3Titles, tier3Sheet.AddRow())
			}
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load snv cycle 2")

		// WGS
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

		ts = append(ts, time.Now())
		step++
		logTime(ts, step0, step, "update info")
	}

	// QC Sheet
	updateQC(stats, qualitys[0])
	addQCSheet(tier1Xlsx, "quality", qualityColumn, qualitys)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "add qc")
	//qcSheet.Cols[1].Width = 12

	// append loh sheet
	if *loh != "" {
		appendLOHs(&xlsxUtil.File{tier1Xlsx}, *loh, *lohSheet, sampleList)
	}

	if *save {
		if *wgs && *snv != "" {
			simpleUtil.CheckErr(wgsXlsx.Save(*prefix + ".WGS.xlsx"))
			ts = append(ts, time.Now())
			step++
			logTime(ts, step-1, step, "save WGS")
		}
	}

	// Tier1 excel
	if *save {
		tagStr := ""
		if *tag != "" {
			tagStr = textUtil.File2Array(*tag)[0]
		}
		var tier1Output string
		if isSMN1 {
			tier1Output = *prefix + ".Tier1" + tagStr + ".SMN1.xlsx"
		} else {
			tier1Output = *prefix + ".Tier1" + tagStr + ".xlsx"
		}
		simpleUtil.CheckErr(tier1Xlsx.Save(tier1Output))
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier1")
	}

	if *save {
		simpleUtil.CheckErr(tier2.save(), "Tier2 save fail")
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier2")
	}

	if *save && *snv != "" && !*noTier3 {
		simpleUtil.CheckErr(tier3Xlsx.Save(*prefix + ".Tier3.xlsx"))
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier3")
	}
	logTime(ts, 0, step, "total work")

	if *memprofile != "" {
		var f, e = os.Create(*memprofile)
		defer simpleUtil.DeferClose(f)
		simpleUtil.CheckErr(e)
		simpleUtil.CheckErr(pprof.WriteHeapProfile(f))
	}
}
