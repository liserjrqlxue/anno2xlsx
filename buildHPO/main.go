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

	var gene2hpo = make(map[string]map[string]string)
	for _, array := range textUtil.File2Slice("genes_to_phenotype.txt", "\t")[1:] {
		var geneID = array[0]
		var hpoID = array[2]
		var item, ok = gene2hpo[geneID]
		if !ok {
			item = make(map[string]string)
		}
		item[hpoID] = getNameCn[hpoID]
		gene2hpo[geneID] = item
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
	for geneID, item := range gene2hpo {
		var hpoIDs, NameCns []string
		for hpoID, nameCn := range item {
			hpoIDs = append(hpoIDs, hpoID)
			if nameCn != "" {
				NameCns = append(NameCns, nameCn)
			}
			simpleUtil.HandleError(
				fmt.Fprintf(
					out,
					"%s\t%s\t%s\n",
					geneID,
					hpoID,
					nameCn,
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
