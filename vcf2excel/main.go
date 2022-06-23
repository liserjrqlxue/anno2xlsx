package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/brentp/irelate/interfaces"
	"github.com/brentp/vcfgo"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/pelletier/go-toml"

	//"compress/gzip"
	gzip "github.com/klauspost/pgzip"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	dbPath  = filepath.Join(exPath, "..", "db")
	etcPath = filepath.Join(exPath, "..", "etc")
)

// flag
var (
	in = flag.String(
		"in",
		"",
		"input vcf",
	)
	out = flag.String(
		"out",
		"",
		"output excel",
	)
	id = flag.String(
		"id",
		"",
		"sampleID, default from VCF",
	)
	cfg = flag.String(
		"cfg",
		filepath.Join(etcPath, "config.toml"),
		"toml config document",
	)
	geneID = flag.String(
		"geneId",
		filepath.Join(dbPath, "gene.id.txt"),
		"gene symbol and ncbi id list",
	)
)

var (
	csqDescriptionMatch = regexp.MustCompile(`Consequence annotations from Ensembl VEP. Format: `)
	crlf                = regexp.MustCompile("\r\n")
	lf                  = regexp.MustCompile("\n")
	tab                 = regexp.MustCompile("\t")
)

// database
var (
	aesCode = "c3d112d6a47a0a04aad2b9d2d2cad266"
	gene2id map[string]string
	chpo    anno.AnnoDb
	// 突变频谱
	spectrumDb anno.EncodeDb
	// 基因-疾病
	diseaseDb anno.EncodeDb
	geneList  = make(map[string]bool)
	// 产前数据库
	prenatalDb anno.EncodeDb
)

func printMapRow(w io.Writer, item map[string]string, keys []string, sep string) {
	var values []string
	for _, key := range keys {
		values = append(values, item[key])
	}
	var line = strings.Join(values, sep)
	line = crlf.ReplaceAllString(line, "<br/>")
	line = lf.ReplaceAllString(line, "<br/>")
	line = tab.ReplaceAllString(line, "&#9;")
	fmtUtil.Fprintln(w, line)
}

