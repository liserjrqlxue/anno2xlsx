package main

import (
	"log"
	"os"
	"time"

	"github.com/liserjrqlxue/simple-util"
)

func logTierStats(stats map[string]int) {
	log.Printf("Total               Count : %7d\n", stats["Total"])
	if stats["Total"] == 0 {
		return
	}
	if *trio {
		log.Printf("  noProband         Count : %7d\n", stats["noProband"])

		log.Printf("Denovo              Hit   : %7d\n", stats["Denovo"])
		log.Printf("  Denovo B/LB       Hit   : %7d\n", stats["Denovo B/LB"])
		log.Printf("  Denovo Tier1      Hit   : %7d\n", stats["Denovo Tier1"])
		log.Printf("  Denovo Tier2      Hit   : %7d\n", stats["Denovo Tier2"])
	}

	log.Printf("ACMG noB/LB         Hit   : %7d\n", stats["noB/LB"])
	if *trio {
		log.Printf("  +isDenovo         Hit   : %7d\n", stats["isDenovo noB/LB"])
		log.Printf("    +isAF           Hit   : %7d\n", stats["Denovo AF"])
		log.Printf("      +isGene       Hit   : %7d\n", stats["Denovo Gene"])
		log.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["Denovo Function"])
		log.Printf("        +noFunction Hit   : %7d\n", stats["Denovo noFunction"])
		log.Printf("      +noGene       Hit   : %7d\n", stats["Denovo noGene"])
		log.Printf("    +noAF           Hit   : %7d\n", stats["Denovo noAF"])
		log.Printf("  +noDenovo         Hit   : %7d\n", stats["noDenovo noB/LB"])
		log.Printf("    +isAF           Hit   : %7d\n", stats["noDenovo AF"])
		log.Printf("      +isGene       Hit   : %7d\n", stats["noDenovo Gene"])
		log.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["noDenovo Function"])
		log.Printf("        +noFunction Hit   : %7d\n", stats["noDenovo noFunction"])
		log.Printf("      +noGene       Hit   : %7d\n", stats["noDenovo noGene"])
		log.Printf("    +noAF           Hit   : %7d\n", stats["noDenovo noAF"])
	} else {
		log.Printf("    +isAF           Hit   : %7d\n", stats["isAF"])
		log.Printf("      +isGene       Hit   : %7d\n", stats["isGene"])
		log.Printf("        +isFunction Hit   : %7d\tTier1\n", stats["isFunction"])
		log.Printf("        +noFunction Hit   : %7d\n", stats["noFunction"])
		log.Printf("      +noGene       Hit   : %7d\n", stats["noGene"])
		log.Printf("    +noAF           Hit   : %7d\n", stats["noAF"])
	}

	log.Printf("HGMD/ClinVar        Hit   : %7d\n", stats["HGMD/ClinVar"])
	log.Printf("  isAF              Hit   : %7d\n", stats["HGMD/ClinVar isAF"])
	log.Printf("    noMT            Hit   : %7d\tTier1\n", stats["HGMD/ClinVar noMT T1"])
	log.Printf("  noAF              Hit   : %7d\n", stats["HGMD/ClinVar noAF"])
	log.Printf("    noMT            Hit   : %7d\tTier2\n", stats["HGMD/ClinVar noMT T2"])

	log.Printf("SpecVar             Hit   : %7d\n", stats["SpecVar"])

	log.Printf("Retain              Count : %7d\n", stats["Retain"])
	log.Printf("  Tier1             Count : %7d\n", stats["Tier1"])
	log.Printf("    遗传相符        Count : %7d\n", stats["遗传相符"])
	log.Printf("  Tier2             Count : %7d\n", stats["Tier2"])
	log.Printf("  Tier3             Count : %7d\n", stats["Tier3"])
}

func logTime(message string) {
	ts = append(ts, time.Now())
	step++
	var trim = 3*8 - 1
	var str = simple_util.FormatWidth(trim, message, ' ')
	log.Printf("%s\ttook %7.3fs to run.\n", str, ts[step].Sub(ts[step-1]).Seconds())
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
	var hostName, err = os.Hostname()
	log.Printf("Runtime hostname : %s%v\n", hostName, err)
}
