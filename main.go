package main

import (
	"flag"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/version"
	"github.com/tealeg/xlsx/v3"
)

type xlsxTemplate struct {
	flag      string
	template  string
	xlsx      *xlsx.File
	sheetName string
	sheet     *xlsx.Sheet
	title     []string
	output    string
}

func (xt *xlsxTemplate) save() error {
	return xt.xlsx.Save(xt.output)
}

func init() {
	version.LogVersion()

	flag.Parse()
	checkFlag()

	// log
	logFile, err = os.Create(*logfile)
	simpleUtil.CheckErr(err)
	log.Printf("Log file         : %v\n", *logfile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Log file         : %v\n", *logfile)
	version.LogVersion()

	// 解析配置
	parseToml()
	parseCfg()

	var funcitonLevel = filepath.Join(etcPath, "function.level.json")
	if osUtil.FileExists(funcitonLevel) {
		anno.FuncInfo = jsonUtil.JsonFile2MapInt(funcitonLevel)
	}
}

func main() {
	defer simpleUtil.DeferClose(logFile)

	//  读取数据库
	loadDb()

	// 准备excel输出
	prepareExcel()
	// 填充sheet
	fillSheet()
	// 保存excel
	saveExcel()

	// json
	if *outJson {
		if *qc != "" {
			qc2json(qualitys[0], *prefix+".quality.json")
		}
		if *snv != "" {
			writeBytes(jsonMarshalIndent(tier1Data, "", "  "), *prefix+".tier1.json")
		}
	}
}
