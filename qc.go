package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

func loadFilterStat(filterStat string, quality map[string]string) {
	var db = make(map[string]float64)
	filters := strings.Split(filterStat, ",")
	for _, filter := range filters {
		var fDb = simpleUtil.HandleError(textUtil.File2Map(filter, "\t", false)).(map[string]string)
		var numberOfReads = simpleUtil.HandleError(strconv.ParseFloat(fDb["Number of Reads:"], 32)).(float64)
		var GCfq1 = simpleUtil.HandleError(strconv.ParseFloat(fDb["GC(%) of fq1:"], 32)).(float64)
		var GCfq2 = simpleUtil.HandleError(strconv.ParseFloat(fDb["GC(%) of fq2:"], 32)).(float64)
		var Q20fq1 = simpleUtil.HandleError(strconv.ParseFloat(fDb["Q20(%) of fq1:"], 32)).(float64)
		var Q20fq2 = simpleUtil.HandleError(strconv.ParseFloat(fDb["Q20(%) of fq2:"], 32)).(float64)
		var Q30fq1 = simpleUtil.HandleError(strconv.ParseFloat(fDb["Q30(%) of fq1:"], 32)).(float64)
		var Q30fq2 = simpleUtil.HandleError(strconv.ParseFloat(fDb["Q30(%) of fq2:"], 32)).(float64)
		fDb["Discard Reads related to low qual:"] = strings.TrimSpace(fDb["Discard Reads related to low qual:"])
		var lowQualReads = 0.0
		if fDb["Discard Reads related to low qual:"] != "" {
			lowQualReads = simpleUtil.HandleError(strconv.ParseFloat(strings.TrimSpace(fDb["Discard Reads related to low qual:"]), 32)).(float64)
		}

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

func loadQC(files, kinship string, quality []map[string]string, isWGS bool) {
	var kinshipHash = make(map[string]map[string]string)
	if kinship != "" {
		kinshipHash, _ = textUtil.File2MapMap(kinship, "样品ID", "\t", nil)
	}
	sep := "\t"
	if isWGS {
		sep = ": "
	}
	for i, in := range strings.Split(files, ",") {
		report := textUtil.File2Array(in)
		for _, line := range report {
			if isSharp.MatchString(line) {
				if m := isBamPath.FindStringSubmatch(line); m != nil {
					if osUtil.FileExists(m[1]) {
						quality[i]["bamPath"] = m[1]
					}
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
		kinshipInfo, ok := kinshipHash[quality[i]["样本编号"]]
		if ok {
			for k, v := range kinshipInfo {
				quality[i][k] = v
			}
		}
	}
}

func parseQC() {
	var karyotypeMap = make(map[string]string)
	if *karyotype != "" {
		karyotypeMap, err = textUtil.Files2Map(*karyotype, "\t", true)
		simpleUtil.CheckErr(err)
	}
	// load coverage.report
	if *qc != "" {
		loadQC(*qc, *kinship, qualitys, *wgs)
		for _, quality := range qualitys {
			for k, v := range qualityKeyMap {
				quality[k] = quality[v]
			}
			var ok bool
			quality["核型预测"], ok = karyotypeMap[quality["样本编号"]]
			if !ok {
				quality["核型预测"] = "NA"
			}

			// WGS -> WES
			var rawDataGb, ok1 = quality["原始数据产出（Gb）"]
			var rawDataMb = quality["原始数据产出（Mb）"]
			if ok1 && rawDataMb == "" {
				var rawData, e = strconv.ParseFloat(rawDataGb, 64)
				if e == nil {
					quality["原始数据产出（Mb）"] = fmt.Sprintf("%.2f", rawData*1000)
				}
			}

			if *wesim {
				var qcArray []string
				for _, key := range qcColumn {
					qcArray = append(qcArray, quality[key])
				}
				fmtUtil.FprintStringArray(qcFile, qcArray, "\t")
			}
		}
		if *wesim {
			simpleUtil.CheckErr(qcFile.Close())
		}

		logTime("load coverage.report")
		if *filterStat != "" {
			loadFilterStat(*filterStat, qualitys[0])
		}
		if *imQc != "" {
			parseIMQC(*imQc, qualitys)
		}
		if *mtQc != "" {
			parseMTQC(*mtQc, qualitys)
		}
	}
}

func parseMTQC(files string, qualitys []map[string]string) {
	for i, s := range strings.Split(files, ",") {
		var qc, err = textUtil.File2Map(s, "\t", false)
		simpleUtil.CheckErr(err)
		for k, v := range qc {
			qualitys[i][k] = v
		}
	}
}

func parseIMQC(files string, qualitys []map[string]string) {
	var imqc = make(map[string]map[string]string)
	for _, s := range strings.Split(files, ",") {
		var qc, _ = textUtil.File2MapMap(s, "sampleID", "\t", nil)
		for s, m := range qc {
			imqc[s] = m
		}
	}
	for _, quality := range qualitys {
		var sampleID = quality["样本编号"]
		var qc, ok = imqc[sampleID]
		if ok {
			quality["Q20 碱基的比例"] = qc["Q20_clean"] + "%"
			quality["Q30 碱基的比例"] = qc["Q30_clean"] + "%"
			quality["测序数据的 GC 含量"] = qc["GC_clean"] + "%"
			quality["低质量 reads 比例"] = qc["lowQual"] + "%"
		}
	}
}

func parseList() {
	for _, sample := range sampleList {
		sampleMap[sample] = true
		var quality = make(map[string]string)
		quality["样本编号"] = sample
		qualitys = append(qualitys, quality)
	}
}
