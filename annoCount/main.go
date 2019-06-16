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
	prefix = flag.String(
		"prefix",
		"",
		"out put prefix, default is -list",
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
	if *prefix == "" {
		*prefix = *list
	}

	out, err := os.Open(*prefix + ".tsv")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(out)
	fmt.Printf("index\tsingleCount\tcumCount\tfile\n")
	fmt.Fprintf(out, "index\tsingleCount\tcumCount\tfile\n")
	xlsxList := simple_util.File2Array(*list)
	var uniqDb = make(map[string]int)
	var singleCount []int
	var cumCount []int

	for i, fileName := range xlsxList {
		_, mapArray := simple_util.Sheet2MapArray(fileName, "filter_variants")
		singleCount = append(singleCount, len(mapArray))
		for _, item := range mapArray {
			var keyArray []string
			for _, key := range keyList {
				keyArray = append(keyArray, item[key])
			}
			uniqDb[strings.Join(keyArray, "\t")]++
		}
		cumCount = append(cumCount, len(uniqDb))
		fmt.Printf("%d\t%d\t%d\t%s\n", i, singleCount[i], cumCount[i], fileName)
		fmt.Fprintf(out, "%d\t%d\t%d\t%s\n", i, singleCount[i], cumCount[i], fileName)
	}
	simple_util.Json2rawFile(*prefix+".singleCount.json", singleCount)
	simple_util.Json2rawFile(*prefix+".cumCount.json", cumCount)
	simple_util.Json2rawFile(*prefix+".uniqDb.json", uniqDb)
}
