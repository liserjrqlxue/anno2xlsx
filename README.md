# anno2xlsx

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/liserjrqlxue/anno2xlsx)
[![GoDoc](https://godoc.org/github.com/liserjrqlxue/anno2xlsx?status.svg)](https://pkg.go.dev/github.com/liserjrqlxue/anno2xlsx)
[![Go Report Card](https://goreportcard.com/badge/github.com/liserjrqlxue/anno2xlsx)](https://goreportcard.com/report/github.com/liserjrqlxue/anno2xlsx)

## USAGE

### PARAM

| arg          | type    | example                                         | note                                                                                 |
|--------------|---------|-------------------------------------------------|--------------------------------------------------------------------------------------|
| -academic    | boolean |                                                 | 学术使用，比如REVEL                                                                         |
| -acmg        | boolean |                                                 | 使用ACMG2015计算证据项PVS1, PS1,PS4, PM1,PM2,PM4,PM5 PP2,PP3, BA1, BS1,BS2, BP1,BP3,BP4,BP7 |
| -autoPVS1    | boolean |                                                 | 使用autoPVS1结果处理证据项PVS1                                                                |
| -allTier1    | boolean |                                                 | 不进行tier1过滤                                                                           |
| -allgene     | boolean |                                                 | tier1过滤不过滤基因                                                                         |
| -cfg         | string  | etc/config.toml                                 | toml配置文件                                                                             |
| -cnvAnnot    | boolean |                                                 | 重新进行UpdateCnvAnnot                                                                   |
| -cnvFlter    | boolean |                                                 | 进行 cnv 结果过滤                                                                          |
| -config      | string  | etc/config.json                                 | json配置文件                                                                             |
| -couple      | boolean |                                                 | 夫妻模式                                                                                 |
| -exon        | string  | sample1.exon.txt,sample2.exon.txt               | exon CNV 输入文件，逗号分割，过滤 -list 内样品列表                                                    |
| -extra       | string  | extra1.txt,extra2.txt                           | 额外简单放入excel的额外sheets中                                                                |
| -extraSheet  | string  | sheet1,sheet2                                   | -extra对应sheet name                                                                   |
| -qc          | string  | sample1.coverage.report,sample2.coverage.report | bamdst质控文件，逗号分割，与 -list 顺序一致                                                         |
| -filterStat  | string  | L01.filter.stat,L02.filter.stat                 | 计算reads QC的文件，逗号分割                                                                   |
| -imqc        | string  | sample1.QC.txt,sample2.QC.txt                   | 一体机QC.txt格式QC输入，逗号分割，过滤 -list 内样品列表                                                  |
| -mtqc        | string  | sample1.MT.QC.txt,sample2.MT.QC.txt             | 线粒体QC.txt，逗号分割，过滤 -list 内样品列表                                                        |
| -gender      | string  | M,M,F                                           | 样品性别，逗号分割，与 -list 顺序一致                                                               |
| -geneId      | string  | db/gene.id.txt                                  | 基因名-基因ID 对应数据库                                                                       |
| -hl          | boolean |                                                 | 使用耳聋变异库                                                                              |
| -json        | boolean |                                                 | 输出json格式结果                                                                           |
| -karyotype   | string  | sample1.karyotpye.txt,sample2.karyotype.txt     | 核型信息，逗号分割                                                                            |
| -kinship     | string  | kinship.txt                                     | trio的亲缘关系                                                                            |
| -large       | string  | sample1.large.txt,sample2.large.txt             | large CNV注释结果，逗号分割                                                                   |
| -list        | string  | sample1,sample2,sample3                         | 样品编号，逗号分割，**有顺序**                                                                    |
| -log         | string  | prefix.log                                      | log输出文件                                                                              |
| -loh         | string  | loh1.xlsx,loh2.xlsx                             | loh结果excel，逗号分割，按 -list 样品编号顺序创建sheet                                                |
| -lohSheet    | string  | LOH_annotation                                  | sheet name后缀                                                                         |
| -nb          | boolean |                                                 | 使用新筛变异库                                                                              |
| -pp          | boolean |                                                 | 使用孕前变异库                                                                              |
| -prefix      | string  | outputPrefix                                    | 输出前缀，默认 -snv 第一个输入                                                                   |
| -product     | string  | DX1516                                          | 产品编号                                                                                 |
| -redis       | boolean |                                                 | 使用redis服务注释本地频率                                                                      |
| -redisAddr   | string  | 127.0.0.1:6380                                  | redis服务器地址                                                                           |
| -save        | boolean |                                                 | 保持excel                                                                              |
| -seqType     | string  | SEQ2000                                         | redis 查询关键词，区分频率库                                                                    |
| -snv         | string  | snv1.txt,snv2.txt                               | snv注释结果，逗号分割                                                                         |
| -specVarList | string  | etc/spec.var.lite.txt                           | 特殊变异库                                                                                |
| -tag         | string  | .tag                                            | tier1结果文件名加入额外标签，[prefix].Tier1[tag].xlsx                                            |
| -transInfo   | string  | db/trans.exonCount.json.new.json                | json格式转录本exon count数据库                                                               |
| -trio        | boolean |                                                 | 标准trio模式                                                                             |
| -trio2       | boolean |                                                 | 非标准trio，但是保持先证者、父亲、母亲顺序                                                              |
| -warn        | boolean |                                                 | 警告基因名无法识别问题，而非中断                                                                     |
| -wgs         | boolean |                                                 | wgs模式                                                                                |

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

## 注意

### 基因-疾病数据库

**基因-疾病数据库**属于保密数据库，仅供内部测试使用  
~~后续会通过数据库服务获取或者直接创建加密数据库~~  
已通过aes对json格式进行加密处理

### 突变频谱数据库

**突变频谱数据库**属于保密数据库，仅供内部测试使用，  
~~后续会通过数据库服务获取或者直接创建加密数据库~~  
已通过aes对json格式进行加密处理

### 特殊位点库

**特殊位点库**根据华大内部流程`bgicg_anno.pl`注释结果中的`MutationName`查找是否特殊位点，
所以以下情形可能发生库失效问题：

1. 输入文件不包含以`MutationName`命名的特定注释结果
2. 输入文件`MutationName`与流程`bgicg_anno.pl`结果不一致
3. 流程`bgicg_anno.pl`有更新，但是数据库未同步更新
4. 位点用与现有配置不同的数据库注释导致的注释结果不一致

## 特性

### 结合性处理

#### 格式转换

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

#### Het->Hom修正

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

#### Hom->Hemi修正

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

## UTIL

### `tier1tags`

WGS 使用anno2xlsx过滤后，进行spliceAI注释和过滤，然后重新进行Tier1判断、"遗传相符"和"筛选标签"
参考 [README.md](util/tier1tags/README.md)

#### TO-DO

- [ ] Tier1判断去冗余
