
# anno2xlsx

[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/liserjrqlxue/anno2xlsx)
[![GoDoc](https://godoc.org/github.com/liserjrqlxue/anno2xlsx?status.svg)](https://pkg.go.dev/github.com/liserjrqlxue/anno2xlsx) 
[![Go Report Card](https://goreportcard.com/badge/github.com/liserjrqlxue/anno2xlsx)](https://goreportcard.com/report/github.com/liserjrqlxue/anno2xlsx)


## AES加密数据库
### 例子
```
PS C:\Users\wangyaoshen\go\src\liser.jrqlxue\anno2xlsx\buildDb> .\buildDb.exe  -input '..\db\全外疾病库2020.Q4-2021.2.5V1.xlsx' -sheet '更新后全外背景库（5755疾病OMIMID，4372个基因）' -key 'entry ID' -rowCount 6566 -keyCount 4372
sheet name:     更新后全外背景库（5755疾病OMIMID，4372个基因）
key column:     entry ID
encode sheet:[更新后全外背景库（5755疾病OMIMID，4372个基因）]
rows:   6566    true
2021/02/05 16:15:42 Skip merge warn of []
keys:   4372    true
write 18086646 byte to ..\db\全外疾病库2020.Q4-2021.2.5V1.xlsx.更新后全外背景库（5755疾病OMIMID，4372个基因）.json.aes
[更新后全外背景库（5755疾病OMIMID，4372个基因）] checked:       true
skip sheet:[基因+表型OMIM号+遗传模式校对情况统计]
skip sheet:[Sheet1]

```

```
PS C:\Users\wangyaoshen\go\src\liser.jrqlxue\anno2xlsx\buildDb> .\buildDb.exe  -input '..\db\V4.3(2021Q1) 基因库-20210205-统一动态突变重复数.xlsx' -sheet '突变谱汇总' -key 'entrez_id' -rowCount 4327 -keyCount 4326 -skipWarn 6,21,22,23,25,27
sheet name:     突变谱汇总
key column:     entrez_id
encode sheet:[突变谱汇总]
rows:   4327    true
2021/02/05 15:18:12 Skip merge warn of [疾病中文名 特殊类型变异 动态突变 突变谱V4 突变谱-V4.1 说明]
keys:   4326    true
write 5955447 byte to ..\db\V4.3(2021Q1) 基因库-20210205-统一动态突变重复数.xlsx.突变谱汇总.json.aes
[突变谱汇总] checked:   true
skip sheet:[删除的原版基因]
skip sheet:[更新备注]

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