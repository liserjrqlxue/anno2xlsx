package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/simple-util"
)

// regexp
var (
	isSNP   = regexp.MustCompile(`^([ACGT])(\d+)([ACGT])$`)
	isDUP   = regexp.MustCompile(`^([ACGT])(\d+)([ACGT]+)$`)
	isDEL   = regexp.MustCompile(`^([ACGT])(\d+)del$`)
	isLDEL  = regexp.MustCompile(`^(\d+)_(\d+)del([ACGT]+)$`)
	isINS   = regexp.MustCompile(`^([ACGT])(\d+)ins([ACGT])$`)
	isLINS  = regexp.MustCompile(`^([ACGT])(\d+)_(\d+)ins([ACGT]+)$`)
	isAfSNP = regexp.MustCompile(`^([ACGT])-([ACGT])$`)
	isAfINS = regexp.MustCompile(`^([ACGT])-([ACGT]+)$`)
)

var (
	db = flag.String(
		"db",
		"",
		"db to be check\n",
	)
	isAF = flag.Bool(
		"af",
		false,
		"if is af db",
	)
)

//Variant struct of var
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
		if *isAF {
			pos := mut.Info["Pos"].(string)
			ref := mut.Info["Ref"].(string)
			alt := mut.Info["Alt"].(string)
			mut.Ref, mut.Alt, mut.Start, mut.End = MTPosRefAlt2Variant(pos, ref, alt)
		} else {
			allele := mut.Info["Allele"].(string)
			mut.Ref, mut.Alt, mut.Start, mut.End = MTAllele2Variant(allele)
		}

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
			if *isAF {
				var info = make(map[string]interface{})
				for _, key := range []string{"# in HG branch with variant", "Total # HG branch seqs"} {
					info[key], err = strconv.Atoi(mut.Info[key].(string))
					simple_util.CheckErr(err)
				}
				if ok {
					for _, key := range []string{"# in HG branch with variant", "Total # HG branch seqs"} {
						count, err := strconv.Atoi(mut.Info[key].(string))
						simple_util.CheckErr(err)
						dup.Info[key] = dup.Info[key].(int) + count
					}
					dup.Info["Fequency in HG branch(%)"] =
						float64(dup.Info["# in HG branch with variant"].(int)) / float64(dup.Info["Total # HG branch seqs"].(int)) * 100
				} else {
					info["Fequency in HG branch(%)"] =
						float64(info["# in HG branch with variant"].(int)) / float64(info["Total # HG branch seqs"].(int)) * 100
					mut.Info = info
					outputDb[key] = mut
				}
			} else {
				if ok {
					jsonByte1, err1 := json.MarshalIndent(dup, "", "\t")
					jsonByte2, err2 := json.MarshalIndent(mut, "", "\t")
					log.Printf(
						"Duplicate key[%s]:\n\t%s,%v\n\t%s,%v\n",
						key, jsonByte1, err1, jsonByte2, err2,
					)
				} else {
					outputDb[key] = mut
				}
			}
		} else {
			jsonByte, err := json.MarshalIndent(item, "", "\t")
			log.Printf("Skip item:%s,%v\n", jsonByte, err)
		}
	}
	simple_util.CheckErr(simple_util.Json2rawFile(*db+".db", outputDb))
}

//MTAllele2Variant convert MT allele to var info
func MTAllele2Variant(allele string) (ref, alt string, start, end int) {
	var err error
	switch {
	case isSNP.MatchString(allele):
		matchs := isSNP.FindStringSubmatch(allele)
		if matchs != nil && len(matchs) == 4 {
			ref = matchs[1]
			alt = matchs[3]
			end, err = strconv.Atoi(matchs[2])
			simple_util.CheckErr(err, matchs...)
			start = end - 1
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

//MTPosRefAlt2Variant convert MT pos ref alt to var info
func MTPosRefAlt2Variant(Pos, Ref, Alt string) (ref, alt string, start, end int) {
	var err error
	if Alt == ":" {
		ref = Ref
		alt = ""
		start, err = strconv.Atoi(Pos)
		simple_util.CheckErr(err)
		start--
		end = start + len(Ref)
		return
	}
	if len(Alt) == 1 {
		ref = Ref
		alt = Alt
		start, err = strconv.Atoi(Pos)
		simple_util.CheckErr(err)
		start--
		end = start + len(Ref)
		return
	}
	altChr := strings.Split(Alt, "")
	if altChr[0] == Ref {
		ref = ""
		alt = strings.Join(altChr[1:], "")
		start, err = strconv.Atoi(Pos)
		simple_util.CheckErr(err)
		end = start + len(Ref)
		return
	}
	log.Fatalf("can not parser:%s:%s>%s\n", Pos, Ref, Alt)
	return
}

//MTPosNC2Variant convert MT pos nc to var info
func MTPosNC2Variant(pos, nc string) (ref, alt string, start, end int) {
	var err error
	switch {
	case isAfSNP.MatchString(nc):
		matchs := isAfSNP.FindStringSubmatch(nc)
		if matchs != nil && len(matchs) == 3 {
			ref = matchs[1]
			alt = matchs[2]
			end, err = strconv.Atoi(pos)
			simple_util.CheckErr(err, pos, nc)
			start = end - 1
			return
		}
		log.Fatalf("can not parser SNP:%s\t[%s]->[%v]\n", pos, nc, matchs)
	case isAfINS.MatchString(nc):
		matchs := isAfINS.FindStringSubmatch(nc)
		if matchs != nil && len(matchs) == 3 {
			ref = matchs[1]
			alt = matchs[2]
			altChr := strings.Split(alt, "")
			if altChr[0] == ref {
				ref = ""
				alt = strings.Join(altChr[1:], "")
				start, err = strconv.Atoi(pos)
				simple_util.CheckErr(err, pos, nc)
				end = start + len(alt)
				return
			}
		}
		log.Fatalf("can not parser SNP:%s\t[%s]->[%v]\n", pos, nc, matchs)
	default:
		log.Printf("can not parser:%s\t[%s]\n", pos, nc)
	}
	return
}
