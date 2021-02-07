package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

type chpo struct {
	HpoID  string `json:"hpo_id"`
	NameCn string `json:"name_cn"`
}

func main() {
	var chpoArray []chpo
	simpleUtil.CheckErr(
		json.Unmarshal(
			simpleUtil.HandleError(ioutil.ReadFile("chpo-2021.json")).([]byte),
			&chpoArray,
		),
	)
	var getNameCn = make(map[string]string)
	for _, hpo := range chpoArray {
		getNameCn[hpo.HpoID] = hpo.NameCn
	}

	var gene2hpo = make(map[string][]chpo)
	for _, array := range textUtil.File2Slice("genes_to_phenotype.txt", "\t")[1:] {
		var geneID = array[0]
		var hpoID = array[2]
		gene2hpo[geneID] = append(
			gene2hpo[geneID],
			chpo{
				HpoID:  hpoID,
				NameCn: getNameCn[hpoID],
			},
		)
	}

	var out = osUtil.Create("gene2chpo.txt")
	defer simpleUtil.DeferClose(out)
	simpleUtil.HandleError(
		fmt.Fprintf(
			out,
			"%s\t%s\t%s\n",
			"entre-gene-id",
			"HPO-Term-ID",
			"HPO-Term-NameCN",
		),
	)
	var out2 = osUtil.Create("gene2chpos.txt")
	defer simpleUtil.DeferClose(out2)
	simpleUtil.HandleError(
		fmt.Fprintf(
			out2,
			"%s\t%s\t%s\n",
			"entre-gene-id",
			"HPO-Term-ID",
			"HPO-Term-NameCN",
		),
	)
	for geneID, hpoArray := range gene2hpo {
		var hpoIDs, NameCns []string
		for _, hpo := range hpoArray {
			hpoIDs = append(hpoIDs, hpo.HpoID)
			NameCns = append(NameCns, hpo.NameCn)
			simpleUtil.HandleError(
				fmt.Fprintf(
					out,
					"%s\t%s\t%s\n",
					geneID,
					hpo.HpoID,
					hpo.NameCn,
				),
			)
		}
		simpleUtil.HandleError(
			fmt.Fprintf(
				out2,
				"%s\t%s\t%s\n",
				geneID,
				strings.Join(hpoIDs, ","),
				strings.Join(NameCns, ","),
			),
		)
	}
}
