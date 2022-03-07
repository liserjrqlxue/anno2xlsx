package anno

import (
	"path/filepath"
	"strings"

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

func (db *AnnoDb) Annos(item map[string]string, sep string, keys []string) {
	var tmp = make(map[string][]string)
	for i := range keys {
		var info, ok = db.db[keys[i]]
		for k, v := range db.titleMap {
			if ok {
				tmp[v] = append(tmp[v], info[k])
			} else {
				tmp[v] = append(tmp[v], "")
			}
		}
		for _, v := range db.titleMap {
			item[v] = strings.Join(tmp[v], sep)
		}
	}
}
