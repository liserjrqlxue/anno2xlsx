package main

import (
	"bytes"
	"encoding/json"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func jsonMarshal(t interface{}) []byte {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	simpleUtil.CheckErr(encoder.Encode(t))
	return buffer.Bytes()
}

func jsonMarshalIndent(t interface{}, prefix, indent string) []byte {
	var buf bytes.Buffer
	simpleUtil.CheckErr(
		json.Indent(&buf, jsonMarshal(t), prefix, indent),
	)
	return buf.Bytes()
}

func select2json(item map[string]string, keys []string) []byte {
	return jsonMarshalIndent(selectMap(item, keys), "", "  ")
}

func writeBytes(b []byte, fileName string) {
	simpleUtil.HandleError(osUtil.Create(fileName).Write(b))
}

func selectMap(item map[string]string, keys []string) map[string]string {
	var selectItem = make(map[string]string)
	for _, k := range keys {
		selectItem[k] = item[k]
	}
	return selectItem
}

func convertMap(item map[string]string, keys map[string]string) map[string]string {
	var selectItem = make(map[string]string)
	for k, v := range keys {
		selectItem[k] = item[v]
	}
	return selectItem
}
