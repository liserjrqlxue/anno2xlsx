package main

import (
	"flag"
	"fmt"
	"github.com/brentp/bix"
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
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "db")
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
	config = flag.String(
		"config",
		filepath.Join(exPath, "..", "etc", "config.json"),
		"default config file, config will be overwrite by flag",
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
)

// 基因-疾病
var geneDiseaseDb = make(map[string]map[string]string)
var geneDiseaseDbColumn = make(map[string]string)

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

var codeKey []byte

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

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

var err error

func main() {
	var ts []time.Time
	ts = append(ts, time.Now())

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *snv == "" {
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
			*prefix = *snv
		}
	}

	out, err := os.Create(*prefix + ".tsv")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(out)

	// parser etc/config.json
	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})

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

	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = anno.GetPath("geneDiseaseDbFile", dbPath, defaultConfig)
	}
	if *geneDiseaseDbTitle == "" {
		*geneDiseaseDbTitle = anno.GetPath("geneDiseaseDbTitle", dbPath, defaultConfig)
	}

	// 基因-疾病
	geneDiseaseDbTitleInfo := simple_util.JsonFile2MapMap(*geneDiseaseDbTitle)
	for key, item := range geneDiseaseDbTitleInfo {
		geneDiseaseDbColumn[key] = item["Key"]
	}
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDiseaseDb = simple_util.Json2MapMap(simple_util.File2Decode(*geneDiseaseDbFile, codeKey))

	// anno
	var data []map[string]string
	var title []string
	if isGz.MatchString(*snv) {
		data, title = simple_util.Gz2MapArray(*snv, "\t", isComment)
	} else {
		data, title = simple_util.File2MapArray(*snv, "\t", isComment)
	}

	addTitle := []string{
		"PVS1",
		"PS1",
		"PM5",
		"PS4",
		"PM1",
		"PM2",
		"PM4",
		"PP2",
		"PP3",
		"BA1",
		"BS1",
		"BS2",
		"BP1",
		"BP3",
		"BP7",
		"自动化判断",
	}
	title = append(title, addTitle...)

	_, err = fmt.Fprintln(out, strings.Join(title, "\t"))
	simple_util.CheckErr(err)
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
		item["自动化判断"] = acmg2015.PredACMG2015(item)

		var array []string
		for _, key := range title {
			array = append(array, item[key])
		}
		_, err = fmt.Fprintln(out, strings.Join(array, "\t"))
		simple_util.CheckErr(err)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer simple_util.DeferClose(f)
	}
}

func updateDisease(gene string, item, geneDiseaseDbColumn map[string]string, geneDiseaseDb map[string]map[string]string) {
	// 基因-疾病
	gDiseaseDb := geneDiseaseDb[gene]
	for key, value := range geneDiseaseDbColumn {
		item[value] = gDiseaseDb[key]
	}
}
