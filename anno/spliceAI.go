package anno

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

type spliceAI struct {
	spliceAI, pred, interpretation, allele, symbol string
	dsAG, dsAL, dsDG, dsDL, dsMax                  float64
	dpAG, dpAL, dpDG, dpDL                         int
}

//Parse parse spliceAI
func (ai *spliceAI) Parse() (err error) {
	if ai.spliceAI == "" {
		return errors.New("nil spliceAI")
	}
	var a = strings.Split(ai.spliceAI, "|")
	ai.allele = a[0]
	ai.symbol = a[1]
	ai.dsAG, err = strconv.ParseFloat(a[2], 64)
	if err != nil {
		return
	}
	ai.dsAL, err = strconv.ParseFloat(a[3], 64)
	if err != nil {
		return
	}
	ai.dsDG, err = strconv.ParseFloat(a[4], 64)
	if err != nil {
		return
	}
	ai.dsDL, err = strconv.ParseFloat(a[5], 64)
	if err != nil {
		return
	}
	ai.dpAG, err = strconv.Atoi(a[6])
	if err != nil {
		return
	}
	ai.dpAL, err = strconv.Atoi(a[7])
	if err != nil {
		return
	}
	ai.dpDG, err = strconv.Atoi(a[8])
	if err != nil {
		return
	}
	ai.dpDL, err = strconv.Atoi(a[9])
	if err != nil {
		return
	}
	for _, s := range []float64{ai.dsAG, ai.dsAL, ai.dsDG, ai.dsDL} {
		ai.dsMax = math.Max(ai.dsMax, s)
	}
	if ai.dsMax >= 0.2 {
		ai.pred = "D"
	} else {
		ai.pred = "P"
	}
	return
}

//Interpreatation calculate value of spliceAI.interpretation
func (ai *spliceAI) Interpreatation(chromosome string, position int) (err error) {
	var interpreation []string
	if ai.dsAG >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) acceptor gain %f",
				chromosome, position+ai.dpAG, position, ai.dpAG, ai.dsAG,
			),
		)
	}
	if ai.dsAL >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) acceptor loss %f",
				chromosome, position+ai.dpAL, position, ai.dpAL, ai.dsAL,
			),
		)
	}
	if ai.dsDG >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) donor gain %f",
				chromosome, position+ai.dpDG, position, ai.dpDG, ai.dsDG,
			),
		)
	}
	if ai.dsDL >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) donor loss %f",
				chromosome, position+ai.dpDL, position, ai.dpDL, ai.dsDL,
			),
		)
	}
	ai.interpretation = strings.Join(interpreation, ";\n")
	return
}

// ParseSpliceAI parse and anno spliceAI result
func ParseSpliceAI(item map[string]string) {
	var ai = spliceAI{
		spliceAI: item["SpliceAI"],
	}
	if ai.Parse() != nil {
		return
	}
	var position, err = strconv.Atoi(item["Start"])
	simpleUtil.CheckErr(err)
	position++
	simpleUtil.CheckErr(ai.Interpreatation(item["#Chr"], position))
	item["SpliceAI Pred"] = ai.pred
	item["SpliceAI Interpretation"] = ai.interpretation
}
