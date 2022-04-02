package main

import (
	"encoding/json"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func map2json(item map[string]string) []byte {
	var b, e = json.Marshal(item)
	simpleUtil.CheckErr(e)
	return b
}

func select2json(item map[string]string, keys []string) []byte {
	var selectItem = make(map[string]string)
	for _, k := range keys {
		selectItem[k] = item[k]
	}
	return map2json(selectItem)
}

func writeBytes(b []byte, fileName string) {
	simpleUtil.HandleError(osUtil.Create(fileName).Write(b))
}
