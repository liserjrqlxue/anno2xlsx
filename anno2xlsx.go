package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	simple_util "github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v3"

	"github.com/pelletier/go-toml"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

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

func init() {
	logVersion()

	flag.Parse()
	checkFlag()

	// log
	logFile, err = os.Create(*logfile)
	simpleUtil.CheckErr(err)
	log.Printf("Log file         : %v\n", *logfile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Log file         : %v\n", *logfile)
	logVersion()

	parseCfg()

}

func parseCfg() {
	// parser etc/config.json
	defaultConfig = jsonUtil.JsonFile2Interface(*config).(map[string]interface{})

	openRedis()
	initAcmg2015()

	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = anno.GetPath("geneDiseaseDbFile", dbPath, defaultConfig)
	}
	if *geneDiseaseDbTitle == "" {
		*geneDiseaseDbTitle = anno.GetPath("geneDiseaseDbTitle", dbPath, defaultConfig)
	}
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

	initIM()

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

	parseList()
	parseQC()
}

func initIM() {
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
		_, err = fmt.Fprintln(resultFile, strings.Join(resultColumn, "\t"))
		simpleUtil.CheckErr(err)

		qcFile, err = os.Create(*prefix + ".qc.tsv")
		simpleUtil.CheckErr(err)
		_, err = fmt.Fprintln(qcFile, strings.Join(qualityColumn, "\t"))
		simpleUtil.CheckErr(err)
	}
}

func openRedis() {
	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = anno.GetStrVal("redisServer", defaultConfig)
		}
		redisDb = redis.NewClient(&redis.Options{
			Addr: *redisAddr,
		})
		log.Printf("Connect [%s]:%s\n", redisDb.String(), simpleUtil.HandleError(redisDb.Ping().Result()).(string))
	}

}

func initAcmg2015() {
	if *acmg {
		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(*acmgDb, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = anno.GuessPath(v, dbPath)
		}
		acmg2015.Init(acmgCfg)
	}
}

func parseList() {
	sampleList = strings.Split(*list, ",")
	for _, sample := range sampleList {
		sampleMap[sample] = true
		quality := make(map[string]string)
		quality["样本编号"] = sample
		qualitys = append(qualitys, quality)
	}
}

func parseQC() {
	var karyotypeMap = make(map[string]string)
	if *karyotype != "" {
		karyotypeMap, err = textUtil.Files2Map(*karyotype, "\t", true)
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
		if *wesim {
			simpleUtil.CheckErr(qcFile.Close())
		}

		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "load coverage.report")
		loadFilterStat(*filterStat, qualitys[0])
	}
}

func prepareTier1() {
	// load tier template
	xlsxUtil.AddSheets(tier1Xlsx, []string{"filter_variants", "exon_cnv", "large_cnv"})
	filterVariantsTitle = addFile2Row(*filterVariants, tier1Xlsx.Sheet["filter_variants"].AddRow())
}

func prepareTier2() {
	// tier2
	tier2 = xlsxTemplate{
		flag:      "Tier2",
		sheetName: *productID + "_" + sampleList[0],
	}
	tier2.output = *prefix + "." + tier2.flag + ".xlsx"
	tier2.xlsx = xlsx.NewFile()

	var tier2Infos = simpleUtil.HandleError(
		simpleUtil.HandleError(xlsx.OpenFile(filepath.Join(templatePath, "Tier2.xlsx"))).(*xlsx.File).ToSlice(),
	).([][][]string)
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
	var tier2NoteSheet = simpleUtil.HandleError(tier2.xlsx.AddSheet(tier2NoteSheetName)).(*xlsx.Sheet)
	for _, line := range tier2Note {
		tier2NoteSheet.AddRow().AddCell().SetString(line)
	}
}

func prepareTier3() {
	// create Tier3.xlsx
	tier3Sheet = xlsxUtil.AddSheet(tier3Xlsx, "总表")
	if !*noTier3 {
		tier3Titles = addFile2Row(*tier3Title, tier3Sheet.AddRow())
	}
}

func prepareExcel() {
	prepareTier1()
	prepareTier2()
	prepareTier3()
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load template")
}

func addExon() {
	var exonCnvTitle = addFile2Row(*exonCnv, tier1Xlsx.Sheet["exon_cnv"].AddRow())
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
		ts = append(ts, time.Now())
		step++
		logTime(ts, step-1, step, "add exon cnv")
	}
}

func addLarge() {
	var largeCNVTitle = addFile2Row(*largeCnv, tier1Xlsx.Sheet["large_cnv"].AddRow())
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
			if osUtil.FileExists(path) {
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
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load anno file")
	return
}

func main() {
	if *cpuprofile != "" {
		var f = osUtil.Create(*cpuprofile)
		simpleUtil.CheckErr(pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
	}
	defer simpleUtil.DeferClose(logFile)

	var tomlConfig = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)
	var hpoCfg = tomlConfig.Get("annotation.hpo").(*toml.Tree)
	var revelCfg = tomlConfig.Get("annotation.REVEL").(*toml.Tree)
	var mtCfg = tomlConfig.Get("annotation.GnomAD.MT").(*toml.Tree)
	var spectrumCfg = tomlConfig.Get("annotation.Gene.spectrum").(*toml.Tree)
	var diseaseCfg = tomlConfig.Get("annotation.Gene.disease").(*toml.Tree)

	chpo.Load(hpoCfg, dbPath)
	if *academic {
		revel.loadRevel(revelCfg)
	}
	mtGnomAD.Load(mtCfg, dbPath)

	// 突变频谱
	spectrumDb.Load(spectrumCfg, dbPath, []byte(aesCode))
	// 基因-疾病
	diseaseDb.Load(diseaseCfg, dbPath, []byte(aesCode))
	for key := range diseaseDb.Db {
		geneList[key] = true
	}
	for k, v := range gene2id {
		if geneList[v] {
			geneList[k] = true
		}
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load Gene-Disease DB")

	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)

	prepareExcel()

	// exonCount
	exonCount = jsonUtil.JsonFile2Map(*transInfo)

	// 特殊位点库
	for _, key := range textUtil.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "load Special mutation DB")

	addExon()
	addLarge()
	addExtra()
	addFamInfoSheet(tier1Xlsx, "fam_info", sampleList)
	addFV()

	// QC Sheet
	updateQC(stats, qualitys[0])
	addQCSheet(tier1Xlsx, "quality", qualityColumn, qualitys)
	ts = append(ts, time.Now())
	step++
	logTime(ts, step-1, step, "add qc")
	//qcSheet.Cols[1].Width = 12

	// append loh sheet
	if *loh != "" {
		appendLOHs(&xlsxUtil.File{File: tier1Xlsx}, *loh, *lohSheet, sampleList)
	}

	saveExcel()

	if *memprofile != "" {
		var f = osUtil.Create(*memprofile)
		defer simpleUtil.DeferClose(f)
		simpleUtil.CheckErr(pprof.WriteHeapProfile(f))
	}
}

func saveExcel() {
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
		if isSMN1 && !*wesim {
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
}
