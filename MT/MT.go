package main

import (
	"flag"
	"github.com/liserjrqlxue/simple-util"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// regexp
var (
	isSNP  = regexp.MustCompile(`^([ACGT])(\d+)([ACGT])$`)
	isDUP  = regexp.MustCompile(`^([ACGT])(\d+)([ACGT]+)$`)
	isDEL  = regexp.MustCompile(`^([ACGT])(\d+)del$`)
	isLDEL = regexp.MustCompile(`^(\d+)_(\d+)del([ACGT]+)$`)
	isINS  = regexp.MustCompile(`^([ACGT])(\d+)ins([ACGT])$`)
	isLINS = regexp.MustCompile(`^([ACGT])(\d+)_(\d+)ins([ACGT]+)$`)
)

var (
	db = flag.String(
		"db",
		"",
		"db to be check\n",
	)
)

type Variant struct {
	Chr   string                 `json:"Chromosome"`
	Ref   string                 `json:"Ref"`
	Alt   string                 `json:"Alt"`
	Start int                    `json:"Start"`
	End   int                    `json:"End"`
	Info  map[string]interface{} `json:"Info"`
}

func main() {
	flag.Parse()
	if *db == "" {
		flag.Usage()
		os.Exit(1)
	}

	logFile, err := os.Create(*db + ".db.log")
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(logFile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Log file:%v \n", *db+".db.log")

	//var database []map[string]string
	database := simple_util.JsonFile2Interface(*db)
	var outputDb = make(map[string]Variant)
	for _, item := range database.([]interface{}) {
		var mut = Variant{
			Chr: "MT",
		}
		mut.Info = item.(map[string]interface{})
		allele := mut.Info["Allele"].(string)
		mut.Ref, mut.Alt, mut.Start, mut.End = MTAllele2Variant(allele)
		key := strings.Join(
			[]string{
				mut.Chr,
				strconv.Itoa(mut.Start),
				strconv.Itoa(mut.End),
				mut.Ref,
				mut.Alt,
			},
			"\t",
		)
		if key != "MT\t0\t0\t\t" {
			dup, ok := outputDb[key]
			if ok {
				log.Printf("Duplicate key[%s]:\n\t%+v\n\t%+v\n", key, dup, mut)
			}
			outputDb[key] = mut
		} else {
			log.Printf("Skip allele:[%s]\n", allele)
		}
	}
	simple_util.CheckErr(simple_util.Json2rawFile(*db+".db", outputDb))
}

func MTAllele2Variant(allele string) (ref, alt string, start, end int) {
	var err error
	switch {
	case isSNP.MatchString(allele):
		matchs := isSNP.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 4 {
			ref = matchs[1]
			alt = matchs[3]
			start, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			end = start
			start--
			return
		}
		log.Fatalf("can not parser SNP:[%s]->[%v]\n", allele, matchs)
	case isDUP.MatchString(allele):
		matchs := isDUP.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 4 {
			ref = matchs[1]
			alt = matchs[3]
			start, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			altChr := strings.Split(alt, "")
			end = start
			if altChr[0] == ref {
				ref = ""
				alt = strings.Join(altChr[1:], "")
				return
			} else if altChr[len(altChr)-1] == ref {
				ref = ""
				start--
				end = start
				alt = strings.Join(altChr[:len(altChr)-1], "")
				return
			}
		}
		log.Fatalf("can not parser LSNP:[%s]->[%v]\n", allele, matchs)
	case isDEL.MatchString(allele):
		matchs := isDEL.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 3 {
			ref = matchs[1]
			alt = ""
			start, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			end = start
			start--
			return
		}
		log.Fatalf("can not parser DEL:[%s]->[%v]\n", allele, matchs)
	case isLDEL.MatchString(allele):
		matchs := isLDEL.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 4 {
			alt = ""
			ref = matchs[3]
			start, err = strconv.Atoi(matchs[1])
			simple_util.CheckErr(err, matchs...)
			end, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			if end-start+1 == len(ref) {
				start--
				return
			}
		}
		log.Fatalf("can not parser LDEL:[%s]->[%v]\n", allele, matchs)
	case isINS.MatchString(allele):
		matchs := isINS.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 4 {
			ref = ""
			alt = matchs[3]
			start, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			end = start
			return
		}
		log.Fatalf("can not parser INS:[%s]->[%v]\n", allele, matchs)
	case isLINS.MatchString(allele):
		matchs := isLINS.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 5 {
			ref = ""
			alt = matchs[4]
			start, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			end, err = strconv.Atoi(matchs[3])
			simple_util.CheckErr(err, matchs...)
			if end-start+1 == len(alt) {
				end = start
				return
			}
		}
		log.Fatalf("can not parser LINS:[%s]->[%v]\n", allele, matchs)
	default:
		log.Printf("can not parser:[%s]\n", allele)
	}
	return
}
