package main

import (
	"github.com/liserjrqlxue/simple-util"
	"log"
	"path/filepath"
)

func getPath(key string, config map[string]interface{}) (path string) {
	path = getStrVal(key, config)

	if !simple_util.FileExists(path) {
		path = filepath.Join(dbPath, path)
	}
	if !simple_util.FileExists(path) {
		log.Fatalf("can not find %s in config[%v]\n", key, path)
	}
	return
}

func getStrVal(key string, config map[string]interface{}) (val string) {
	val, ok := config[key].(string)
	if !ok {
		log.Fatalf("Error load cfg[%s]:%v\n", key, config[key])
	} else {
		log.Printf("load cfg[%s]:%v\n", key, config[key])
	}
	return
}
