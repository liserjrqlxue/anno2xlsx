package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	simple_util "github.com/liserjrqlxue/simple-util"
	"github.com/pelletier/go-toml"
	"github.com/tealeg/xlsx/v2"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
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
		"output prefix.tier1.{tsv,xlsx}, default is same to first file of -snv",
	)
	columns = flag.String(
		"columns",
		filepath.Join(exPath, "..", "template", "Tier1.filter_variants.columns.list"),
		"output titles")
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
	specVarList = flag.String(
		"specVarList",
		"",
		"特殊位点库",
	)
	cfg = flag.String(
		"cfg",
		filepath.Join(exPath, "..", "etc", "config.toml"),
		"toml config document",
	)
	config = flag.String(
		"config",
		filepath.Join(exPath, "..", "etc", "config.json"),
		"default config file, config will be overwrite by flag",
	)
	geneID = flag.String(
		"geneId",
		filepath.Join(dbPath, "gene.id.txt"),
		"gene symbol and ncbi id list",
	)
)

var gene2id = make(map[string]string)

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isXlsx    = regexp.MustCompile(`\.xlsx$`)
	isComment = regexp.MustCompile(`^##`)
)

// 特殊位点库
var specVarDb = make(map[string]bool)

// TomlTree Global toml config
var TomlTree *toml.Tree

// 基因-疾病
var (
	aesCode   = "c3d112d6a47a0a04aad2b9d2d2cad266"
	geneList  = make(map[string]bool)
	diseaseDb anno.EncodeDb
)

// 遗传相符
var inheritDb = make(map[string]map[string]int)

// Tier1
var tier1GeneList = make(map[string]bool)

var defaultConfig map[string]interface{}

func init() {
	flag.Parse()
	if *snv == "" {
		flag.Usage()
		fmt.Println("\n-db is required!")
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = strings.Split(*snv, ",")[0]
	}

	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)

	// parser etc/config.json
	defaultConfig = simple_util.JsonFile2Interface(*config).(map[string]interface{})

	if *specVarList == "" {
		*specVarList = anno.GetPath("specVarList", dbPath, defaultConfig)
	}

	// 特殊位点库
	for _, key := range textUtil.File2Array(*specVarList) {
		specVarDb[key] = true
	}

	TomlTree = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)
	// 基因-疾病
	diseaseDb.Load(
		TomlTree.Get("annotation.Gene.disease").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	for key := range diseaseDb.Db {
		geneList[key] = true
	}
	for k, v := range gene2id {
		if geneList[v] {
			geneList[k] = true
		}
	}
}
func main() {
	title := textUtil.File2Array(*columns)
	out, err := os.Create(*prefix + ".tier1.tsv")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(out)
	_, err = fmt.Fprintln(out, strings.Join(title, "\t"))
	simple_util.CheckErr(err)

	outExcel := xlsx.NewFile()
	sheet, err := outExcel.AddSheet("filter_variants")
	simple_util.CheckErr(err)
	row := sheet.AddRow()
	for _, key := range title {
		row.AddCell().SetString(key)
	}

	// tier1 filter
	var data []map[string]string
	for _, db := range strings.Split(*snv, ",") {
		if isGz.MatchString(db) {
			d, _ := simple_util.Gz2MapArray(db, "\t", isComment)
			data = append(data, d...)
		} else if isXlsx.MatchString(db) {
			_, d := simple_util.Sheet2MapArray(db, "sheet1")
			data = append(data, d...)
		} else {
			d, _ := simple_util.File2MapArray(db, "\t", isComment)
			data = append(data, d...)
		}
	}
	// cycle 1 find common tier1 gene list
	var stats = make(map[string]int)
	for _, item := range data {
		anno.AddTier(item, stats, geneList, specVarDb, *trio, true, false, anno.AFlist)
		if item["Tier"] == "Tier1" {
			anno.InheritCheck(item, inheritDb)
		}
	}
	// cycle 2 for 遗传相符
	for _, item := range data {
		if item["Tier"] == "Tier1" {
			// 遗传相符
			item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
			item["遗传相符-经典trio"] = anno.InheritCoincide(item, inheritDb, true)
			item["遗传相符-非经典trio"] = anno.InheritCoincide(item, inheritDb, false)
			if item["遗传相符"] == "相符" {
				stats["遗传相符"]++
			}
			item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio, *trio2)
			var array []string
			row = sheet.AddRow()
			for _, key := range title {
				row.AddCell().SetString(item[key])
				array = append(array, item[key])
			}
			_, err = fmt.Fprintln(out, strings.Join(array, "\t"))
			simple_util.CheckErr(err)
		}
	}
	err = outExcel.Save(*prefix + ".tier1.xlsx")
	simple_util.CheckErr(err)
}
