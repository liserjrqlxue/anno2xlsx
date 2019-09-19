package main

import (
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"log"
	"time"
)

func logTierStats(stats map[string]int) {
	fmt.Printf("Total               Count : %7d\n", stats["Total"])
	if stats["Total"] == 0 {
		return
	}
	if *trio {
		fmt.Printf("  noProband         Count : %7d\n", stats["noProband"])

		fmt.Printf("Denovo              Hit   : %7d\n", stats["Denovo"])
		fmt.Printf("  Denovo B/LB       Hit   : %7d\n", stats["Denovo B/LB"])
		fmt.Printf("  Denovo Tier1      Hit   : %7d\n", stats["Denovo Tier1"])
		fmt.Printf("  Denovo Tier2      Hit   : %7d\n", stats["Denovo Tier2"])
	}

	fmt.Printf("ACMG noB/LB         Hit   : %7d\n", stats["noB/LB"])
	if *trio {
		fmt.Printf("  +isDenovo         Hit   : %7d\n", stats["isDenovo noB/LB"])
		fmt.Printf("    +isAF           Hit   : %7d\n", stats["Denovo AF"])
		fmt.Printf("      +isGene       Hit   : %7d\n", stats["Denovo Gene"])
		fmt.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["Denovo Function"])
		fmt.Printf("        +noFunction Hit   : %7d\n", stats["Denovo noFunction"])
		fmt.Printf("      +noGene       Hit   : %7d\n", stats["Denovo noGene"])
		fmt.Printf("    +noAF           Hit   : %7d\n", stats["Denovo noAF"])
		fmt.Printf("  +noDenovo         Hit   : %7d\n", stats["noDenovo noB/LB"])
		fmt.Printf("    +isAF           Hit   : %7d\n", stats["noDenovo AF"])
		fmt.Printf("      +isGene       Hit   : %7d\n", stats["noDenovo Gene"])
		fmt.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["noDenovo Function"])
		fmt.Printf("        +noFunction Hit   : %7d\n", stats["noDenovo noFunction"])
		fmt.Printf("      +noGene       Hit   : %7d\n", stats["noDenovo noGene"])
		fmt.Printf("    +noAF           Hit   : %7d\n", stats["noDenovo noAF"])
	} else {
		fmt.Printf("    +isAF           Hit   : %7d\n", stats["isAF"])
		fmt.Printf("      +isGene       Hit   : %7d\n", stats["isGene"])
		fmt.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["isFunction"])
		fmt.Printf("        +noFunction Hit   : %7d\n", stats["noFunction"])
		fmt.Printf("      +noGene       Hit   : %7d\n", stats["noGene"])
		fmt.Printf("    +noAF           Hit   : %7d\n", stats["noAF"])
	}

	fmt.Printf("HGMD/ClinVar        Hit   : %7d\n", stats["HGMD/ClinVar"])
	fmt.Printf("  isAF              Hit   : %7d\n", stats["HGMD/ClinVar isAF"])
	fmt.Printf("    noMT            Hit   : %7d\tTier1\n", stats["HGMD/ClinVar noMT T1"])
	fmt.Printf("  noAF              Hit   : %7d\n", stats["HGMD/ClinVar noAF"])
	fmt.Printf("    noMT            Hit   : %7d\tTier2\n", stats["HGMD/ClinVar noMT T2"])

	fmt.Printf("SpecVar             Hit   : %7d\n", stats["SpecVar"])

	fmt.Printf("Retain              Count : %7d\n", stats["Retain"])
	fmt.Printf("  Tier1             Count : %7d\n", stats["Tier1"])
	fmt.Printf("    遗传相符        Count : %7d\n", stats["遗传相符"])
	fmt.Printf("  Tier2             Count : %7d\n", stats["Tier2"])
	fmt.Printf("  Tier3             Count : %7d\n", stats["Tier3"])

	fmt.Printf("罕见变异占比（Tier1/总） : %0.2f%%\n", float64(stats["Tier1"])/float64(stats["Total"])*100)
	fmt.Printf("罕见烈性变异占比in tier1 : %0.2f%%\n", float64(stats["Tier1LoF"])/float64(stats["Tier1"])*100)
	fmt.Printf("罕见纯合变异占比in tier1 : %0.2f%%\n", float64(stats["Tier1Hom"])/float64(stats["Tier1"])*100)
	fmt.Printf("纯合变异占比in all   : %0.2f%%\n", float64(stats["Hom"])/float64(stats["Total"])*100)

	for _, chr := range []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13",
		"14", "15", "16", "17", "18", "19", "20", "21", "22", "X",
	} {
		fmt.Printf("chr%-2s纯合变异占比in tier1: %0.2f%%\n", chr, float64(stats["Hom:chr"+chr])/float64(stats["chr"+chr])*100)
	}
	fmt.Printf("SNVs : %7d\n", stats["snv"])
	fmt.Printf("SNVs : %7d\n", stats["Tier1snv"])
	fmt.Printf("Small insertion（包含 dup）: %7d\n", stats["ins"])
	fmt.Printf("Small insertion（包含 dup）: %7d\n", stats["Tier1ins"])
	fmt.Printf("Small deletion: %7d\n", stats["del"])
	fmt.Printf("Small deletion: %7d\n", stats["Tier1del"])
	fmt.Printf("exon CNV: %7d\n", stats["exonCNV"])
	fmt.Printf("exon CNV: %7d\n", stats["Tier1exonCNV"])
	fmt.Printf("large CNV: %7d\n", stats["largeCNV"])
	fmt.Printf("large CNV: %7d\n", stats["Tier1largeCNV"])
}

func logTime(timeList []time.Time, step1, step2 int, message string) {
	trim := 3*8 - 1
	str := simple_util.FormatWidth(trim, message, ' ')
	log.Printf("%s\ttook %7.3fs to run.\n", str, timeList[step2].Sub(timeList[step1]).Seconds())
}

func logVersion() {
	if gitHash != "" || buildStamp != "" || goVersion != "" {
		log.Printf("Git Commit Hash: %s\n", gitHash)
		log.Printf("UTC Build Time : %s\n", buildStamp)
		log.Printf("Golang Version : %s\n", goVersion)
	}
}
