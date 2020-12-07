package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	AES "github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
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
	var valid = true
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
	for _, sheet := range inputFh.GetSheetMap() {
		if *sheetName != sheet {
			fmt.Printf("skip sheet:[%s]\n", sheet)
			continue
		}
		fmt.Printf("encode sheet:[%s]\n", *sheetName)
		var rows = simpleUtil.HandleError(inputFh.GetRows(sheet)).([][]string)
		fmt.Printf("rows:\t%d\t%v\n", len(rows), len(rows) == *rowCount)
		valid = valid && len(rows) == *rowCount
		var outputFile = *prefix + *output + "." + sheet + *suffix
		var d []byte
		var data, _ = simpleUtil.Slice2MapMapArrayMerge1(rows, *key, *mergeSep, skip)
		for key := range data {
			if !geneIDkeys[key] {
				fmt.Printf("key:[%s] not contain in %s\n", key, *geneID)
				valid = false
			}
		}
		d = simpleUtil.HandleError(json.MarshalIndent(data, "", "  ")).([]byte)
		fmt.Printf("keys:\t%d\t%v\n", len(data), len(data) == *keyCount)
		valid = valid && len(data) == *keyCount
		AES.Encode2File(outputFile, d, codeKeyBytes)
		fmt.Printf("[%s] checked:\t%v\n", *sheetName, valid)
	}
}
