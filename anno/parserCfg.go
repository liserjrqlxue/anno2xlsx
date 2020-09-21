package anno

import (
	"log"
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/osUtil"
)

func GetPath(key, dbPath string, config map[string]interface{}) (path string) {
	path = GuessPath(GetStrVal(key, config), dbPath)
	return
}

//GuessPath guess path as abs path or relative path in dbPath
func GuessPath(path, dbPath string) string {
	if !osUtil.FileExists(path) {
		path = filepath.Join(dbPath, path)
	}
	if !osUtil.FileExists(path) {
		log.Fatalf("can not find %s\n", path)
	}
	return path
}

func GetStrVal(key string, config map[string]interface{}) (val string) {
	val, ok := config[key].(string)
	if !ok {
		log.Fatalf("Error load cfg[%s]:%v\n", key, config[key])
	} else {
		log.Printf("load cfg[%s]:%v\n", key, config[key])
	}
	return
}
