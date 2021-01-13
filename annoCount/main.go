package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/simple-util"
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
	_, err = fmt.Fprintf(out, "index\tsingleCount\tcumCount\tfile\n")
	simpleUtil.CheckErr(err)
	xlsxList := textUtil.File2Array(*list)
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
		_, err = fmt.Fprintf(out, "%d\t%d\t%d\t%s\n", i, singleCount[i], cumCount[i], fileName)
		simpleUtil.CheckErr(err)
	}
	simpleUtil.CheckErr(simple_util.Json2rawFile(*prefix+".singleCount.json", singleCount))
	simpleUtil.CheckErr(simple_util.Json2rawFile(*prefix+".cumCount.json", cumCount))
	simpleUtil.CheckErr(simple_util.Json2rawFile(*prefix+".uniqDb.json", uniqDb))
}
