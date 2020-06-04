package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/liserjrqlxue/simple-util"
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
}

func logTime(timeList []time.Time, step1, step2 int, message string) {
	trim := 3*8 - 1
	str := simple_util.FormatWidth(trim, message, ' ')
	log.Printf("%s\ttook %7.3fs to run.\n", str, timeList[step2].Sub(timeList[step1]).Seconds())
}

// version
var (
	buildStamp string
	gitHash    string
	goVersion  string
)

func logVersion() {
	log.Printf("Git Commit Hash  : %s\n", gitHash)
	log.Printf("UTC Build Time   : %s\n", buildStamp)
	log.Printf("Golang Version   : %s\n", goVersion)
	log.Printf("Runtime hostname : %s\n", os.Hostname())
}
