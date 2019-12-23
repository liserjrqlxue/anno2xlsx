package main

import (
	"flag"
	"fmt"
	"github.com/brentp/bix"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/acmg2015/evidence"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	simple_util "github.com/liserjrqlxue/simple-util"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	acmg = flag.Bool(
		"acmg",
		false,
		"if use new ACMG, fix PVS1, PS1,PS4, PM1,PM2,PM4,PM5 PP2,PP3, BA1, BS1,BS2, BP1,BP3,BP4,BP7",
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
	trio = flag.Bool(
		"trio",
		false,
		"if trio mode",
	)
	gender = flag.String(
		"gender",
		"NA",
		"gender of sample list, comma as sep, if M then change Hom to Hemi in XY not PAR region",
	)
	config = flag.String(
		"config",
		filepath.Join(exPath, "..", "etc", "config.json"),
		"default config file, config will be overwrite by flag",
	)
)

// 突变频谱
var geneDb = make(map[string]string)

// 基因-疾病
var geneList = make(map[string]bool)
var geneDiseaseDb = make(map[string]map[string]string)
var geneDiseaseDbColumn = make(map[string]string)

// 特殊位点库
var specVarDb = make(map[string]bool)

var codeKey []byte

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

var snvs []string

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

var err error

func main() {
	flag.Parse()
	if *snv == "" {
		flag.Usage()
		fmt.Println("\n-snv is required!")
		os.Exit(1)
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
	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})

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

	// 突变频谱
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDbExt := simple_util.Json2MapMap(simple_util.File2Decode(*geneDbFile, codeKey))
	for k := range geneDbExt {
		geneDb[k] = geneDbExt[k][geneDbKey]
	}

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

	// 特殊位点库
	for _, key := range simple_util.File2Array(*specVarList) {
		specVarDb[key] = true
	}

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
	var stats = make(map[string]int)
	tier1Count := 0
	tier1Count2 := 0
	for _, item := range data {

		// score to prediction
		anno.Score2Pred(item)

		// update Function
		anno.UpdateFunction(item)

		gene := item["Gene Symbol"]
		// 基因-疾病
		anno.UpdateDisease(gene, item, geneDiseaseDbColumn, geneDiseaseDb)
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

		anno.UpdateSnv(item, *gender, false)

		// 突变频谱
		item["突变频谱"] = geneDb[gene]

		anno.AddTier(item, stats, geneList, specVarDb, *trio, false, false, AFlist1)
		if item["Tier"] == "Tier1" {
			tier1Count++
		}
		anno.AddTier(item, stats, geneList, specVarDb, *trio, false, false, AFlist2)
		if item["Tier"] == "Tier1" {
			tier1Count2++
		}
	}
	fmt.Printf("Tier1:\t%d\t%d\n", tier1Count, tier1Count2)
}

var AFlist1 = []string{
	"GnomAD EAS AF",
	"GnomAD AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC EAS AF",
	"ExAC AF",
	"PVFD AF",
	"Panel AlleleFreq",
}

var AFlist2 = []string{
	"GnomAD AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC AF",
	"PVFD AF",
	"Panel AlleleFreq",
}
