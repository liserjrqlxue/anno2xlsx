package main

import (
	"flag"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/version"
)

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

	var productEn = textUtil.File2Array(filepath.Join(etcPath, "product.en.list"))
	for i := range productEn {
		isEnProduct[productEn[i]] = true
	}
	isEN = isEnProduct[*productID]
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
