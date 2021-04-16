package main

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-redis/redis"
	"github.com/tealeg/xlsx/v3"

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

var tier1GeneList = make(map[string]bool)

// WESIM
var (
	resultColumn, qualityColumn []string
	resultFile, qcFile          *os.File
)

var qualitys []map[string]string
var qualityKeyMap = make(map[string]string)

// tier2
var isEnProduct map[string]bool

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
	wgsXlsx   *xlsx.File
	TIPdb     = make(map[string]variant)
	MTdisease = make(map[string]variant)
	MTAFdb    = make(map[string]variant)
	MTTitle   []string
	tier1Db   = make(map[string]bool)
)

var (
	logFile             *os.File
	defaultConfig       map[string]interface{}
	tier2TemplateInfo   templateInfo
	tier2               xlsxTemplate
	err                 error
	ts                  = []time.Time{time.Now()}
	step                = 0
	sampleMap           = make(map[string]bool)
	stats               = make(map[string]int)
	tier1Xlsx           = xlsx.NewFile()
	filterVariantsTitle []string
	tier3Titles         []string
	tier3Xlsx           = xlsx.NewFile()
	tier3Sheet          *xlsx.Sheet
)

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
)
