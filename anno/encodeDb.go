package anno

import (
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	simple_util "github.com/liserjrqlxue/simple-util"
	"github.com/pelletier/go-toml"
)

type EncodeDb struct {
	File     string
	MainKey  string
	TitleKey []string
	Title    []string
	titleMap map[string]string
	codeKey  []byte
	db       map[string]map[string]string
}

func (db *EncodeDb) Load(cfg *toml.Tree, dbPath string, codeKey []byte) {
	simpleUtil.CheckErr(cfg.Unmarshal(db))
	if !osUtil.FileExists(db.File) {
		db.File = filepath.Join(dbPath, db.File)
	}
	db.codeKey = codeKey
	db.db = jsonUtil.Json2MapMap(simple_util.File2Decode(db.File, db.codeKey))
	db.titleMap = make(map[string]string)
	for i := range db.Title {
		db.titleMap[db.TitleKey[i]] = db.Title[i]
	}
}

func (db *EncodeDb) Anno(item map[string]string, key string) {
	var info, ok = db.db[key]
	if ok {
		for k, v := range db.titleMap {
			item[v] = info[k]
		}
	}
}

func (db *EncodeDb) Annos(item map[string]string, sep string, keys []string) {
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
	}
	for _, v := range db.titleMap {
		item[v] = strings.Join(tmp[v], sep)
	}
}
