package main

import (
	"flag"
	"path/filepath"
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
		filepath.Join(etcPath, "config.json"),
		"default config file, config will be overwrite by flag",
	)
	filterVariants = flag.String(
		"filter_variants",
		filepath.Join(etcPath, "Tier1.filter_variants.txt"),
		"overwrite template/tier1.xlsx filter_variants sheet columns' title",
	)
	exonCnv = flag.String(
		"exon_cnv",
		filepath.Join(etcPath, "Tier1.exon_cnv.txt"),
		"overwrite template/tier1.xlsx exon_cnv sheet columns' title",
	)
	largeCnv = flag.String(
		"large_cnv",
		filepath.Join(etcPath, "Tier1.large_cnv.txt"),
		"overwrite template/tier1.xlsx large_cnv sheet columns' title",
	)
	tier3Title = flag.String(
		"tier3Title",
		filepath.Join(etcPath, "Tier3.总表.txt"),
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
		"if use autoPVS1 for acmg",
	)
	acmgDb = flag.String(
		"acmgDb",
		filepath.Join(etcPath, "acmg.db.list.txt"),
		"acmg db list",
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
	mt = flag.Bool(
		"mt",
		false,
		"force all MT variant to Tier1",
	)
)
