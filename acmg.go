package main

import (
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

func initAcmg2015() {
	if *acmg {
		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(acmgDb, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = anno.GuessPath(v, dbPath)
		}
		acmg2015.Init(acmgCfg)
	}
}
