package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/anno2xlsx/anno"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/simple-util"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strings"
	"time"
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
		"input tsv",
	)
	output = flag.String(
		"output",
		"",
		"output, default is -input.tsv",
	)
	cnvType = flag.String(
		"cnvType",
		"",
		"cnvType[exon_cnv|large_cnv]",
	)
	title = flag.String(
		"title",
		filepath.Join(exPath, "title.list"),
		"output title",
	)
	geneDiseaseDbFile = flag.String(
		"geneDisease",
		"",
		"database of 基因-疾病数据库",
	)
	geneDiseaseDbTitle = flag.String(
		"geneDiseaseTitle",
		"",
		"Title map of 基因-疾病数据库",
	)
	config = flag.String(
		"config",
		filepath.Join(exPath, "..", "etc", "config.json"),
		"default config file, config will be overwrite by flag",
	)
	cpuprofile = flag.String(
		"cpuprofile",
		"",
		"cpu profile",
	)
	memprofile = flag.String(
		"memprofile",
		"",
		"mem profile",
	)
)

// 基因-疾病
var geneDiseaseDb = make(map[string]map[string]string)
var geneDiseaseDbColumn = make(map[string]string)

var codeKey []byte

//var err error

// \n -> <br/>
var isLF = regexp.MustCompile(`\n`)

func main() {
	var ts []time.Time
	ts = append(ts, time.Now())

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		simple_util.CheckErr(pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
	}
	if *input == "" {
		flag.Usage()
		fmt.Println("\n-input is required")
		os.Exit(0)
	}
	if *output == "" {
		*output = *input + ".tsv"
	}

	out, err := os.Create(*output)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(out)

	// parser etc/config.json
	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})
	if *geneDiseaseDbFile == "" {
		*geneDiseaseDbFile = anno.GetPath("geneDiseaseDbFile", dbPath, defaultConfig)
	}
	if *geneDiseaseDbTitle == "" {
		*geneDiseaseDbTitle = anno.GetPath("geneDiseaseDbTitle", dbPath, defaultConfig)
	}

	// 基因-疾病
	geneDiseaseDbTitleInfo := simple_util.JsonFile2MapMap(*geneDiseaseDbTitle)
	for key, item := range geneDiseaseDbTitleInfo {
		geneDiseaseDbColumn[key] = item["Key"]
	}
	codeKey = []byte("c3d112d6a47a0a04aad2b9d2d2cad266")
	geneDiseaseDb = simple_util.Json2MapMap(simple_util.File2Decode(*geneDiseaseDbFile, codeKey))

	cnvDb, _ := simple_util.LongFile2MapArray(*input, "\t", nil)
	titles := textUtil.File2Array(*title)

	_, err = fmt.Fprintln(out, strings.Join(titles, "\t"))
	simple_util.CheckErr(err)
	for _, item := range cnvDb {
		gene := item["OMIM_Gene"]
		// 基因-疾病
		anno.UpdateDiseaseMultiGene("<br/>", strings.Split(gene, ";"), item, geneDiseaseDbColumn, geneDiseaseDb)
		item["OMIM"] = item["OMIM_Phenotype_ID"]
		// Primer
		item["Primer"] = anno.CnvPrimer(item, *cnvType)

		var array []string
		for _, key := range titles {
			array = append(array, isLF.ReplaceAllString(item[key], "<br/>"))
		}
		_, err = fmt.Fprintln(out, strings.Join(array, "\t"))
		simple_util.CheckErr(err)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		simple_util.CheckErr(pprof.WriteHeapProfile(f))
		defer simple_util.DeferClose(f)
	}
}
