package main

import (
	"flag"
	"log"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

var (
	in = flag.String(
		"i",
		"",
		"input",
	)
	add = flag.String(
		"a",
		"",
		"addition",
	)
	out = flag.String(
		"o",
		"",
		"output",
	)
)

func main() {
	flag.Parse()
	if *in == "" || *add == "" || *out == "" {
		flag.Usage()
		log.Fatalln("-i/-a/-o required!")
	}
	var input, title = textUtil.File2MapArray(*in, "\t", nil)
	var adds, _ = textUtil.File2MapArray(*add, "\t", nil)
	if len(input) != len(adds) {
		log.Fatalf("Conflict: [%s:%d]vs.[%s:%d]\n", *in, len(input), *add, len(adds))
	}
	var output = osUtil.Create(*out)
	defer simpleUtil.DeferClose(output)
	fmtUtil.FprintStringArray(output, title, "\t")
	for i, m := range input {
		m["Filter"] = adds[i]["pre_class"]
		var a []string
		for _, s := range title {
			a = append(a, m[s])
		}
		fmtUtil.FprintStringArray(output, a, "\t")
	}
}
