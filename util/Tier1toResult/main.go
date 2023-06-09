package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/tealeg/xlsx/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/simple-util"
	"github.com/pelletier/go-toml"
)

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	etcPath = filepath.Join(exPath, "..", "..", "etc")
)

var (
	input = flag.String(
		"xlsx",
		"",
		"input xlsx",
	)
	prefix = flag.String(
		"prefix",
		"",
		"output to prefix.result.tsv, default is -xlsx",
	)
	sheetName = flag.String(
		"sheetName",
		"filter_variants",
		"sheetName of input",
	)
	cfg = flag.String(
		"cfg",
		filepath.Join(etcPath, "config.toml"),
		"default config file, config will be overwrite by flag",
	)
	trio = flag.Bool(
		"trio",
		false,
		"if trio",
	)
	top = flag.Int(
		"top",
		20,
		"output only top -top item  (exclude acmgSFGene)",
	)
)

// TomlTree Global toml config
var TomlTree *toml.Tree

var acmgSFGene = make(map[string]bool)
var resultColumn []string

func main() {
	flag.Parse()
	if *input == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *input
	}

	TomlTree = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

	for _, gene := range textUtil.File2Array(anno.GuessPath(TomlTree.Get("acmg.SF").(string), etcPath)) {
		acmgSFGene[gene] = true
	}
	resultColumn = TomlTree.GetArray("wesim.resultColumn").([]string)
	if *trio {
		resultColumn = append(resultColumn, "Genotype of Family Member 1", "Genotype of Family Member 2")
	}

	var xlF, err = xlsx.OpenFile(*input)
	simple_util.CheckErr(err)

	var resultFile = osUtil.Create(*prefix + ".result.tsv")
	defer simple_util.DeferClose(resultFile)

	simpleUtil.HandleError(fmt.Fprintln(resultFile, strings.Join(resultColumn, "\t")))

	var title []string
	var count = 0
	for i, row := range xlF.Sheet[*sheetName].Rows {
		if i == 0 {
			for _, cell := range row.Cells {
				title = append(title, cell.Value)
			}
		} else {
			var item = make(map[string]string)
			for j, cell := range row.Cells {
				if j < len(title) && title != nil {
					item[title[j]] = cell.Value
				}
			}
			item["IsACMG59"] = "N"
			if acmgSFGene[item["Gene Symbol"]] {
				item["IsACMG59"] = "Y"
			} else {
				item["IsACMG59"] = "N"
				count++
			}
			if item["IsACMG59"] == "N" && count > *top {
				continue
			}
			if *trio {
				zygosity := strings.Split(item["Zygosity"], ";")
				zygosity = append(zygosity, "NA", "NA")
				item["Zygosity"] = zygosity[0]
				item["Genotype of Family Member 1"] = zygosity[1]
				item["Genotype of Family Member 2"] = zygosity[2]
			}
			var resultArray []string
			for _, key := range resultColumn {
				resultArray = append(resultArray, item[key])
			}
			_, err = fmt.Fprintln(resultFile, strings.Join(resultArray, "\t"))
			simple_util.CheckErr(err)
		}
	}
}
