package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	simpleUtil "github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v2"

	"github.com/liserjrqlxue/anno2xlsx/anno"
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
		"if trio mode",
	)
	geneDiseaseDbFile = flag.String(
		"geneDisease",
		"",
		"database of 基因-疾病数据库",
	)
	specVarList = flag.String(
		"specVarList",
		"",
		"特殊位点库",
	)
	config = flag.String(
		"config",
		filepath.Join(exPath, "..", "etc", "config.json"),
		"default config file, config will be overwrite by flag",
	)
)

// regexp
var (
	isGz      = regexp.MustCompile(`\.gz$`)
	isComment = regexp.MustCompile(`^##`)
)

// 特殊位点库
var specVarDb = make(map[string]bool)

// 基因-疾病
var (
	codeKey       []byte
	geneList      = make(map[string]bool)
	geneDiseaseDb = make(map[string]map[string]string)
)

// 遗传相符
var inheritDb = make(map[string]map[string]int)

// Tier1
var tier1GeneList = make(map[string]bool)

func main() {
	flag.Parse()
	if *snv == "" {
		flag.Usage()
		fmt.Println("\n-db is required!")
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = strings.Split(*snv, ",")[0]
	}

	title := simpleUtil.File2Array(*columns)
	out, err := os.Create(*prefix + ".tier1.tsv")
	simpleUtil.CheckErr(err)
	defer simpleUtil.DeferClose(out)
	_, err = fmt.Fprintln(out, strings.Join(title, "\t"))
	simpleUtil.CheckErr(err)

	outExcel := xlsx.NewFile()
	sheet, err := outExcel.AddSheet("filter_variants")
	simpleUtil.CheckErr(err)
	row := sheet.AddRow()
	for _, key := range title {
		row.AddCell().SetString(key)
	}

	// parser etc/config.json
	defaultConfig := simpleUtil.JsonFile2Interface(*config).(map[string]interface{})

	if *specVarList == "" {
		*specVarList = anno.GetPath("specVarList", dbPath, defaultConfig)
	}
	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = anno.GetPath("geneDiseaseDbFile", dbPath, defaultConfig)
	}

	// 特殊位点库
	for _, key := range simpleUtil.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	// 基因-疾病
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDiseaseDb = simpleUtil.Json2MapMap(simpleUtil.File2Decode(*geneDiseaseDbFile, codeKey))
	for key := range geneDiseaseDb {
		geneList[key] = true
	}

	// tier1 filter
	var data []map[string]string
	for _, db := range strings.Split(*snv, ",") {
		if isGz.MatchString(db) {
			d, _ := simpleUtil.Gz2MapArray(db, "\t", isComment)
			data = append(data, d...)
		} else {
			d, _ := simpleUtil.File2MapArray(db, "\t", isComment)
			data = append(data, d...)
		}
	}
	// cycle 1 find common tier1 gene list
	var stats = make(map[string]int)
	for _, item := range data {
		anno.AddTier(item, stats, geneList, specVarDb, *trio, false, false, anno.AFlist)
		if item["Tier"] == "Tier1" {
			tier1GeneList[item["Gene Symbol"]] = true
		}
		anno.AddTier(item, stats, geneList, specVarDb, *trio, true, false, anno.AFlist)
		if item["Tier"] == "Tier1" {
			anno.InheritCheck(item, inheritDb)
		}
	}
	// cycle 2 for 遗传相符
	for _, item := range data {
		if item["Tier"] == "Tier1" && tier1GeneList[item["Gene Symbol"]] {
			// 遗传相符
			item["遗传相符"] = anno.InheritCoincide(item, inheritDb, *trio)
			if item["遗传相符"] == "相符" {
				stats["遗传相符"]++
			}
			item["筛选标签"] = anno.UpdateTags(item, specVarDb, *trio)
			var array []string
			row = sheet.AddRow()
			for _, key := range title {
				row.AddCell().SetString(item[key])
				array = append(array, item[key])
			}
			_, err = fmt.Fprintln(out, strings.Join(array, "\t"))
			simpleUtil.CheckErr(err)
		}
	}
	err = outExcel.Save(*prefix + ".tier1.xlsx")
	simpleUtil.CheckErr(err)
}
