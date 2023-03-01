package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/go-redis/redis"
	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/jsonUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/pelletier/go-toml"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
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
}

func parseToml() {
	TomlTree = simpleUtil.HandleError(toml.LoadFile(*cfg)).(*toml.Tree)

	acmgDb = filepath.Join(etcPath, TomlTree.Get("acmg.list").(string))
	openRedis()

	var tier3 = TomlTree.Get("tier3")
	if tier3 != nil {
		outputTier3 = tier3.(bool)
	}
	var homRatio = TomlTree.Get("homFixRatioThreshold")
	if homRatio != nil {
		homFixRatioThreshold = homRatio.(float64)
	}

	// update tier1 AF threshold
	var tier1AFThreshold = TomlTree.Get("tier1.Tier1AFThreshold")
	if tier1AFThreshold != nil {
		anno.Tier1AFThreshold = tier1AFThreshold.(float64)
	}
	var tier1PLPAFThreshold = TomlTree.Get("tier1.Tier1PLPAFThreshold")
	if tier1PLPAFThreshold != nil {
		anno.Tier1PLPAFThreshold = tier1PLPAFThreshold.(float64)
	}
	var tier1InHouseAFThreshold = TomlTree.Get("tier1.Tier1InHouseAFThreshold")
	if tier1InHouseAFThreshold != nil {
		anno.Tier1InHouseAFThreshold = tier1InHouseAFThreshold.(float64)
	}
}

func openRedis() {
	if *ifRedis {
		if *redisAddr == "" {
			*redisAddr = TomlTree.Get("redis.addr").(string)
		}
		redisDb = redis.NewClient(
			&redis.Options{
				Addr:     *redisAddr,
				Password: TomlTree.Get("redis.pass").(string),
			},
		)
		log.Printf("Connect [%s]:%s\n", redisDb.String(), simpleUtil.HandleError(redisDb.Ping().Result()).(string))
	}
}

func initIM() {
	if *wesim {
		acmg59GeneList := textUtil.File2Array(anno.GuessPath(TomlTree.Get("acmg.59gene").(string), etcPath))
		for _, gene := range acmg59GeneList {
			acmg59Gene[gene] = true
		}

		resultColumn = TomlTree.GetArray("wesim.resultColumn").([]string)
		if *trio {
			resultColumn = append(resultColumn, "Genotype of Family Member 1", "Genotype of Family Member 2")
		}
		resultFile = osUtil.Create(*prefix + ".result.tsv")
		fmtUtil.Fprintln(resultFile, strings.Join(resultColumn, "\t"))

		qcColumn = TomlTree.GetArray("wesim.qcColumn").([]string)
		qcFile = osUtil.Create(*prefix + ".qc.tsv")
		fmtUtil.Fprintln(qcFile, strings.Join(qcColumn, "\t"))

		cnvColumn = TomlTree.GetArray("wesim.cnvColumn").([]string)

		exonFile = osUtil.Create(*prefix + ".exonCNV.tsv")
		fmtUtil.Fprintln(exonFile, strings.Join(cnvColumn, "\t"))

		largeFile = osUtil.Create(*prefix + ".largeCNV.tsv")
		fmtUtil.Fprintln(largeFile, strings.Join(cnvColumn, "\t"))
	}
}
