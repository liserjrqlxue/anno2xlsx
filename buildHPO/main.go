package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

type cHPO struct {
	HpoID  string `json:"hpo_id"`
	NameCn string `json:"name_cn"`
}

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
	dbPath = filepath.Join(exPath, "..", "db")
)

//flag
var (
	output = flag.String(
		"output",
		filepath.Join(dbPath, "gene2hpo.txt"),
		"path to output",
	)
	g2p = flag.String(
		"g2p",
		filepath.Join(exPath, "genes_to_phenotype.txt"),
		"path to gene2_to_phenotype.txt",
	)
	chpo = flag.String(
		"chpo",
		filepath.Join(exPath, "chpo-2021.json"),
		"path to chpo.json",
	)
)

func main() {
	flag.Parse()

	var out = osUtil.Create(*output)
	defer simpleUtil.DeferClose(out)

	var chpoArray []cHPO
	var getNameCn = make(map[string]string)
	var gene2hpo = make(map[string]map[string]string)

	simpleUtil.CheckErr(
		json.Unmarshal(
			simpleUtil.HandleError(ioutil.ReadFile(*chpo)).([]byte),
			&chpoArray,
		),
	)
	for _, hpo := range chpoArray {
		getNameCn[hpo.HpoID] = hpo.NameCn
	}

	for _, array := range textUtil.File2Slice(*g2p, "\t")[1:] {
		var geneID = array[0]
		var hpoID = array[2]
		var item, ok = gene2hpo[geneID]
		if !ok {
			item = make(map[string]string)
		}
		item[hpoID] = getNameCn[hpoID]
		gene2hpo[geneID] = item
	}

	simpleUtil.HandleError(
		fmt.Fprintf(
			out,
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
		}
		simpleUtil.HandleError(
			fmt.Fprintf(
				out,
				"%s\t%s\t%s\n",
				geneID,
				strings.Join(hpoIDs, ","),
				strings.Join(NameCns, ","),
			),
		)
	}
}
