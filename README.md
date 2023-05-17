# anno2xlsx

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/liserjrqlxue/anno2xlsx)
[![GoDoc](https://godoc.org/github.com/liserjrqlxue/anno2xlsx?status.svg)](https://pkg.go.dev/github.com/liserjrqlxue/anno2xlsx)
[![Go Report Card](https://goreportcard.com/badge/github.com/liserjrqlxue/anno2xlsx)](https://goreportcard.com/report/github.com/liserjrqlxue/anno2xlsx)

## AES加密数据库

### 疾病库/基因频谱

```shell
#!/bin/bash
wget -N https://ftp.ebi.ac.uk/pub/databases/genenames/hgnc/tsv/non_alt_loci_set.txt
stat non_alt_loci_set.txt
buildDb/buildDb \
  -prefix db/全外疾病库 \
  -key 'entry ID' \
  -input db/backup/全外疾病库2023.Q1-2023.05.17.xlsx \
  -sheet '更新后全外背景库（6272疾病OMIMID，4787个基因）' \
  -rowCount 8425 -keyCount 4787

```

or

```shell
sh buildDb/buildDb.sh 'db/backup/全外疾病库2023.Q1-2023.05.17.xlsx' '更新后全外背景库（6272疾病OMIMID，4787个基因）' 'entry ID' 8425 4787 db/全外疾病库
```

![buildDb.png](docs/buildDb.png)

## CHPO

```shell
buildHPO -chpo chpo-2021.json -g2p genes_to_phenotype.txt -output db/gene2chpo.txt
```

# 注意

## 基因-疾病数据库

**基因-疾病数据库**属于保密数据库，仅供内部测试使用  
~~后续会通过数据库服务获取或者直接创建加密数据库~~  
已通过aes对json格式进行加密处理

## 突变频谱数据库
**突变频谱数据库**属于保密数据库，仅供内部测试使用，  
~~后续会通过数据库服务获取或者直接创建加密数据库~~  
已通过aes对json格式进行加密处理

## 特殊位点库
**特殊位点库**根据华大内部流程`bgicg_anno.pl`注释结果中的`MutationName`查找是否特殊位点，
所以以下情形可能发生库失效问题：
1. 输入文件不包含以`MutationName`命名的特定注释结果
2. 输入文件`MutationName`与流程`bgicg_anno.pl`结果不一致
3. 流程`bgicg_anno.pl`有更新，但是数据库未同步更新
4. 位点用与现有配置不同的数据库注释导致的注释结果不一致

# 特性
## 结合性处理
1. 格式转换
```go
func zygosityFormat(zygosity string) string {
	zygosity = strings.Replace(zygosity, "het-ref", "Het", -1)
	zygosity = strings.Replace(zygosity, "het-alt", "Het", -1)
	zygosity = strings.Replace(zygosity, "hom-alt", "Hom", -1)
	zygosity = strings.Replace(zygosity, "hem-alt", "Hemi", -1)
	zygosity = strings.Replace(zygosity, "hemi-alt", "Hemi", -1)
	return zygosity
}
```
2. Het->Hom修正
```go
func homRatio(item map[string]string, threshold float64) {
	var aRatio = strings.Split(item["A.Ratio"], ";")
	var zygositys = strings.Split(item["Zygosity"], ";")
	if len(aRatio) <= len(zygositys) {
		for i := range aRatio {
			var zygosity = zygositys[i]
			if zygosity == "Het" {
				var ratio, err = strconv.ParseFloat(aRatio[i], 64)
				if err != nil {
					ratio = 0
				}
				if ratio >= threshold {
					zygositys[i] = "Hom"
				}
			}
		}
	}
	item["Zygosity"] = strings.Join(zygositys, ";")
}
```
3. Hom->Hemi修正
```go
func hemiPAR(item map[string]string, gender string) {
	var chromosome = item["#Chr"]
	if isChrXY.MatchString(chromosome) && isMale.MatchString(gender) {
		start, e := strconv.Atoi(item["Start"])
		simpleUtil.CheckErr(e, "Start")
		stop, e := strconv.Atoi(item["Stop"])
		simpleUtil.CheckErr(e, "Stop")
		if !inPAR(chromosome, start, stop) && withHom.MatchString(item["Zygosity"]) {
			zygosity := strings.Split(item["Zygosity"], ";")
			genders := strings.Split(gender, ",")
			if len(genders) <= len(zygosity) {
				for i := range genders {
					if isMale.MatchString(genders[i]) && isHom.MatchString(zygosity[i]) {
						zygosity[i] = strings.Replace(zygosity[i], "Hom", "Hemi", 1)
					}
				}
				item["Zygosity"] = strings.Join(zygosity, ";")
			} else {
				log.Fatalf("conflict gender[%s]and Zygosity[%s]\n", gender, item["Zygosity"])
			}
		}
	}
}
```

## 复用拼接字段

因下游数据库结构新增字段开发工作量大，部分额外字段拼接进已有字段内

| 复用字段                   | 拼接字段         | 连接字符串 | 是否可选拼接 |
|------------------------|--------------|-------|--------|
| `ClinVar Significance` | `CLNSIGCONF` | ':'   | 是      |
| `flank`                | `HGVSc`      | ' '   | 是      |

## 注意

- exon cnv输入文件不存在时仅log报错，不中断
