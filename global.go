package main

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-redis/redis"
	"github.com/pelletier/go-toml"
	"github.com/tealeg/xlsx/v3"
	"github.com/xuri/excelize/v2"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "db")
	etcPath      = filepath.Join(exPath, "etc")
	templatePath = filepath.Join(exPath, "template")
)

// family list
var sampleList []string

// to-do add exon count info of transcript
var exonCount = make(map[string]string)

// 特殊位点库
var specVarDb = make(map[string]bool)

// 遗传相符
var inheritDb = make(map[string]map[string]int)

var qualityColumn []string

// WESIM
var (
	resultColumn, qcColumn, cnvColumn       []string
	resultFile, qcFile, exonFile, largeFile *os.File
)

var qualitys []map[string]string
var qualityKeyMap = make(map[string]string)

// tier2
var isEnProduct = make(map[string]bool)

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

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
	isMT      = regexp.MustCompile(`MT|chrM`)
	isHom     = regexp.MustCompile(`^Hom`)
)

var redisDb *redis.Client

var isSMN1 bool

var snvs []string

var acmg59Gene = make(map[string]bool)

// WGS
var (
	wgsXlsx *xlsx.File
	MTTitle []string
	tier1Db = make(map[string]bool)
)

var (
	logFile           *os.File
	defaultConfig     map[string]interface{}
	tier2TemplateInfo templateInfo
	tier2             xlsxTemplate
	err               error
	ts                = []time.Time{time.Now()}
	step              = 0
	sampleMap         = make(map[string]bool)
	stats             = make(map[string]int)

	tier1Xlsx           *xlsx.File
	filterVariantsTitle []string
	exonCnvTitle        []string
	largeCnvTitle       []string
	tier3Titles         []string

	tier3Xlsx  *excelize.File
	tier3SW    *excelize.StreamWriter
	tier3RowID = 1
)

// TomlTree Global toml config
var TomlTree *toml.Tree

// database
var (
	aesCode  = "c3d112d6a47a0a04aad2b9d2d2cad266"
	gene2id  map[string]string
	chpo     anno.AnnoDb
	revel    revelDb
	mtGnomAD anno.AnnoDb
	// 突变频谱
	spectrumDb anno.EncodeDb
	// 基因-疾病
	diseaseDb anno.EncodeDb
	geneList  = make(map[string]bool)
	// 产前数据库
	prenatalDb anno.EncodeDb
)

// ACMG
var (
	acmgDb string
)

// find duplicate
var countVar = make(map[string]int)
var duplicateVar = make(map[string][]map[string]string)
var deleteVar = make(map[string]bool)
var transcriptLevel = make(map[string]int)
var tier1Count int

// log
var cycle1Count int
var cycle2Count int

// flag to var
var outputTier3 = false
var homFixRatioThreshold = 0.85

// json
//var tier1Json *os.File
var tier1Data []map[string]string
var qualityJsonColumn []string
