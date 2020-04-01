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
	"github.com/tealeg/xlsx/v2"
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
	dbPath       = filepath.Join(exPath, "..", "db")
	templatePath = filepath.Join(exPath, "..", "template")
)

// flag
var (
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
		filepath.Join(exPath, "..", "etc", "config.json"),
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
	debug = flag.Bool(
		"debug",
		false,
		"if print some log",
	)
	tag = flag.String(
		"tag",
		"",
		"read tag from file, add to tier1 file name:[prefix].Tier1[tag].xlsx",
	)
	filterVariants = flag.String(
		"filter_variants",
		filepath.Join(exPath, "..", "etc", "Tier1.filter_variants.txt"),
		"overwrite template/tier1.xlsx filter_variants sheet columns' title",
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
		template:  filepath.Join(templatePath, flag+".xlsx"),
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

var codeKey []byte

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

var redisDb *redis.Client

var isSMN1 bool

var snvs []string

var Acmg59Gene = make(map[string]bool)

// ACMG
// PVS1
var LOFList map[string]int
var transcriptInfo map[string][]evidence.Region

// PS1 & PM5
var (
	HGMDAAPosList    map[string]int
	ClinVarAAPosList map[string]int
	HGMDPHGVSlist    map[string]int
	HGMDMissense     map[string]int
	ClinVarPHGVSlist map[string]int
	ClinVarMissense  map[string]int
)

// PM1
var tbx *bix.Bix
var (
	PfamDomain   map[string]bool
	dbNSFPDomain map[string]bool
)

// PP2
var (
	HgmdPP2GeneList    map[string]float64
	ClinVarPP2GeneList map[string]float64
)

// BS2
var lateOnsetList map[string]int

// BP1
var (
	HgmdBP1GeneList    map[string]float64
	ClinVarBP1GeneList map[string]float64
)

func main() {
	var ts []time.Time
	var step = 0
	ts = append(ts, time.Now())

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		simple_util.CheckErr(pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
	}
	if *snv == "" {
		flag.Usage()
		fmt.Println("\nshold have input -snv")
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

	// parser etc/config.json
	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})

	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = anno.GetStrVal("redisServer", defaultConfig)
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
		simple_util.JsonFile2Data(anno.GetPath("LOFList", dbPath, defaultConfig), &LOFList)
		simple_util.JsonFile2Data(anno.GetPath("transcriptInfo", dbPath, defaultConfig), &transcriptInfo)

		// PS1 & PM5
		simple_util.JsonFile2Data(anno.GetPath("ClinVarPathogenicMissense", dbPath, defaultConfig), &ClinVarMissense)
		simple_util.JsonFile2Data(anno.GetPath("ClinVarPHGVSlist", dbPath, defaultConfig), &ClinVarPHGVSlist)
		simple_util.JsonFile2Data(anno.GetPath("HGMDPathogenicMissense", dbPath, defaultConfig), &HGMDMissense)
		simple_util.JsonFile2Data(anno.GetPath("HGMDPHGVSlist", dbPath, defaultConfig), &HGMDPHGVSlist)
		simple_util.JsonFile2Data(anno.GetPath("ClinVarAAPosList", dbPath, defaultConfig), &ClinVarAAPosList)
		simple_util.JsonFile2Data(anno.GetPath("HGMDAAPosList", dbPath, defaultConfig), &HGMDAAPosList)

		// PM1
		simple_util.JsonFile2Data(anno.GetPath("PM1dbNSFPDomain", dbPath, defaultConfig), &dbNSFPDomain)
		simple_util.JsonFile2Data(anno.GetPath("PM1PfamDomain", dbPath, defaultConfig), &PfamDomain)
		tbx, err = bix.New(anno.GetPath("PathogenicLite", dbPath, defaultConfig))
		simple_util.CheckErr(err, "load tabix")

		// PP2
		simple_util.JsonFile2Data(anno.GetPath("ClinVarPP2GeneList", dbPath, defaultConfig), &ClinVarPP2GeneList)
		simple_util.JsonFile2Data(anno.GetPath("HgmdPP2GeneList", dbPath, defaultConfig), &HgmdPP2GeneList)

		// BS2
		simple_util.JsonFile2Data(anno.GetPath("LateOnset", dbPath, defaultConfig), &lateOnsetList)

		// BP1
		simple_util.JsonFile2Data(anno.GetPath("ClinVarBP1GeneList", dbPath, defaultConfig), &ClinVarBP1GeneList)
		simple_util.JsonFile2Data(anno.GetPath("HgmdBP1GeneList", dbPath, defaultConfig), &HgmdBP1GeneList)
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

	for _, key := range defaultConfig["qualityColumn"].([]interface{}) {
		qualityColumn = append(qualityColumn, key.(string))
	}

	if *wesim {
		acmg59GeneList := simple_util.File2Array(anno.GetPath("Acmg59Gene", dbPath, defaultConfig))
		for _, gene := range acmg59GeneList {
			Acmg59Gene[gene] = true
		}

		for _, key := range defaultConfig["resultColumn"].([]interface{}) {
			resultColumn = append(resultColumn, key.(string))
		}
		if *trio {
			resultColumn = append(resultColumn, "Genotype of Family Member 1", "Genotype of Family Member 2")
		}
		resultFile, err = os.Create(*prefix + ".result.tsv")
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(resultFile)
		_, err = fmt.Fprintln(resultFile, strings.Join(resultColumn, "\t"))
		simple_util.CheckErr(err)

		qcFile, err = os.Create(*prefix + ".qc.tsv")
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(qcFile)
		_, err = fmt.Fprintln(qcFile, strings.Join(qualityColumn, "\t"))
		simple_util.CheckErr(err)
	}

	sampleList = strings.Split(*list, ",")
	var sampleMap = make(map[string]bool)
	for _, sample := range sampleList {
		sampleMap[sample] = true
	}

	// load tier template
	tier1 := newXlsxTemplate("Tier1")
	// update tier1 titles
	titleRow := tier1.sheet.Row(0)
	tier1.title = simple_util.File2Array(*filterVariants)
	titleCells := titleRow.Cells
	for i, v := range tier1.title {
		if i < len(titleCells) {
			titleRow.Cells[i].SetString(v)
		} else {
			titleRow.AddCell().SetString(v)
		}
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
		geneDb[k] = geneDbExt[k][geneDbKey]
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

	var stats = make(map[string]int)

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

		stats["Total"] = len(data)
		for _, item := range data {

			// score to prediction
			anno.Score2Pred(item)

			// update Function
			anno.UpdateFunction(item)

			gene := item["Gene Symbol"]
			// 基因-疾病
			updateDisease(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
			item["Gene"] = item["Omim Gene"]
			item["OMIM"] = item["OMIM_Phenotype_ID"]
			item["death age"] = item["hpo_cn"]

			// ues acmg of go
			if *acmg {
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

			anno.UpdateSnv(item, *gender, *debug)

			// 突变频谱
			item["突变频谱"] = geneDb[gene]

			// 引物设计
			item["exonCount"] = exonCount[item["Transcript"]]
			item["引物设计"] = anno.PrimerDesign(item)

			// 变异来源
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
			item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio)
			anno.FloatFormat(item)

			tier1Row := tier1.sheet.AddRow()
			for _, str := range tier1.title {
				tier1Row.AddCell().SetString(item[str])
			}
			// WESIM
			if *wesim {
				if Acmg59Gene[item["Gene Symbol"]] {
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
				simple_util.CheckErr(err)
			}

			tier1GeneList[item["Gene Symbol"]] = true
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
		tagStr := ""
		if *tag != "" {
			tagStr = simple_util.File2Array(*tag)[0]
		}
		if isSMN1 {
			tier1.output = *prefix + ".Tier1" + tagStr + ".SMN1.xlsx"
		} else {
			tier1.output = *prefix + ".Tier1" + tagStr + ".xlsx"
		}
		err = tier1.save()
		simple_util.CheckErr(err)
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "save Tier1")
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		simple_util.CheckErr(pprof.WriteHeapProfile(f))
		defer simple_util.DeferClose(f)
	}
}

func logTime(timeList []time.Time, step1, step2 int, message string) {
	trim := 3*8 - 1
	str := simple_util.FormatWidth(trim, message, ' ')
	fmt.Printf("%s\ttook %7.3fs to run.\n", str, timeList[step2].Sub(timeList[step1]).Seconds())
}

func updateDisease(gene string, item, geneDisDbColumn map[string]string, geneDisDb map[string]map[string]string) {
	// 基因-疾病
	gDiseaseDb := geneDisDb[gene]
	for key, value := range geneDisDbColumn {
		item[value] = gDiseaseDb[key]
	}
}
