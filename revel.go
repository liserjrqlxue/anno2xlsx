package main

import (
	"io"
	"path/filepath"
	"strconv"

	"github.com/brentp/bix"
	"github.com/brentp/irelate/interfaces"
	"github.com/brentp/irelate/parsers"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/pelletier/go-toml"
)

type revelDb struct {
	File       string
	Title      []string
	TitleIndex []int
	Key        []string
	KeyIndex   []int
	tbi        *bix.Bix
}

func (db *revelDb) loadRevel(revelCfg *toml.Tree) {
	simpleUtil.CheckErr(revelCfg.Unmarshal(db))
	if !osUtil.FileExists(db.File) {
		db.File = filepath.Join(dbPath, db.File)
	}
	db.tbi = simpleUtil.HandleError(bix.New(db.File)).(*bix.Bix)
}

func (db *revelDb) anno(item map[string]string) {
	var chr = item["chromosome"]
	var pos = simpleUtil.HandleError(strconv.Atoi(item["Stop"])).(int)
	var rdr = simpleUtil.HandleError(
		db.tbi.Query(
			interfaces.AsIPosition(chr, pos, pos),
		),
	).(interfaces.RelatableIterator)
	defer simpleUtil.DeferClose(rdr)
	for {
		r, err := rdr.Next()
		if err == io.EOF {
			break
		}
		simpleUtil.CheckErr(err)
		var f = r.(*parsers.RefAltInterval).Fields
		var hit = true
		for i := range db.KeyIndex {
			if string(f[db.KeyIndex[i]]) != item[db.Key[i]] {
				hit = false
			}
		}
		if hit {
			for i := range db.TitleIndex {
				item[db.Title[i]] = string(f[db.TitleIndex[i]])
			}
		}
	}
}
