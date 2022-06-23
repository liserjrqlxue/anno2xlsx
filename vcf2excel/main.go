package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/brentp/irelate/interfaces"
	"github.com/brentp/vcfgo"
	//"compress/gzip"
	gzip "github.com/klauspost/pgzip"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

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
)

var (
	csqDescriptionMatch = regexp.MustCompile(`Consequence annotations from Ensembl VEP. Format: `)
)

func main() {
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		log.Fatalln("-in/-out required!")
	}
	var rdr *vcfgo.Reader
	var inF = osUtil.Open(*in)
	defer simpleUtil.DeferClose(inF)

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
	var csq, csqOk = header.Infos["CSQ"]
	var csqKeys []string
	if csqOk {
		var csqDescription = csq.Description
		if csqDescriptionMatch.MatchString(csqDescription) {
			csqKeys = strings.Split(csqDescriptionMatch.ReplaceAllString(csqDescription, ""), "|")
		} else {
			log.Fatalf("CSQ Description not support:[%s]\n", csqDescription)
		}
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
		item["#Chr"] = variant.Chromosome
		var start = int(variant.Start())
		item["Start"] = strconv.Itoa(start)
		item["Stop"] = strconv.Itoa(start + len(variant.Reference))
		item["Ref"] = variant.Reference
		item["Call"] = variant.Alternate[0]

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

		if csqOk {
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
		return strconv.Itoa(value.(int))
	}
	return ""
}
