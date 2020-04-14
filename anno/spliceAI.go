package anno

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

type SpliceAI struct {
	SpliceAI, SpliceAI_Pred, SpliceAI_Interpretation, ALLELE, SYMBOL string
	DS_AG, DS_AL, DS_DG, DS_DL, DS_Max                               float64
	DP_AG, DP_AL, DP_DG, DP_DL                                       int
}

func (ai *SpliceAI) Parse() (err error) {
	if ai.SpliceAI == "" {
		return errors.New("nil spliceAI")
	}
	var a = strings.Split(ai.SpliceAI, "|")
	ai.ALLELE = a[0]
	ai.SYMBOL = a[1]
	ai.DS_AG, err = strconv.ParseFloat(a[2], 64)
	if err != nil {
		return
	}
	ai.DS_AL, err = strconv.ParseFloat(a[3], 64)
	if err != nil {
		return
	}
	ai.DS_DG, err = strconv.ParseFloat(a[4], 64)
	if err != nil {
		return
	}
	ai.DS_DL, err = strconv.ParseFloat(a[5], 64)
	if err != nil {
		return
	}
	ai.DP_AG, err = strconv.Atoi(a[6])
	if err != nil {
		return
	}
	ai.DP_AL, err = strconv.Atoi(a[7])
	if err != nil {
		return
	}
	ai.DP_DG, err = strconv.Atoi(a[8])
	if err != nil {
		return
	}
	ai.DP_DL, err = strconv.Atoi(a[9])
	if err != nil {
		return
	}
	for _, s := range []float64{ai.DS_AG, ai.DS_AL, ai.DS_DG, ai.DS_DL} {
		ai.DS_Max = math.Max(ai.DS_Max, s)
	}
	if ai.DS_Max >= 0.2 {
		ai.SpliceAI_Pred = "D"
	} else {
		ai.SpliceAI_Pred = "P"
	}
	return
}

func (ai *SpliceAI) Interpreatation(chromesome string, position int) (err error) {
	var interpreation []string
	if ai.DS_AG >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) acceptor gain %f",
				chromesome, position+ai.DP_AG, position, ai.DP_AG, ai.DS_AG,
			),
		)
	}
	if ai.DS_AL >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) acceptor loss %f",
				chromesome, position+ai.DP_AL, position, ai.DP_AL, ai.DS_AL,
			),
		)
	}
	if ai.DS_DG >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) donor gain %f",
				chromesome, position+ai.DP_DG, position, ai.DP_DG, ai.DS_DG,
			),
		)
	}
	if ai.DS_DL >= 0.1 {
		interpreation = append(
			interpreation,
			fmt.Sprintf(
				"%s:%d (=%d%d) donor loss %f",
				chromesome, position+ai.DP_DL, position, ai.DP_DL, ai.DS_DL,
			),
		)
	}
	ai.SpliceAI_Interpretation = strings.Join(interpreation, ";\n")
	return
}

func AnnoSpliceAI(item map[string]string) {
	var ai = SpliceAI{
		SpliceAI: item["SpliceAI"],
	}
	if ai.Parse() != nil {
		return
	}
	var position, err = strconv.Atoi(item["Start"])
	simpleUtil.CheckErr(err)
	position += 1
	simpleUtil.CheckErr(ai.Interpreatation(item["#Chr"], position))
	item["SpliceAI Pred"] = ai.SpliceAI_Pred
	item["SpliceAI Interpretation"] = ai.SpliceAI_Interpretation
}
