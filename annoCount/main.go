package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"os"
	"strings"
)

var (
	list = flag.String(
		"list",
		"",
		"anno file list",
	)
)

var keyList = []string{
	"#Chr",
	"Start",
	"Stop",
	"Ref",
	"Call",
	"Transcript",
}

func main() {
	flag.Parse()
	if *list == "" {
		flag.Usage()
		os.Exit(1)
	}

	xlsxList := simple_util.File2Array(*list)
	var uniqDb = make(map[string]int)
	var singeleCount [len(xlsxList)]int
	var cumCount [len(xlsxList)]int

	for i, fileName := range xlsxList {
		_, mapArray := simple_util.Sheet2MapArray(fileName, "filter_variants")
		singeleCount[i] = len(mapArray)
		for _, item := range mapArray {
			var keyArray []string
			for _, key := range keyList {
				keyArray = append(keyArray, item[key])
			}
			uniqDb[strings.Join(keyArray, "\t")]++
		}
		cumCount[i] = len(uniqDb)
		fmt.Printf("%d\t%d\t%d\t%s\n", i, singeleCount[i], cumCount[i], fileName)
	}
	simple_util.Json2rawFile("singelCount.json", singeleCount)
	simple_util.Json2rawFile("cumCount.json", cumCount)
}
