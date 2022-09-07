package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
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

	// json
	if *outJson {
		if *qc != "" {
			var qualityJsonInfo, _ = textUtil.File2MapMap(filepath.Join(etcPath, "quality.json.txt"), "name", "\t", nil)
			var qualityJsonKeyMap = make(map[string]string)
			for k, m := range qualityJsonInfo {
				qualityJsonKeyMap[k] = m["describe"]
			}
			var qualityJson = make(map[string]string)
			for k, v := range qualityJsonKeyMap {
				qualityJson[k] = qualitys[0][v]
			}
			qualityJson["targetRegionSize"] = fmt.Sprintf("%.0f", simpleUtil.HandleError(strconv.ParseFloat(qualityJson["targetRegionSize"], 64)).(float64))
			qualityJson["rawDataSize"] = fmt.Sprintf("%.2f", simpleUtil.HandleError(strconv.ParseFloat(qualityJson["rawDataSize"], 64)).(float64)*1000)
			for _, s := range []string{
				"targetRegionCoverage",
				"averageDepthGt4X",
				"averageDepthGt10X",
				"averageDepthGt20X",
				"averageDepthGt30X",
				"mtTargetRegionGt2000X",
			} {
				if !strings.HasSuffix(qualityJson[s], "%") {
					qualityJson[s] += "%"
				}
			}
			writeBytes(
				jsonMarshalIndent(qualityJson, "", "  "), *prefix+".quality."+qualitys[0]["样本编号"]+".json",
			)
		}
		if *snv != "" {
			writeBytes(jsonMarshalIndent(tier1Data, "", "  "), *prefix+".tier1.json")
		}
	}

	// pprof.WriteHeapProfile
	if *memprofile != "" {
		var f = osUtil.Create(*memprofile)
		defer simpleUtil.DeferClose(f)
		simpleUtil.CheckErr(pprof.WriteHeapProfile(f))
	}
}