func main() {
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		log.Fatalln("-in/-out required!")
	}

	// config
	var TomlTree = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

	var filterVariantsTitle = textUtil.File2Array(
		anno.GuessPath(TomlTree.Get("template.tier1.filter_variants").(string), etcPath),
	)

	// load DB
	chpo.Load(
		TomlTree.Get("annotation.hpo").(*toml.Tree),
		dbPath,
	)

	// 突变频谱
	spectrumDb.Load(
		TomlTree.Get("annotation.Gene.spectrum").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	// 产前数据库
	prenatalDb.Load(
		TomlTree.Get("annotation.Gene.prenatal").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	// 基因-疾病
	diseaseDb.Load(
		TomlTree.Get("annotation.Gene.disease").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	for key := range diseaseDb.Db {
		geneList[key] = true
	}
	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)
	for k, v := range gene2id {
		if geneList[v] {
			geneList[k] = true
		}
	}

	var rdr *vcfgo.Reader
	var inF = osUtil.Open(*in)
	defer simpleUtil.DeferClose(inF)

	var outF = osUtil.Create(*out)
	defer simpleUtil.DeferClose(outF)

	if strings.HasSuffix(*in, ".gz") || strings.HasSuffix(*in, ".bgz") {
		var gr = simpleUtil.HandleError(gzip.NewReader(inF)).(*gzip.Reader)
		defer simpleUtil.DeferClose(gr)

		rdr = simpleUtil.HandleError(vcfgo.NewReader(gr, false)).(*vcfgo.Reader)
	} else {
		rdr = simpleUtil.HandleError(vcfgo.NewReader(inF, false)).(*vcfgo.Reader)
	}

	// header
	var header = rdr.Header
	var hasSample = false
	//simpleUtil.HandleError(vcfgo.NewWriter(os.Stdout, header))
	if len(header.SampleNames) > 0 {
		hasSample = true
	}
	if *id == "" && len(header.SampleNames) > 0 {
		*id = header.SampleNames[0]
	}
	// CSQ
	var csq = header.Infos["CSQ"]
	var csqKeys []string
	var csqDescription = csq.Description
	if csqDescriptionMatch.MatchString(csqDescription) {
		csqKeys = strings.Split(csqDescriptionMatch.ReplaceAllString(csqDescription, ""), "|")
	} else {
		log.Fatalf("CSQ Description not support:[%s]\n", csqDescription)
	}

	// variant
	for {
		var variant = rdr.Read()
		if variant == nil {
			break
		}
		//fmt.Println(variant)
		var item = make(map[string]string)
		item["SampleID"] = *id
		var chromosome = strings.Replace(variant.Chromosome, "chr", "", 1)
		item["#chr"] = "chr" + chromosome
		var start = int(variant.Start())
		item["Start"] = strconv.Itoa(start)
		item["Stop"] = strconv.Itoa(start + len(variant.Reference))
		item["Ref"] = variant.Reference
		item["Call"] = variant.Alternate[0]
		variant.Chrom()

		// INFO
		var info = variant.Info_
		var rs = getInfoInteger(info, "rsID")
		if rs != "" {
			item["rsID"] = "rs" + rs
		}

		item["1000G EAS AF"] = getInfoFloat(info, "1000G_EAS_AF")
		item["1000G AF"] = getInfoFloat(info, "1000G_AF")
		item["ESP6500 AF"] = getInfoFloat(info, "ESP6500_AF")
		item["ExAC EAS AF"] = getInfoFloat(info, "ExAC_EAS_AF")
		item["ExAC AF"] = getInfoFloat(info, "ExAC_AF")
		item["GnomAD EAS AF"] = getInfoFloat(info, "GnomAD_EAS_AF")
		item["GnomAD AF"] = getInfoFloat(info, "GnomAD_AF")
		item["ExAC EAS HomoAlt Count"] = getInfoInteger(info, "ExAC_EAS_Hom")
		item["ExAC HomoAlt Count"] = getInfoInteger(info, "ExAC_Hom")
		item["GnomAD EAS HomoAlt Count"] = getInfoInteger(info, "GnomAD_EAS_Hom")
		item["GnomAD HomoAlt Count"] = getInfoInteger(info, "GnomAD_Hom")

		if hasSample {
			var sample = variant.Samples[0]
			var sumGT = 0
			for _, i := range sample.GT {
				sumGT += i
			}
			var zyosity = "Het"
			if sumGT != 1 {
				zyosity = "Hom"
			}
			item["Zygosity"] = zyosity
			item["Depth"] = strconv.Itoa(sample.DP)
			var altDepth = simpleUtil.HandleError(sample.AltDepths()).([]int)[0]
			item["A.Depth"] = strconv.Itoa(altDepth)
			item["A.Ratio"] = fmt.Sprintf("%.3f", float64(altDepth)/float64(sample.DP))
		}

		var csqItems []map[string]string
		var csqInfos = simpleUtil.HandleError(variant.Info_.Get("CSQ"))
		switch csqInfos.(type) {
		case []string:
			for _, csqInfo := range csqInfos.([]string) {
				var csqArray = strings.Split(csqInfo, "|")
				var csqItem = make(map[string]string)
				for i, key := range csqKeys {
					csqItem[key] = csqArray[i]
				}
				csqItems = append(csqItems, csqItem)
			}
		case string:
			var csqArray = strings.Split(csqInfos.(string), "|")
			var csqItem = make(map[string]string)
			for i, key := range csqKeys {
				csqItem[key] = csqArray[i]
			}
			csqItems = append(csqItems, csqItem)
		default:
			fmt.Printf("I don't know about type %T!\n", csqInfos)
		}
		for _, csqItem := range csqItems {
			for k, v := range item {
				csqItem[k] = v
			}

			// gene
			csqItem["Gene Symbol"] = csqItem["SYMBOL"]
			var gene = csqItem["Gene Symbol"]
			var geneId, ok = gene2id[gene]
			if !ok {
				if gene != "" {
					log.Printf("can not find gene id of [%s]\n", gene)
				}
			}
			item["geneID"] = geneId

			// anno
			chpo.Anno(csqItem, geneId)
			diseaseDb.Anno(csqItem, geneId)
			spectrumDb.Anno(csqItem, geneId)
			prenatalDb.Anno(csqItem, geneId)

			//anno.ParseSpliceAI(csqItem)
			csqItem["cHGVS_org"] = csqItem["HGVSc"]
			csqItem["cHGVS"] = csqItem["HGVSc"]
			acmg2015.AddEvidences(csqItem)
			csqItem["自动化判断"] = acmg2015.PredACMG2015(csqItem, false)

			// updatePos
			csqItem["Stop"] = strconv.Itoa(start + len(variant.Reference))
			switch csqItem["VARIANT_CLASS"] {
			case "SNV":
				csqItem["#Chr+Stop"] = csqItem["#Chr"] + ":" + csqItem["Stop"]
				csqItem["chr-show"] = csqItem["#Chr"] + ":" + csqItem["Stop"]
			case "insertion":
				csqItem["#Chr+Stop"] = csqItem["#Chr"] + ":" + csqItem["Start"] + "-" + csqItem["Stop"]
				csqItem["chr-show"] = fmt.Sprintf("%s:%d..%s", csqItem["#Chr"], variant.Pos, csqItem["Stop"])
			case "substitution", "deletion", "indel":
				csqItem["#Chr+Stop"] = csqItem["#Chr"] + ":" + csqItem["Start"] + "-" + csqItem["Stop"]
				csqItem["chr-show"] = fmt.Sprintf("%s:%s..%d", csqItem["#Chr"], csqItem["Start"], int(variant.Pos)+len(variant.Reference))
			default:
				csqItem["#Chr+Stop"] = csqItem["#Chr"] + ":" + csqItem["Start"] + "-" + csqItem["Stop"]
				csqItem["chr-show"] = fmt.Sprintf("%s:%s..%d", csqItem["#Chr"], csqItem["Start"], int(variant.Pos)+len(variant.Reference))
				log.Printf("unkown VARIANT_CLASS:[%s]\n", csqItem["VARIANT_CLASS"])
			}

			// HGVS
			csqItem["gHGVS"] = strings.Split(csqItem["HGVSg"], ",")[0]
			if csqItem["HGVSc"] != "" {
				csqItem["cHGVS"] = strings.Split(csqItem["HGVSc"], ":")[1]
				item["MutationName"] = fmt.Sprintf("%s(%s):%s", csqItem["Feature"], csqItem["SYMBOL"], csqItem["cHGVS"])
				csqItem["HGVSp"] = strings.ReplaceAll(csqItem["HGVSp"], "%3D", "=")
				if csqItem["HGVSp"] != "" {
					csqItem["pHGVS"] = strings.Split(csqItem["HGVSp"], ":")[1]
					item["MutationName"] += "(" + csqItem["pHGVS"] + ")"
				}
			} else {
				item["MutationName"] = csqItem["gHGVS"]
			}
			printMapRow(outF, csqItem, filterVariantsTitle, "\t")
		}
	}
}

func getInfoInteger(info interfaces.Info, key string) string {
	var value, err = info.Get(key)
	if err == nil && value.(int) > -1 {
		return strconv.Itoa(value.(int))
	}
	return ""
}
func getInfoFloat(info interfaces.Info, key string) string {
	var value, err = info.Get(key)
	if err == nil && value.(float64) > -1 {
		return strconv.FormatFloat(value.(float64), 'g', -1, 64)
	}
	return ""
}
