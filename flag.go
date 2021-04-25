package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// flag
var (
	cfg = flag.String(
		"cfg",
		filepath.Join(etcPath, "config.toml"),
		"toml config document",
	)
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
	config = flag.String(
		"config",
		filepath.Join(etcPath, "config.json"),
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

// bool flag
var (
	academic = flag.Bool(
		"academic",
		false,
		"if non-commercial use",
	)
	save = flag.Bool(
		"save",
		true,
		"if save to excel",
	)
	trio = flag.Bool(
		"trio",
		false,
		"if standard trio mode",
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
		"if use autoPVS1 for acmg",
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
	mt = flag.Bool(
		"mt",
		false,
		"force all MT variant to Tier1",
	)
)

func checkFlag() {
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
}
