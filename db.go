package main

import (
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/stringsUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/pelletier/go-toml"
)

func loadDb() {
	sampleList = strings.Split(*list, ",")

	chpo.Load(
		TomlTree.Get("annotation.hpo").(*toml.Tree),
		dbPath,
	)
	if *academic {
		revel.loadRevel(
			TomlTree.Get("annotation.REVEL").(*toml.Tree),
		)
	}

	// ACMG SF
	if *sf {
		acmgSecondaryFindingDb.Load(
			TomlTree.Get("annotation.Mutation.ACMGSF").(*toml.Tree),
			dbPath,
			[]byte(sfCode),
		)
	}

	// 孕前数据库
	if *pp {
		prePregnancyDb.Load(
			TomlTree.Get("annotation.Mutation.PrePregnancy").(*toml.Tree),
			dbPath,
			[]byte(ppCode),
		)
	}

	// 新生儿数据库
	if *nb {
		newBornDb.Load(
			TomlTree.Get("annotation.Mutation.NBSP").(*toml.Tree),
			dbPath,
			[]byte(nbCode),
		)
	}

	// 耳聋数据库
	if *hl {
		hearingLossDb.Load(
			TomlTree.Get("annotation.Mutation.VIPHL").(*toml.Tree),
			dbPath,
			[]byte(hlCode),
		)
	}

	// 突变频谱
	spectrumDb.Load(
		TomlTree.Get("annotation.Gene.spectrum").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	// 基因-疾病
	diseaseDb.Load(
		TomlTree.Get("annotation.Gene.disease").(*toml.Tree),
		dbPath,
		[]byte(aesCode),
	)
	for key := range diseaseDb.Db {
		geneList[key] = true
	}
	gene2id = simpleUtil.HandleError(textUtil.File2Map(*geneID, "\t", false)).(map[string]string)
	for k, v := range gene2id {
		if geneList[v] {
			geneList[k] = true
		}
	}
	logTime("load Gene-Disease DB")

	// 特殊位点库
	for _, key := range textUtil.File2Array(*specVarList) {
		specVarDb[key] = true
	}
	logTime("load Special mutation DB")

	for transcript, level := range simpleUtil.HandleError(textUtil.File2Map(filepath.Join(etcPath, "转录本优先级.txt"), "\t", false)).(map[string]string) {
		transcriptLevel[transcript] = stringsUtil.Atoi(level)
	}

	var fpMA, _ = textUtil.File2MapArray(filepath.Join(dbPath, "snvindel假阳性统计.xlsx.Sheet1.txt"), "\t", nil)
	for _, m := range fpMA {
		fpDb[m["NM"]+":"+m["c."]] = m
	}
}
