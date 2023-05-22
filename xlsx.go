package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/goUtil/xlsxUtil"
	"github.com/tealeg/xlsx/v3"
	"github.com/xuri/excelize/v2"
	"path/filepath"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
)

func prepareExcel() {
	prepareTier1()

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
