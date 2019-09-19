package main

import (
	"fmt"
	simple_util "github.com/liserjrqlxue/simple-util"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func loadFilterStat(filterStat string, quality map[string]string) {
	if filterStat == "" {
		return
	}
	var db = make(map[string]float64)
	filters := strings.Split(filterStat, ",")
	for _, filter := range filters {
		fDb, err := simple_util.File2Map(filter, "\t", false)
		simple_util.CheckErr(err)
		numberOfReads, err := strconv.ParseFloat(fDb["Number of Reads:"], 32)
		simple_util.CheckErr(err)
		GCfq1, err := strconv.ParseFloat(fDb["GC(%) of fq1:"], 32)
		simple_util.CheckErr(err)
		GCfq2, err := strconv.ParseFloat(fDb["GC(%) of fq2:"], 32)
		simple_util.CheckErr(err)
		Q20fq1, err := strconv.ParseFloat(fDb["Q20(%) of fq1:"], 32)
		simple_util.CheckErr(err)
		Q20fq2, err := strconv.ParseFloat(fDb["Q20(%) of fq2:"], 32)
		simple_util.CheckErr(err)
		Q30fq1, err := strconv.ParseFloat(fDb["Q30(%) of fq1:"], 32)
		simple_util.CheckErr(err)
		Q30fq2, err := strconv.ParseFloat(fDb["Q30(%) of fq2:"], 32)
		simple_util.CheckErr(err)
		lowQualReads, err := strconv.ParseFloat(strings.TrimSpace(fDb["Discard Reads related to low qual:"]), 32)
		simple_util.CheckErr(err)
		db["numberOfReads"] += numberOfReads
		db["lowQualReads"] += lowQualReads
		db["GC"] += (GCfq1 + GCfq2) / 2 * numberOfReads
		db["Q20"] += (Q20fq1 + Q20fq2) / 2 * numberOfReads
		db["Q30"] += (Q30fq1 + Q30fq2) / 2 * numberOfReads
	}
	quality["Q20 碱基的比例"] = strconv.FormatFloat(db["Q20"]/db["numberOfReads"], 'f', 2, 32) + "%"
	quality["Q30 碱基的比例"] = strconv.FormatFloat(db["Q30"]/db["numberOfReads"], 'f', 2, 32) + "%"
	quality["测序数据的 GC 含量"] = strconv.FormatFloat(db["GC"]/db["numberOfReads"], 'f', 2, 32) + "%"
	quality["低质量 reads 比例"] = strconv.FormatFloat(db["lowQualReads"]/db["numberOfReads"], 'f', 2, 32) + "%"
}

func updateQC(stats map[string]int, quality map[string]string) {
	quality["罕见变异占比（Tier1/总）"] = fmt.Sprintf("%0.2f%%", float64(stats["Tier1"])/float64(stats["Total"])*100)
	quality["罕见烈性变异占比 in tier1"] = fmt.Sprintf("%0.2f%%", float64(stats["Tier1LoF"])/float64(stats["Tier1"])*100)
	quality["罕见纯合变异占比 in tier1"] = fmt.Sprintf("%0.2f%%", float64(stats["Tier1Hom"])/float64(stats["Tier1"])*100)
	quality["纯合变异占比 in all"] = fmt.Sprintf("%0.2f%%", float64(stats["Hom"])/float64(stats["Total"])*100)
	for _, chr := range []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13",
		"14", "15", "16", "17", "18", "19", "20", "21", "22", "X",
	} {
		quality["chr"+chr+"纯合变异占比"] = fmt.Sprintf("%0.2f%%", float64(stats["Hom:chr"+chr])/float64(stats["chr"+chr])*100)
	}
	quality["SNVs_all"] = strconv.Itoa(stats["snv"])
	quality["SNVs_tier1"] = strconv.Itoa(stats["Tier1snv"])
	quality["Small insertion（包含 dup）_all"] = strconv.Itoa(stats["ins"])
	quality["Small insertion（包含 dup）_tier1"] = strconv.Itoa(stats["Tier1ins"])
	quality["Small deletion_all"] = strconv.Itoa(stats["del"])
	quality["Small deletion_tier1"] = strconv.Itoa(stats["Tier1del"])
	quality["exon CNV_all"] = strconv.Itoa(stats["exonCNV"])
	quality["exon CNV_tier1"] = strconv.Itoa(stats["Tier1exonCNV"])
	quality["large CNV_all"] = strconv.Itoa(stats["largeCNV"])
	quality["large CNV_tier1"] = strconv.Itoa(stats["Tier1largeCNV"])
}

var isSharp = regexp.MustCompile(`^#`)
var isBamPath = regexp.MustCompile(`^## Files : (\S+)`)

func loadQC(files string, quality []map[string]string, isWGS bool) {
	sep := "\t"
	if isWGS {
		sep = ": "
	}
	file := strings.Split(files, ",")
	for i, in := range file {
		report := simple_util.File2Array(in)
		for _, line := range report {
			if isSharp.MatchString(line) {
				if m := isBamPath.FindStringSubmatch(line); m != nil {
					quality[i]["bamPath"] = m[1]
				}
			} else {
				m := strings.Split(line, sep)
				if len(m) > 1 {
					quality[i][strings.TrimSpace(m[0])] = strings.TrimSpace(m[1])
				}
			}
		}
		if isWGS {
			absPath, err := filepath.Abs(in)
			if err == nil {
				quality[i]["bamPath"] = filepath.Join(filepath.Dir(absPath), "..", "bam_chr")
			} else {
				log.Println(err, in)
				quality[i]["bamPath"] = filepath.Join(filepath.Dir(in), "..", "bam_chr")
			}
		}
	}
}
