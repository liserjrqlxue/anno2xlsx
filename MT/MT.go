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

func main() {
	flag.Parse()
	if *db == "" {
		flag.Usage()
		os.Exit(1)
	}
	//var database []map[string]string
	database := simple_util.JsonFile2Interface(*db)
	for _, item := range database.([]interface{}) {
		MTAllele2Variant(item.(map[string]interface{})["Allele"].(string))
	}
}

func MTAllele2Variant(allele string) (chr, ref, alt string, start, end int) {
	var err error
	switch {
	case isSNP.MatchString(allele):
		matchs := isSNP.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 4 {
			chr = "MT"
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
			chr = "MT"
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
			chr = "MT"
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
			chr = "MT"
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
			chr = "MT"
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
			chr = "MT"
			ref = ""
			alt = matchs[4]
			start, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			end, err = strconv.Atoi(matchs[3])
			simple_util.CheckErr(err, matchs...)
			if end-start+1 == len(alt) {
				end = start
				log.Printf("%s %d %d %s %s %v\n", chr, start, end, ref, alt, matchs)
				return
			}
		}
		log.Fatalf("can not parser LINS:[%s]->[%v]\n", allele, matchs)
	default:
		log.Printf("can not parser:[%s]\n", allele)
	}
	return
}
