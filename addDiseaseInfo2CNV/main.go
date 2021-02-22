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
	"github.com/pelletier/go-toml"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	etcPath = filepath.Join(exPath, "..", "etc")
	dbPath  = filepath.Join(exPath, "..", "db")
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
	geneID = flag.String(
		"geneId",
		filepath.Join(dbPath, "gene.id.txt"),
		"gene symbol and ncbi id list",
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

var tomlCfg *toml.Tree

// database
var (
	aesCode = "c3d112d6a47a0a04aad2b9d2d2cad266"
	gene2id map[string]string
	// 突变频谱
	spectrumDb anno.EncodeDb
	// 基因-疾病
	diseaseDb anno.EncodeDb
)

//var err error

// \n -> <br/>
var isLF = regexp.MustCompile(`\n`)

func main() {
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

	tomlCfg = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

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

	var out = osUtil.Create(*output)
	defer simple_util.DeferClose(out)

	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)

	// parser etc/config.json
	defaultConfig := simple_util.JsonFile2Interface(*config).(map[string]interface{})

	cnvDb, _ := simple_util.LongFile2MapArray(*input, "\t", nil)
	titles := textUtil.File2Array(*title)

	anno.LoadGeneTrans(anno.GetPath("geneSymbol.transcript", dbPath, defaultConfig))

	fmtUtil.Fprintln(out, strings.Join(titles, "\t"))
	for _, item := range cnvDb {
		// Primer
		item["Primer"] = anno.CnvPrimer(item, *cnvType)

		var gene = item["OMIM_Gene"]

		var geneIDs []string
		for _, g := range strings.Split(gene, ";") {
			var id, ok = gene2id[g]
			if !ok {
				if g != "-" && g != "." {
					log.Fatalf("can not find gene id of [%s]\n", gene)
				}
			}
			geneIDs = append(geneIDs, id)
		}

		// 基因-疾病
		diseaseDb.Annos(item, "<br/>", geneIDs)
		// 突变频谱
		spectrumDb.Annos(item, "<br/>", geneIDs)

		item["OMIM"] = item["OMIM_Phenotype_ID"]

		anno.UpdateCnvAnnot(gene, item, gene2id, diseaseDb.Db)

		var array []string
		for _, key := range titles {
			array = append(array, isLF.ReplaceAllString(item[key], "<br/>"))
		}
		fmtUtil.Fprintln(out, strings.Join(array, "\t"))
	}

	if *memprofile != "" {
		var f = osUtil.Create(*memprofile)
		defer simpleUtil.DeferClose(f)
		simpleUtil.CheckErr(pprof.WriteHeapProfile(f))
	}
}
