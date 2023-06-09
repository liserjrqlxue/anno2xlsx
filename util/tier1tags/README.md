### 功能
1. 过滤Tier1
2. 添加"遗传相符"和"筛选标签"

### 注意
1. 过滤Tier1规则的前序步骤不在本程序内
2. 输出tsv格式

### 示例
`tier1tags -snv input.anno.tsv,intron.anno.filter.tsv`  
output:`input.anno.tsv.tier1.tsv`,`input.anno.tsv.tier1.xlsx`

`tier1tags -snv input.anno.tsv,intron.anno.filter.tsv -prefix test`  
output:`test.tier1.tsv`,`test.tier1.xlsx`

### 预期用途
输入tier1数据和过滤后的intron数据
过滤Tier1并添加"遗传相符"和"筛选标签"