package main

import (
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	"github.com/tealeg/xlsx/v3"
	"github.com/xuri/excelize/v2"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
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

func prepareExcel() {
	prepareTier1()

	// 使用模板处理Tier2.xlsx
	tier2 = prepareTier2(
		sampleList[0],
		*productID,
		*prefix+".Tier2.xlsx",
		filepath.Join(templatePath, "Tier2.xlsx"),
		isEN,
	)

	if outputTier3 {
		prepareTier3()
	}
	logTime("load template")
}

func prepareTier1() {
	tier1Xlsx = xlsx.NewFile()
	// load tier template
	if *snv != "" {
		xlsxUtil.AddSheet(tier1Xlsx, "filter_variants")
		filterVariantsTitle = addFile2Row(
			anno.GuessPath(
				TomlTree.Get("template.tier1.filter_variants").(string),
				etcPath,
			),
			tier1Xlsx.Sheet["filter_variants"].AddRow(),
		)
	}
	if *exon != "" {
		xlsxUtil.AddSheet(tier1Xlsx, "exon_cnv")
		exonCnvTitle = addFile2Row(
			anno.GuessPath(
				TomlTree.Get("template.tier1.exon_cnv").(string),
				etcPath,
			),
			tier1Xlsx.Sheet["exon_cnv"].AddRow(),
		)
	}
	if *large != "" {
		xlsxUtil.AddSheet(tier1Xlsx, "large_cnv")
		largeCnvTitle = addFile2Row(
			anno.GuessPath(
				TomlTree.Get("template.tier1.large_cnv").(string),
				etcPath,
			),
			tier1Xlsx.Sheet["large_cnv"].AddRow(),
		)
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

func prepareTier3() {
	// create Tier3.xlsx
	tier3Xlsx = excelize.NewFile()
	tier3Xlsx.SetSheetName("Sheet1", "总表")
	tier3SW = simpleUtil.HandleError(tier3Xlsx.NewStreamWriter("总表")).(*excelize.StreamWriter)
	tier3Titles = textUtil.File2Array(
		anno.GuessPath(
			TomlTree.Get("template.tier3.title").(string),
			etcPath,
		),
	)
	SteamWriterSetString2Row(tier3SW, 1, tier3RowID, tier3Titles)
	tier3RowID++
}
