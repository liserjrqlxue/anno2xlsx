package anno

import (
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/pelletier/go-toml"
)

type AnnoDb struct {
	File     string
	MainKey  string
	TitleKey []string
	Title    []string
	titleMap map[string]string
	db       map[string]map[string]string
}

func (db *AnnoDb) Load(cfg *toml.Tree, dbPath string) {
	simpleUtil.CheckErr(cfg.Unmarshal(db))
	if !osUtil.FileExists(db.File) {
		db.File = filepath.Join(dbPath, db.File)
	}
	db.db, _ = textUtil.File2MapMap(db.File, db.MainKey, "\t", nil)
	db.titleMap = make(map[string]string)
	for i := range db.Title {
		db.titleMap[db.TitleKey[i]] = db.Title[i]
	}
}

func (db *AnnoDb) Anno(item map[string]string, key string) {
	var info, ok = db.db[key]
	if ok {
		for k, v := range db.titleMap {
			item[v] = info[k]
		}
	}
}
