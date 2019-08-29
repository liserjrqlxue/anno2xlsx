package anno

import (
	"github.com/liserjrqlxue/simple-util"
	"log"
	"path/filepath"
)

func GetPath(key, dbPath string, config map[string]interface{}) (path string) {
	path = GetStrVal(key, config)

	if !simple_util.FileExists(path) {
		path = filepath.Join(dbPath, path)
	}
	if !simple_util.FileExists(path) {
		log.Fatalf("can not find %s in config[%v]\n", key, path)
	}
	return
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
