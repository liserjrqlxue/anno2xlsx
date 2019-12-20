package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx/v2"
	"os"
	"path/filepath"
	"strings"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "db")
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
	config = flag.String(
		"config",
		filepath.Join(exPath, "..", "etc", "config.json"),
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
		"output only top -top item  (exclude Acmg59Gene)",
	)
)

var Acmg59Gene = make(map[string]bool)
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

	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})

	acmg59GeneList := simple_util.File2Array(anno.GetPath("Acmg59Gene", dbPath, defaultConfig))
	for _, gene := range acmg59GeneList {
		Acmg59Gene[gene] = true
	}
	for _, key := range defaultConfig["resultColumn"].([]interface{}) {
		resultColumn = append(resultColumn, key.(string))
	}
	if *trio {
		resultColumn = append(resultColumn, "Genotype of Family Member 1", "Genotype of Family Member 2")
	}

	xlF, err := xlsx.OpenFile(*input)
	simple_util.CheckErr(err)

	resultFile, err := os.Create(*prefix + ".result.tsv")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(resultFile)
	_, err = fmt.Fprintln(resultFile, strings.Join(resultColumn, "\t"))
	simple_util.CheckErr(err)

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
			if Acmg59Gene[item["Gene Symbol"]] {
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
