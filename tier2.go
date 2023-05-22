package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/tealeg/xlsx/v3"
	"path/filepath"
)

type xlsxTemplate struct {
	template  string
	xlsx      *xlsx.File
	sheetName string
	sheet     *xlsx.Sheet
	title     []string
	output    string
}

func (t *xlsxTemplate) Save() error {
	return t.xlsx.Save(t.output)
}

type templateInfo struct {
	cols      []string
	titles    [2][]string
	noteTitle [2]string
	note      [2][]string
}

func (t *templateInfo) Load(template string) {
	var tier2Infos = simpleUtil.HandleError(
		simpleUtil.HandleError(xlsx.OpenFile(template)).(*xlsx.File).ToSlice(),
	).([][][]string)
	for i, item := range tier2Infos[0] {
		if i > 0 {
			t.cols = append(t.cols, item[0])
			t.titles[0] = append(t.titles[0], item[1])
			t.titles[1] = append(t.titles[1], item[2])
		}
	}
	for _, item := range tier2Infos[1] {
		t.note[0] = append(t.note[0], item[0])
		t.note[1] = append(t.note[1], item[1])
	}
}

func prepareTier2(sampleID, productID, output, template string, en bool) *xlsxTemplate {
	var info = &templateInfo{}
	info.Load(filepath.Join(template))
	var (
		sheetName     = productID + "_" + sampleID
		noteSheetName = "备注说明"
		xt            = &xlsxTemplate{
			output:    output,
			xlsx:      xlsx.NewFile(),
			sheetName: sheetName,
			title:     info.cols,
		}
	)

	if len(sheetName) > 31 {
		xt.sheetName = xt.sheetName[:31]
	}

	xt.sheet = simpleUtil.HandleError(xt.xlsx.AddSheet(xt.sheetName)).(*xlsx.Sheet)
	var row = xt.sheet.AddRow()
	for i := range xt.title {
		if en {
			row.AddCell().SetString(info.titles[1][i])
		} else {
			row.AddCell().SetString(info.titles[0][i])
		}
	}

	var tier2Note = info.note[0]
	if en {
		tier2Note = info.note[1]
		noteSheetName = transEN[noteSheetName]
	}
	var noteSheet = simpleUtil.HandleError(xt.xlsx.AddSheet(noteSheetName)).(*xlsx.Sheet)
	for _, line := range tier2Note {
		noteSheet.AddRow().AddCell().SetString(line)
	}

	return xt
}
