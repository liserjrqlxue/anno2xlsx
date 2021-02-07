package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/pelletier/go-toml"
)

type chpoDb struct {
	File     string
	MainKey  string
	TitleKey []string
	Title    []string
	titleMap map[string]string
	db       map[string]map[string]string
}

func (db *chpoDb) loadCHPO(hpoCfg *toml.Tree) {
	simpleUtil.CheckErr(hpoCfg.Unmarshal(db))
	db.db, _ = textUtil.File2MapMap(db.File, db.MainKey, "\t", nil)
	for i := range db.Title {
		db.titleMap[db.TitleKey[i]] = db.Title[i]
	}
}

func (db *chpoDb) anno(item map[string]string) {
	var key = item["geneID"]
	var info, ok = db.db[key]
	if ok {
		for k, v := range db.titleMap {
			item[v] = info[k]
		}
	}
}
