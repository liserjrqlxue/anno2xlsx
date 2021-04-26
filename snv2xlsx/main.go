package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	"github.com/liserjrqlxue/simple-util"
	"github.com/liserjrqlxue/version"
	"github.com/pelletier/go-toml"
	"github.com/tealeg/xlsx/v3"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	etcPath = filepath.Join(exPath, "..", "etc")
	dbPath  = filepath.Join(exPath, "..", "db")
)

var (
	cfg = flag.String(
		"cfg",
		filepath.Join(etcPath, "config.toml"),
		"toml config document",
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
	geneID = flag.String(
		"geneId",
		filepath.Join(dbPath, "gene.id.txt"),
		"gene symbol and ncbi id list",
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
	list = flag.String(
		"list",
		"proband,father,mother",
		"sample list for family mode, comma as sep",
	)
	gender = flag.String(
		"gender",
		"NA",
		"gender of sample list, comma as sep, if M then change Hom to Hemi in XY not PAR region",
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
	config = flag.String(
		"config",
		filepath.Join(etcPath, "config.json"),
		"default config file, config will be overwrite by flag",
	)
	acmg = flag.Bool(
		"acmg",
		false,
		"if use new ACMG, fix PVS1, PS1,PS4, PM1,PM2,PM4,PM5 PP2,PP3, BA1, BS1,BS2, BP1,BP3,BP4,BP7",
	)
	autoPVS1 = flag.Bool(
		"autoPVS1",
		false,
		"if use autoPVS1 for acmg",
	)
	tag = flag.String(
		"tag",
		"",
		"read tag from file, add to tier1 file name:[prefix].Tier1[tag].xlsx",
	)
	filterVariants = flag.String(
		"filter_variants",
		filepath.Join(etcPath, "Tier1.filter_variants.txt"),
		"overwrite template/tier1.xlsx filter_variants sheet columns' title",
	)
)

var tomlCfg *toml.Tree

// database
var (
	aesCode = "c3d112d6a47a0a04aad2b9d2d2cad266"
	gene2id map[string]string
	chpo    anno.AnnoDb
	// 突变频谱
	spectrumDb anno.EncodeDb
	// 基因-疾病
	diseaseDb anno.EncodeDb
	geneList  = make(map[string]bool)
)

// family list
var sampleList []string

// to-do add exon count info of transcript
var exonCount = make(map[string]string)

// 特殊位点库
var specVarDb = make(map[string]bool)

var tier1GeneList = make(map[string]bool)

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

var redisDb *redis.Client

var snvs []string

var defaultConfig map[string]interface{}

func init() {
	version.LogVersion()

	flag.Parse()
	if *snv == "" {
		flag.Usage()
		fmt.Println("\nshold have input -snv")
		os.Exit(0)
	}
	snvs = strings.Split(*snv, ",")
	if *prefix == "" {
		*prefix = snvs[0]
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

	tomlCfg = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

	chpo.Load(
		tomlCfg.Get("annotation.hpo").(*toml.Tree),
		dbPath,
	)

	// 突变频谱
	spectrumDb.Load(
		tomlCfg.Get("annotation.Gene.spectrum").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	// 基因-疾病
	diseaseDb.Load(
		tomlCfg.Get("annotation.Gene.disease").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	for key := range diseaseDb.Db {
		geneList[key] = true
	}
	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)
	for k, v := range gene2id {
		if geneList[v] {
			geneList[k] = true
		}
	}

	// parser etc/config.json
	defaultConfig = jsonUtil.JsonFile2Interface(*config).(map[string]interface{})

	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = anno.GetStrVal("redisServer", defaultConfig)
		}
		redisDb = redis.NewClient(&redis.Options{
			Addr: *redisAddr,
		})
		pong, e := redisDb.Ping().Result()
		log.Println("connect redis:", pong, e)
		if e != nil {
			log.Fatalf("Error connect redis[%+v]\n", e)
		}
	}

	if *acmg {
		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(tomlCfg.Get("acmg.list").(string), "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = anno.GuessPath(v, dbPath)
		}
		acmg2015.Init(acmgCfg)
	}

	if *specVarList == "" {
		*specVarList = anno.GetPath("specVarList", dbPath, defaultConfig)
	}
	if *transInfo == "" {
		*transInfo = anno.GetPath("transInfo", dbPath, defaultConfig)
	}

	sampleList = strings.Split(*list, ",")
	var sampleMap = make(map[string]bool)
	for _, sample := range sampleList {
		sampleMap[sample] = true
	}
}

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	var tier1Xlsx = xlsx.NewFile()
	var filterVariantsSheet = xlsxUtil.AddSheet(tier1Xlsx, "filter_variants")
	var filterVariantsTitle = addFile2Row(*filterVariants, filterVariantsSheet.AddRow())

	// exonCount
	exonCount = jsonUtil.JsonFile2Map(*transInfo)

	// 特殊位点库
	for _, key := range textUtil.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load Special mutation DB")

	var stats = make(map[string]int)

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
			updateSNV(item, stats)
			xlsxUtil.AddMap2Row(item, filterVariantsTitle, filterVariantsSheet.AddRow())
		}
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load snv cycle 2")

		ts = append(ts, time.Now())
		step++
		logTime(ts, step0, step, "update info")
	}

	// Tier1 excel
	if *save {
		var tagStr = ""
		if *tag != "" {
			tagStr = textUtil.File2Array(*tag)[0]
		}
		simpleUtil.CheckErr(tier1Xlsx.Save(*prefix + ".Tier1" + tagStr + ".xlsx"))
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier1")
	}
}

func updateSNV(item map[string]string, stats map[string]int) {

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

	chpo.Anno(item, id)
	// 基因-疾病
	diseaseDb.Anno(item, id)
	// 突变频谱
	spectrumDb.Anno(item, id)

	item["Gene"] = item["Omim Gene"]
	item["OMIM"] = item["OMIM_Phenotype_ID"]

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

	anno.AddTier(item, stats, geneList, specVarDb, *trio, false, false, anno.AFlist)

	anno.UpdateSnvTier1(item)
	if *ifRedis {
		anno.UpdateRedis(item, redisDb, *seqType)
	}

	anno.UpdateAutoRule(item)
	anno.UpdateManualRule(item)
	item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio, *trio2)
	anno.Format(item)

	tier1GeneList[item["Gene Symbol"]] = true
}

func logTime(timeList []time.Time, step1, step2 int, message string) {
	trim := 3*8 - 1
	str := simple_util.FormatWidth(trim, message, ' ')
	fmt.Printf("%s\ttook %7.3fs to run.\n", str, timeList[step2].Sub(timeList[step1]).Seconds())
}

func addFile2Row(file string, row *xlsx.Row) (rows []string) {
	rows = textUtil.File2Array(file)
	xlsxUtil.AddArray2Row(rows, row)
	return
}
