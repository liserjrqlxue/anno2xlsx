package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/simple-util"
	"github.com/liserjrqlxue/version"
	"github.com/pelletier/go-toml"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	etcPath = filepath.Join(exPath, "..", "..", "etc")
	dbPath  = filepath.Join(exPath, "..", "..", "db")
)

// flag
var (
	cfg = flag.String(
		"cfg",
		filepath.Join(etcPath, "config.toml"),
		"toml config document",
	)
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
	title = flag.String(
		"title",
		filepath.Join(exPath, "addition.txt"),
		"output addition title",
	)
	geneID = flag.String(
		"geneId",
		filepath.Join(dbPath, "gene.id.txt"),
		"gene symbol and ncbi id list",
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
	warn = flag.Bool(
		"warn",
		false,
		"warn gene id lost rather than fatal",
	)
)
var tomlCfg *toml.Tree

// database
var (
	aesCode = "c3d112d6a47a0a04aad2b9d2d2cad266"
	gene2id map[string]string
	// 突变频谱
	spectrumDb anno.EncodeDb
	// 基因-疾病
	diseaseDb anno.EncodeDb
	chpo      anno.AnnoDb
)

// \n -> <br/>
var isLF = regexp.MustCompile(`\n`)

func init() {
	version.LogVersion()
	flag.Parse()
	if *cpuprofile != "" {
		var f = osUtil.Create(*cpuprofile)
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

	tomlCfg = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

	// CHPO
	chpo.Load(
		tomlCfg.Get("annotation.hpo").(*toml.Tree),
		dbPath,
	)
	// 突变频谱
	spectrumDb.Load(
		tomlCfg.Get("annotation.Gene.spectrum").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	// 基因-疾病
	diseaseDb.Load(
		tomlCfg.Get("annotation.Gene.disease").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
}
func main() {
	var out = osUtil.Create(*output)
	defer simple_util.DeferClose(out)

	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)

	data, titles := simple_util.LongFile2MapArray(*input, "\t", nil)
	additionTitles := textUtil.File2Array(*title)
	titles = append(titles, additionTitles...)

	fmtUtil.Fprintln(out, strings.Join(titles, "\t"))

	for _, item := range data {
		var gene = item["Gene Symbol"]
		var geneIDs []string
		for _, g := range strings.Split(gene, ";") {
			var id, ok = gene2id[g]
			if !ok {
				if !(g == "-" || g == "." || g == "") {
					if *warn {
						log.Printf("can not find gene id of [%s]\n", gene)
					} else {
						log.Fatalf("can not find gene id of [%s]\n", gene)
					}
				}
				id = g
			}
			geneIDs = append(geneIDs, id)
		}

		// CHPO
		chpo.Annos(item, "<br/>", geneIDs)
		// 基因-疾病
		diseaseDb.Annos(item, "<br/>", geneIDs)
		// 突变频谱
		spectrumDb.Annos(item, "<br/>", geneIDs)

		item["Gene"] = item["Omim Gene"]
		item["OMIM"] = item["OMIM_Phenotype_ID"]

		var array []string
		for _, key := range titles {
			array = append(array, isLF.ReplaceAllString(item[key], "<br/>"))
		}
		fmtUtil.FprintStringArray(out, array, "\t")
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
