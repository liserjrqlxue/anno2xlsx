package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	AES "github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "db")
)

// flag
var (
	input = flag.String(
		"input",
		"",
		"input",
	)
	output = flag.String(
		"output",
		"",
		"output file name",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output prefix",
	)
	suffix = flag.String(
		"suffix",
		"",
		"output suffix",
	)
	sheetName = flag.String(
		"sheet",
		"",
		"sheet name of xlsx",
	)
	key = flag.String(
		"key",
		"",
		"key of each line/row",
	)
	mergeSep = flag.String(
		"mergeSep",
		"\n",
		"sep of merge",
	)
	codeKey = flag.String(
		"codeKey",
		"c3d112d6a47a0a04aad2b9d2d2cad266",
		"codeKey for aes",
	)
	skipWarn = flag.String(
		"skipWarn",
		"",
		"skip warn of columns index (0-based), comma as sep",
	)
	rowCount = flag.Int(
		"rowCount",
		0,
		"check row count",
	)
	keyCount = flag.Int(
		"keyCount",
		0,
		"check key count",
	)
	geneID = flag.String(
		"geneID",
		filepath.Join(dbPath, "gene.id.txt"),
		"check key valid",
	)
)

func main() {
	flag.Parse()
	if *input == "" || *key == "" || *sheetName == "" || *rowCount == 0 || *keyCount == 0 {
		flag.Usage()
		fmt.Println("-input/-key/-sheet/-rowCount/-keyCount are required!")
		os.Exit(1)
	}
	if *output == "" {
		if *prefix == "" {
			*prefix = *input
		}
		if *suffix == "" {
			*suffix = ".json.aes"
		}
	}
	var geneIDdb = textUtil.File2Slice(*geneID, "\t")
	var geneIDkeys = make(map[string]bool)
	for _, k := range geneIDdb {
		geneIDkeys[k[1]] = true
	}

	var skip = make(map[int]bool)
	if *skipWarn != "" {
		for _, index := range strings.Split(*skipWarn, ",") {
			var i, err = strconv.Atoi(index)
			simpleUtil.CheckErr(err, "can not parse "+*skipWarn)
			skip[i] = true
		}
	}
	var codeKeyBytes = []byte(*codeKey)

	var inputFh, err = excelize.OpenFile(*input)
	simpleUtil.CheckErr(err)
	//var inputFh = simpleUtil.HandleError(excelize.OpenFile(*input)).(*excelize.File)
	//fmt.Printf("%+v\n",inputFh.GetSheetMap())
	fmt.Printf("sheet name:\t%s\n", *sheetName)
	fmt.Printf("key column:\t%s\n", *key)
	for _, sheet := range inputFh.GetSheetMap() {
		if *sheetName != sheet {
			fmt.Printf("skip sheet:[%s]\n", sheet)
			continue
		}
		fmt.Printf("encode sheet:[%s]\n", *sheetName)
		sheet2db(inputFh, sheet, geneIDkeys, skip, codeKeyBytes)
	}
}

func sheet2db(inputFh *excelize.File, sheet string, geneIDkeys map[string]bool, skip map[int]bool, code []byte) {
	var valid = true
	var rows = simpleUtil.HandleError(inputFh.GetRows(sheet)).([][]string)
	fmt.Printf("rows:\t%d\t%v\n", len(rows), len(rows) == *rowCount)
	var outputFile = *prefix + *output + "." + sheet + *suffix
	var geneList = *prefix + *output + "." + sheet + "geneList.txt"
	var geneListFH = osUtil.Create(geneList)
	defer simpleUtil.DeferClose(geneListFH)
	var d []byte
	var data, _ = simpleUtil.Slice2MapMapArrayMerge1(rows, *key, *mergeSep, skip)
	var gene2id = make(map[string]string)
	for id := range data {
		if !geneIDkeys[id] {
			fmt.Printf("id:[%s] not contain in %s\n", id, *geneID)
			valid = false
		}
		var geneSymbols = strings.Split(data[id]["Gene/Locus"], *mergeSep)
		sort.Strings(geneSymbols)
		for _, gene := range geneSymbols {
			if gene2id[gene] == "" {
				gene2id[gene] = id
			} else if gene2id[gene] != id {
				fmt.Printf("conflict gene id [%s]vs.[%s] of [%s]\n", id, gene2id[gene], gene)
			}
		}
	}
	var genes = make([]string, 0, len(gene2id))
	for s := range gene2id {
		genes = append(genes, s)
	}
	sort.Strings(genes)
	for _, gene := range genes {
		simpleUtil.HandleError(fmt.Fprintf(geneListFH, "%s\t%s\n", gene, gene2id[gene]))
	}

	d = simpleUtil.HandleError(json.MarshalIndent(data, "", "  ")).([]byte)
	fmt.Printf("keys:\t%d\t%v\n", len(data), len(data) == *keyCount)
	AES.Encode2File(outputFile, d, code)
	fmt.Printf("[%s] checked:\t%v\n", *sheetName, valid)
}
