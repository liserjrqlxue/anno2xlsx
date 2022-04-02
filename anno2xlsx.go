package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"

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

}

func main() {
	// pprof.StartCPUProfile
	if *cpuprofile != "" {
		var f = osUtil.Create(*cpuprofile)
		simpleUtil.CheckErr(pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
	}
	defer simpleUtil.DeferClose(logFile)

	//  读取数据库
	loadDb()

	// 准备excel输出
	prepareExcel()
	// 填充sheet
	fillSheet()
	// 保存excel
	saveExcel()

	if *qc != "" {
		writeBytes(select2json(qualitys[0], qualityColumn), *prefix+".quality."+qualitys[0]["样本编号"]+".json")
	}

	// pprof.WriteHeapProfile
	if *memprofile != "" {
		var f = osUtil.Create(*memprofile)
		defer simpleUtil.DeferClose(f)
		simpleUtil.CheckErr(pprof.WriteHeapProfile(f))
	}
}
