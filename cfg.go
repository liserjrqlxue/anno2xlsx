package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/pelletier/go-toml"
)

func parseCfg() {
	// parser etc/config.json
	defaultConfig = jsonUtil.JsonFile2Interface(*config).(map[string]interface{})

	initAcmg2015()

	if *specVarList == "" {
		*specVarList = anno.GetPath("specVarList", dbPath, defaultConfig)
	}
	if *transInfo == "" {
		*transInfo = anno.GetPath("transInfo", dbPath, defaultConfig)
	}
	if *wgs {
		for _, key := range defaultConfig["qualityColumnWGS"].([]interface{}) {
			qualityColumn = append(qualityColumn, key.(string))
		}
	} else {
		for _, key := range defaultConfig["qualityColumn"].([]interface{}) {
			qualityColumn = append(qualityColumn, key.(string))
		}
	}

	initIM()

	if *wgs {
		MTTitle = textUtil.File2Array(filepath.Join(etcPath, "MT.title.txt"))
		qualityKeyMap = simpleUtil.HandleError(
			textUtil.File2Map(filepath.Join(etcPath, "wgs.qc.txt"), "\t", false),
		).(map[string]string)
	} else {
		qualityKeyMap = simpleUtil.HandleError(
			textUtil.File2Map(filepath.Join(etcPath, "coverage.report.txt"), "\t", false),
		).(map[string]string)
	}

	parseList()
	parseQC()
}

func parseToml() {
	TomlTree = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

	acmgDb = filepath.Join(etcPath, TomlTree.Get("acmg.list").(string))
	openRedis()
}

func openRedis() {
	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = TomlTree.Get("redis.addr").(string)
		}
		redisDb = redis.NewClient(&redis.Options{
			Addr: *redisAddr,
		})
		log.Printf("Connect [%s]:%s\n", redisDb.String(), simpleUtil.HandleError(redisDb.Ping().Result()).(string))
	}
}

func initIM() {
	if *wesim {
		acmg59GeneList := textUtil.File2Array(anno.GuessPath(TomlTree.Get("acmg.59gene").(string), etcPath))
		for _, gene := range acmg59GeneList {
			acmg59Gene[gene] = true
		}

		resultColumn = TomlTree.Get("wesim.resultColumn").([]string)
		if *trio {
			resultColumn = append(resultColumn, "Genotype of Family Member 1", "Genotype of Family Member 2")
		}
		resultFile = osUtil.Create(*prefix + ".result.tsv")
		fmtUtil.Fprintln(resultFile, strings.Join(resultColumn, "\t"))

		qcColumn = TomlTree.Get("wesim.qcColumn").([]string)
		qcFile = osUtil.Create(*prefix + ".qc.tsv")
		fmtUtil.Fprintln(qcFile, strings.Join(qcColumn, "\t"))
	}
}
