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

func prepareExcel() {
	prepareTier1()
	prepareTier2()
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

func prepareTier2() {
	// 准备英文产品列表
	var productEn = textUtil.File2Array(filepath.Join(etcPath, "product.en.list"))
	for i := range productEn {
		isEnProduct[productEn[i]] = true
	}
	// tier2
	tier2 = xlsxTemplate{
		flag:      "Tier2",
		sheetName: *productID + "_" + sampleList[0],
	}
	tier2.output = *prefix + "." + tier2.flag + ".xlsx"
	tier2.xlsx = xlsx.NewFile()

	var tier2Infos = simpleUtil.HandleError(
		simpleUtil.HandleError(xlsx.OpenFile(filepath.Join(templatePath, "Tier2.xlsx"))).(*xlsx.File).ToSlice(),
	).([][][]string)
	for i, item := range tier2Infos[0] {
		if i > 0 {
			tier2TemplateInfo.cols = append(tier2TemplateInfo.cols, item[0])
			tier2TemplateInfo.titles[0] = append(tier2TemplateInfo.titles[0], item[1])
			tier2TemplateInfo.titles[1] = append(tier2TemplateInfo.titles[1], item[2])
		}
	}
	for _, item := range tier2Infos[1] {
		tier2TemplateInfo.note[0] = append(tier2TemplateInfo.note[0], item[0])
		tier2TemplateInfo.note[1] = append(tier2TemplateInfo.note[1], item[1])
	}

	tier2.sheet, err = tier2.xlsx.AddSheet(tier2.sheetName)
	simpleUtil.CheckErr(err)
	tier2row := tier2.sheet.AddRow()
	for i, col := range tier2TemplateInfo.cols {
		tier2.title = append(tier2.title, col)
		var title string
		if isEnProduct[*productID] {
			title = tier2TemplateInfo.titles[0][i]
		} else {
			title = tier2TemplateInfo.titles[1][i]
		}
		tier2row.AddCell().SetString(title)
	}

	var tier2NoteSheetName = "备注说明"
	var tier2Note []string
	if isEnProduct[*productID] {
		tier2NoteSheetName = transEN[tier2NoteSheetName]
		tier2Note = tier2TemplateInfo.note[1]
	} else {
		tier2Note = tier2TemplateInfo.note[0]
	}
	var tier2NoteSheet = simpleUtil.HandleError(tier2.xlsx.AddSheet(tier2NoteSheetName)).(*xlsx.Sheet)
	for _, line := range tier2Note {
		tier2NoteSheet.AddRow().AddCell().SetString(line)
	}
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
