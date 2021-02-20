package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "..", "db")
)
var (
	tsv = flag.String(
		"tsv",
		filepath.Join(dbPath, "gnomad.genomes.v3.1.sites.chrM.reduced_annotations.tsv"),
		"input db",
	)
	output = flag.String(
		"output",
		"",
		"output db, default is -tsv.db",
	)
)

func main() {
	flag.Parse()
	if *tsv == "" {
		flag.Usage()
		fmt.Println("-tsv is required!")
		os.Exit(1)
	}
	if *output == "" {
		*output = *tsv + ".db"
	}
	var db, _ = textUtil.File2MapArray(*tsv, "\t", nil)
	var out = osUtil.Create(*output)
	defer simpleUtil.DeferClose(out)
	simpleUtil.HandleError(
		fmt.Fprintln(
			out,
			strings.Join(
				[]string{
					"MTmut",
					"AC_hom",
					"AC_het",
					"AF_hom",
					"AF_het",
					"AN",
				},
				"\t",
			),
		),
	)
	for _, item := range db {
		var pos = simpleUtil.HandleError(strconv.Atoi(item["position"])).(int)
		var ref = []byte(item["ref"])
		var alt = []byte(item["alt"])
		for ref[0] == alt[0] {
			ref = ref[1:]
			alt = alt[1:]
			pos++
			if len(ref) == 0 {
				ref = []byte(".")
			}
			if len(alt) == 0 {
				alt = []byte(".")
			}
		}
		simpleUtil.HandleError(
			fmt.Fprintln(
				out,
				strings.Join(
					[]string{
						fmt.Sprintf("%s:%d%s>%s", item["chromosome"], pos, ref, alt),
						item["AC_hom"],
						item["AC_het"],
						item["AF_hom"],
						item["AF_het"],
						item["AN"],
					},
					"\t",
				),
			),
		)
	}
}
