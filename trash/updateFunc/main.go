package main

import (
	"bufio"
	"flag"
	"regexp"
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/scannerUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/stringsUtil"
)

var (
	input = flag.String(
		"input",
		"",
		"input",
	)
	output = flag.String(
		"output",
		"",
		"output, default is -input.updateFunc.tsv",
	)
)

var (
	sep   = "\t"
	chgvs = regexp.MustCompile(`c\.\d+([+-])(\d+)`)
)

func main() {
	flag.Parse()
	if *input == "" {
		flag.Usage()
		panic("-input is required!")
	}
	if *output == "" {
		*output = *input + ".updateFunc.tsv"
	}

	var file = osUtil.Open(*input)
	defer simpleUtil.DeferClose(file)
	var scannaer = bufio.NewScanner(file)

	var outF = osUtil.Create(*output)
	defer simpleUtil.DeferClose(outF)

	var title = scannerUtil.ScanTitle(scannaer, sep, nil)
	var index = make(map[string]int)
	for i, k := range title {
		index[k] = i
	}
	var functionIndex, ok1 = index["Function"]
	if !ok1 {
		panic("no \"Function\" in title of input:" + *input)
	}
	var chgvsIndex, ok2 = index["cHGVS"]
	if !ok2 {
		panic("no \"cHGVS\" in title of input:" + *input)
	}
	fmtUtil.FprintStringArray(outF, title, sep)

	for scannaer.Scan() {
		var line = scannaer.Text()
		var array = strings.Split(line, sep)
		var function = array[functionIndex]
		var cHGVS = array[chgvsIndex]
		var newFunction = updateFunction(function, cHGVS)
		array[functionIndex] = newFunction
		fmtUtil.FprintStringArray(outF, array, sep)
	}
	simpleUtil.CheckErr(scannaer.Err())
}

func updateFunction(function, cHGVS string) string {
	if function == "intron" {
		var matches = chgvs.FindStringSubmatch(cHGVS)
		if matches != nil {
			var strand = matches[1]
			var distance = stringsUtil.Atoi(matches[2])
			if distance <= 10 {
				return "splice" + strand + "10"
			} else if distance <= 20 {
				return "splice" + strand + "20"
			}
		}
	}
	return function
}
