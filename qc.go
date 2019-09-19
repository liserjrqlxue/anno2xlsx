package main

import (
	"fmt"
	"strconv"
)

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
