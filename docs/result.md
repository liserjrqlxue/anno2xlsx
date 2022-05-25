# `anno2xlsx`输出文件格式

## `quality.json`
输出文件`prefix.sampleID.quality.json`，
该文件包含了先证者样品检测结果的详细质控信息。
该文件为`JSON`格式的`object`，其`value`均为`string`格式,具体格式如下：

| name                  | describe                 | value type | value example                                                                            |
|-----------------------|--------------------------|------------|------------------------------------------------------------------------------------------|
| sampleID              | 样本编号                     | `string`   | 19L0565526-187-188                                                                       |
| rawDataSize           | 原始数据产出（Mb）               | `string`   | 201860.00                                                                                |
| targetRegionSize      | 目标区长度（bp）                | `string`   | 2684595047                                                                               |
| targetRegionCoverage  | 目标区覆盖度                   | `string`   | 99.92%                                                                                   |
| averageDepth          | 标区平均深度（X）                | `string`   | 56.80                                                                                    |
| averageDepthGt4X      | 目标区平均深度>4X位点所占比例         | `string`   | 99.07%                                                                                   |
| averageDepthGt10X     | 目标区平均深度>10X位点所占比例        | `string`   | 98.87%                                                                                   |
| averageDepthGt20X     | 目标区平均深度>20X位点所占比例        | `string`   | 98.33%                                                                                   |
| averageDepthGt30X     | 目标区平均深度>30X位点所占比例        | `string`   | 96.87%                                                                                   |
| bamPath               | bam文件路径                  | `string`   | /jdfsyt1/B2C_RD_P2/PMO/yansaiying/WGS/T10/PCR/project/NA12878/19L0565526-187-188/bam_chr |
| karyotypePrediction   | 核型预测                     | `string`   | 46,XX                                                                                    |
| mtTargetRegionGt2000X | 线粒体基因组区域平均深度>2000X位点所占比例 | `string`   | 99.58%                                                                                   |


### example
`19L0565526-187-188.quality.19L0565526-187-188.json`:
```json
{
  "averageDepth": "56.80",
  "averageDepthGt10X": "98.87%",
  "averageDepthGt20X": "98.33%",
  "averageDepthGt30X": "96.87%",
  "averageDepthGt4X": "99.08%",
  "bamPath": "/jdfsyt1/B2C_RD_P2/PMO/yansaiying/WGS/T10/PCR/project/NA12878/19L0565526-187-188/bam_chr",
  "karyotypePrediction": "46,XX",
  "mtTargetRegionGt2000X": "99.58%",
  "rawDataSize": "201860.00",
  "sampleID": "19L0565526-187-188",
  "targetRegionCoverage": "99.92%",
  "targetRegionSize": "2684595047"
}
```